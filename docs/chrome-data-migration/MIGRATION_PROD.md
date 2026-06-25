# Widget Dashboard Templates Migration Plan (Production)

Migrate the `dashboard_templates` table from chrome-service DB to widget-layout-backend DB in production.

**IMPORTANT: Complete and verify the [staging migration](./MIGRATION_STAGING.md) before proceeding with production.**

---

## Table of Contents

1. [Overview](#1-overview)
2. [Prerequisites](#2-prerequisites)
3. [Schema Translation](#3-schema-translation)
4. [Step 1: Obtain Production Breakglass Access (app-interface MR)](#4-step-1-obtain-production-breakglass-access-app-interface-mr)
5. [Step 2: Wait for MR Merge and Reconciliation](#5-step-2-wait-for-mr-merge-and-reconciliation)
6. [Step 3: Log In to the Cluster](#6-step-3-log-in-to-the-cluster)
7. [Step 4: Verify Namespace Access](#7-step-4-verify-namespace-access)
8. [Step 5: Launch Debug-Container in Chrome-Service Namespace (Source)](#8-step-5-launch-debug-container-in-chrome-service-namespace-source)
9. [Step 6: Export Data from Chrome-Service DB](#9-step-6-export-data-from-chrome-service-db)
10. [Step 7: Transfer Files to Local Machine](#10-step-7-transfer-files-to-local-machine)
11. [Step 8: Launch Debug-Container in Widget-Layout Namespace (Target)](#11-step-8-launch-debug-container-in-widget-layout-namespace-target)
12. [Step 9: Transfer Files to Target Debug-Container](#12-step-9-transfer-files-to-target-debug-container)
13. [Step 10: Run Preflight Check on Target DB](#13-step-10-run-preflight-check-on-target-db)
14. [Step 11: Import Data into Widget-Layout DB](#14-step-11-import-data-into-widget-layout-db)
15. [Step 12: Verify the Migration](#15-step-12-verify-the-migration)
16. [Step 13: Clean Up](#16-step-13-clean-up)
17. [Rollback Procedure](#17-rollback-procedure)
18. [Reference](#18-reference)

---

## 1. Overview

| Item | Detail |
|------|--------|
| **Source DB** | chrome-service PostgreSQL (prod) |
| **Target DB** | widget-layout-backend PostgreSQL (prod) |
| **Source table** | `dashboard_templates` (chrome-service) |
| **Target table** | `dashboard_templates` (widget-layout-backend) |
| **Cluster** | `crcp01ue1` — both namespaces are on the same cluster |
| **Source namespace** | `chrome-service-prod` |
| **Target namespace** | `widget-layout-backend-prod` |
| **Source DB secret** | `chrome-service-db` |
| **Target DB secret** | `widget-layout-backend-db` |
| **Method** | Python script using `psycopg2` inside debug-containers |
| **Migration script** | `widget-migration-script.py` |
| **Access method** | Breakglass role (required for production) |
| **Authorized users** | tefaz, khala, bflorkie, mmarosi |

### Key Differences from Staging

| Aspect | Staging | Production |
|--------|---------|------------|
| **Access type** | `edit` (staging dev role) | Breakglass role (scoped RBAC Role) |
| **Cluster** | `crcs02ue1` | `crcp01ue1` |
| **Namespaces** | `chrome-service-stage`, `widget-layout-backend-stage` | `chrome-service-prod`, `widget-layout-backend-prod` |
| **Read replica requirement** | Not required | Required for InProgress/OnBoarded apps (check with AppSRE) |
| **Approval** | Self-serviceable via `platform-experience` role | Requires AppSRE review for breakglass |
| **Risk** | Low — staging data only | High — real user data |

---

## 2. Prerequisites

- Staging migration completed and verified successfully
- `oc` CLI installed locally
- Access to the `crcp01ue1` OpenShift cluster console
- The breakglass MR merged (see Step 1)
- The migration script `widget-migration-script.py` available locally (same script used in staging)
- Staging migration results reviewed and confirmed correct
- Read the "Lessons Learned" section in `MIGRATION_STAGING.md` before proceeding

---

## 3. Schema Translation

Same as staging. The migration script handles these translations automatically:

| Chrome-Service Column | Widget-Layout Column | Translation |
|-----------------------|----------------------|-------------|
| `id` (uint, PK) | `id` (uint, PK) | Auto-generated in target (not copied) |
| `user_identity_id` (uint, FK) | `user_id` (string) | JOIN `user_identities` table to resolve `account_id` |
| _(does not exist)_ | `dashboard_name` (string) | Copied from `display_name` |
| `default` (bool) | `default` (bool) | Direct copy |
| `name` (string, embedded) | `name` (string, embedded) | Translated via NAME_MAP (e.g., `landingPage` → `landing-landingPage`) |
| `display_name` (string, embedded) | `display_name` (string, embedded) | Direct copy |
| `sm` (JSON) | `sm` (JSON) | Widget item `"i"` fields transformed via WIDGET_ID_MAP (e.g., `rhel#rhel` → `landing-./RhelWidget`) |
| `md` (JSON) | `md` (JSON) | Widget item `"i"` fields transformed via WIDGET_ID_MAP |
| `lg` (JSON) | `lg` (JSON) | Widget item `"i"` fields transformed via WIDGET_ID_MAP |
| `xl` (JSON) | `xl` (JSON) | Widget item `"i"` fields transformed via WIDGET_ID_MAP |
| `created_at` (timestamp) | `created_at` (timestamp) | Direct copy |
| `updated_at` (timestamp) | `updated_at` (timestamp) | Direct copy |
| `deleted_at` (timestamp) | `deleted_at` (timestamp) | Direct copy |

> **Note on `x`/`y` vs `cx`/`cy` coordinates:** The widget grid items stored in the `sm`/`md`/`lg`/`xl` JSONB columns use `x`/`y` keys. The widget-layout-backend API also uses `x`/`y` at runtime. The `cx`/`cy` naming is only required in YAML configuration files (ConfigMaps, env vars) because YAML parsers treat bare `y` as a boolean. Since this migration copies JSONB directly between PostgreSQL databases — never passing through a YAML parser — no coordinate key renaming is needed.
>
> **Note on widget item `"i"` field transformation:** Chrome-service stores widget item identifiers in `"shortKey#shortKey"` format (e.g., `"rhel#rhel"`), while widget-layout-backend uses `"{scope}-{module}"` format (e.g., `"landing-./RhelWidget"`). The migration script transforms these via `WIDGET_ID_MAP` during export. The full mapping is documented in `docs/WIDGET_MIGRATION.md`.

---

## 4. Step 1: Obtain Production Breakglass Access (app-interface MR)

Production requires a **breakglass role** — `edit` is prohibited on production namespaces per Self-SRE OpenShift Access Standards. Breakglass uses a scoped Kubernetes RBAC Role bound via a group-based RoleBinding, following the [breakglass access documentation](https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/docs/platform-users/access/breakglass-access.md).

The breakglass pattern consists of five components:

1. **Role Definition** — defines the breakglass role with expiration and cluster group
2. **RBAC Role** — Kubernetes Role with least-privilege permissions (one per namespace)
3. **RBAC RoleBinding** — binds the breakglass group to the Role (one per namespace)
4. **Cluster Group Registration** — registers the group in the cluster's `managedGroups`
5. **Namespace Resource References** — adds RBAC resources to each namespace's `openshiftResources`

### 4.1 Create the Breakglass RBAC Role Resources

File: `resources/insights-prod/chrome-service-prod/widget-migration-breakglass.role.yaml`

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: widget-migration-breakglass
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "create", "delete"]
- apiGroups: [""]
  resources: ["pods/exec"]
  verbs: ["create", "get"]
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["chrome-service-db"]
  verbs: ["get"]
```

File: `resources/insights-prod/widget-layout-backend-prod/widget-migration-breakglass.role.yaml`

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: widget-migration-breakglass
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "create", "delete"]
- apiGroups: [""]
  resources: ["pods/exec"]
  verbs: ["create", "get"]
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["widget-layout-backend-db"]
  verbs: ["get"]
```

### 4.2 Create the RBAC RoleBindings

File: `resources/insights-prod/chrome-service-prod/widget-migration-breakglass.rolebinding.yaml`

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: widget-migration-breakglass
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: widget-migration-breakglass
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: Group
  name: widget-migration-breakglass-group
```

File: `resources/insights-prod/widget-layout-backend-prod/widget-migration-breakglass.rolebinding.yaml`

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: widget-migration-breakglass
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: widget-migration-breakglass
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: Group
  name: widget-migration-breakglass-group
```

### 4.3 Create the Breakglass Role Definition

File: `data/teams/insights/roles/widget-migration-breakglass.yml`

```yaml
---
# Temporary role for Widget Migration production breakglass access
#
# IMPORTANT: This role has an expirationDate.
# The expiration date serves as a reminder to review whether this breakglass access is still needed.

$schema: /access/role-1.yml

labels: {}
name: widget-migration-breakglass
description: |
  Breakglass role for Widget Migration production debugging. Provides minimal access to
  create debug pods and read secrets.

permissions: []

# Set an expiration date to minimal required window
expirationDate: '2026-07-31'

access:
- cluster:
    $ref: /openshift/crcp01ue1/cluster.yml
  group: widget-migration-breakglass-group

self_service:
- change_type:
    $ref: /app-interface/changetype/breakglass-role-manager.yml
  datafiles:
  - $ref: /teams/insights/roles/widget-migration-breakglass.yml
```

### 4.4 Register the Group in Cluster Configuration

Edit `data/openshift/crcp01ue1/cluster.yml` and add `widget-migration-breakglass-group` to the `managedGroups` list:

```yaml
managedGroups:
- dedicated-admins
- dedicated-readers
# ... existing groups ...
- widget-migration-breakglass-group
```

### 4.5 Reference Resources in Namespace Definitions

Edit `data/services/insights/chrome-service/namespaces/chrome-service-prod.yml` — add to `openshiftResources`:

```yaml
- provider: resource
  path: /insights-prod/chrome-service-prod/widget-migration-breakglass.role.yaml
- provider: resource
  path: /insights-prod/chrome-service-prod/widget-migration-breakglass.rolebinding.yaml
```

Edit `data/services/insights/widget-layout-backend/namespaces/crcp01ue1-widget-layout-backend-prod.yml` — add to `openshiftResources`:

```yaml
- provider: resource
  path: /insights-prod/widget-layout-backend-prod/widget-migration-breakglass.role.yaml
- provider: resource
  path: /insights-prod/widget-layout-backend-prod/widget-migration-breakglass.rolebinding.yaml
```

**IMPORTANT: `managedResourceTypes` is required.** If the namespace file does not already include a `managedResourceTypes` field, the `openshift-resources` integration will skip Role and RoleBinding types entirely — the resources will appear in `openshiftResources` but never get deployed to the cluster. Add the following to both namespace files:

```yaml
managedResourceTypes:
- Role.rbac.authorization.k8s.io
- RoleBinding.rbac.authorization.k8s.io
```

See [troubleshooting docs](https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/docs/platform-users/FAQ/troubleshooting.md#my-configuration-is-merged-into-app-interface-but-it-isnt-applied) for details.

### 4.6 Add Role to User Files

Add the following line to the `roles:` section of each user file in `data/teams/insights/users/`:

```yaml
- $ref: /teams/insights/roles/widget-migration-breakglass.yml
```

Users to update:
- `tefaz.yml`
- `khala.yml`
- `bflorkie.yml`
- `mmarosi.yml`

### 4.7 Add Resource Paths to Self-Service

Add the RBAC resource files to the `resource-owner` self-service section in `data/teams/insights/roles/platform-experience-services.yml`:

```yaml
- change_type:
    $ref: /app-interface/changetype/resource-owner.yml
  resources:
  - /insights-prod/chrome-service-prod/widget-migration-breakglass.role.yaml
  - /insights-prod/chrome-service-prod/widget-migration-breakglass.rolebinding.yaml
  - /insights-prod/widget-layout-backend-prod/widget-migration-breakglass.role.yaml
  - /insights-prod/widget-layout-backend-prod/widget-migration-breakglass.rolebinding.yaml
```

### 4.8 Commit, Push, and Create MR

```bash
cd /path/to/app-interface
git fetch upstream master
git checkout -b widget-migration-prod-breakglass upstream/master
# ... make the changes above ...
git add <all new and modified files>
git commit -m "Add breakglass role for prod widget migration"
git push origin widget-migration-prod-breakglass
```

Create MR targeting `service/app-interface` master. This MR will require AppSRE review since it involves production access.

---

## 5. Step 2: Wait for MR Merge and Reconciliation

1. Get MR reviewed and approved (requires AppSRE approval for production breakglass)
2. Once merged, wait for `qontract-reconcile` to apply the RBAC changes (5-15 minutes)
3. Verify:
   ```bash
   oc auth can-i create pods -n chrome-service-prod
   ```
   Should return `yes` once reconciliation is complete.

---

## 6. Step 3: Log In to the Cluster

1. Navigate to the `crcp01ue1` cluster console:
   ```
   https://console-openshift-console.apps.crcp01ue1.o9m8.p1.openshiftapps.com
   ```

2. Click your username (top right) -> **Copy login command** -> **Display Token**

3. Copy the `oc login` command and run it in your terminal:
   ```bash
   oc login --token=sha256~<your-token> --server=https://api.crcp01ue1.o9m8.p1.openshiftapps.com:6443
   ```

4. Verify login:
   ```bash
   oc whoami
   ```

---

## 7. Step 4: Verify Namespace Access

```bash
# Check chrome-service-prod
oc auth can-i create pods -n chrome-service-prod
# Expected: yes

oc auth can-i create pods/exec -n chrome-service-prod
# Expected: yes

oc auth can-i get secrets --field-selector=metadata.name=chrome-service-db -n chrome-service-prod
# Expected: yes

# Check widget-layout-backend-prod
oc auth can-i create pods -n widget-layout-backend-prod
# Expected: yes

oc auth can-i create pods/exec -n widget-layout-backend-prod
# Expected: yes

oc auth can-i get secrets --field-selector=metadata.name=widget-layout-backend-db -n widget-layout-backend-prod
# Expected: yes
```

If any return `no`, the RBAC hasn't reconciled yet. Wait a few more minutes and retry.

---

## 8. Step 5: Launch Debug-Container in Chrome-Service Namespace (Source)

```bash
oc process --local \
    -f https://raw.githubusercontent.com/app-sre/container-images/master/debug-container/openshift.yml \
    -p POSTGRES_DB_SECRET_NAME="chrome-service-db" \
| oc -n chrome-service-prod apply -f -
```

Wait for the deployment to roll out and get the pod name:

```bash
oc -n chrome-service-prod rollout status deployment/debug-container --timeout=120s

# Get the actual pod name (it has a hash suffix)
SOURCE_POD=$(oc -n chrome-service-prod get pods -l app=debug-container -o jsonpath='{.items[0].metadata.name}')
echo "Source pod: $SOURCE_POD"
```

---

## 9. Step 6: Export Data from Chrome-Service DB

### 9.1 Copy the Migration Script into the Debug-Container

```bash
oc -n chrome-service-prod cp ./widget-migration-script.py "$SOURCE_POD:/tmp/widget-migration-script.py"
```

### 9.2 Exec into the Debug-Container

```bash
oc -n chrome-service-prod exec -it "$SOURCE_POD" -- bash
```

### 9.3 Run Preflight Check

```bash
python3 /tmp/widget-migration-script.py preflight-export
```

This checks:
- DB connectivity and SELECT permissions
- Active row count in `dashboard_templates`
- Orphaned rows (no matching `user_identity`) — these will be dropped by the JOIN
- Users with NULL/empty `account_id` — these would produce NULL `user_id` in target
- Total exportable rows after JOIN

**IMPORTANT:** All checks must show `[PASS]`. Document the row counts. **STOP if any check shows `[FAIL]`** — investigate before proceeding. Production data is critical.

If the DB secret is not mounted correctly:
```bash
env | grep -E 'PG|DB_'
```

> **Note:** The debug-container mounts the secret as `PG*` env vars (`PGHOST`, `PGUSER`, etc.), not `DB_*`. The migration script accepts both formats automatically.

### 9.4 Verify NAME_MAP Completeness

The migration script translates template names via `NAME_MAP`. Before exporting, verify that all distinct names in production are covered:

```sql
SELECT DISTINCT name FROM dashboard_templates WHERE deleted_at IS NULL;
```

Compare the output against `NAME_MAP` in `widget-migration-script.py`. If any names are missing, add them to the map before proceeding — the export will abort if it encounters an unmapped name.

### 9.5 Run the Export

```bash
python3 /tmp/widget-migration-script.py export
```

Expected output:
```
Found N active dashboard_templates rows
Fetched N rows (with user_id resolved)
Transformed widget IDs in JSONB columns (sm/md/lg/xl)
Generated N INSERT statements in /tmp/widget_migration.sql
```

### 9.6 Verify Export Data

```bash
# Check row count matches
grep -c "^INSERT" /tmp/widget_migration.sql

# Inspect first few statements
head -30 /tmp/widget_migration.sql

# Check for NULL user_ids (failed JOINs)
grep "VALUES (''" /tmp/widget_migration.sql | wc -l
# Should be 0
```

**STOP if any NULL user_ids are found.** This means some `user_identity_id` values don't have matching `user_identities` rows. Investigate before proceeding.

### 9.7 Exit the Container

```bash
exit
```

---

## 10. Step 7: Transfer Files to Local Machine

> **WARNING:** `oc cp` and `oc exec cat > file` can silently truncate large files (~9.5MB+). Use the base64 method below to ensure complete transfer. See "Lessons Learned" in `MIGRATION_STAGING.md` for details.

Transfer the SQL and metadata files from the source debug-container:

```bash
# Transfer SQL file via base64 (reliable for large files)
oc -n chrome-service-prod exec "$SOURCE_POD" -- base64 /tmp/widget_migration.sql > ./widget_migration_prod.sql.b64
base64 -D -i ./widget_migration_prod.sql.b64 -o ./widget_migration_prod.sql
rm ./widget_migration_prod.sql.b64

# Transfer metadata file (small, plain copy is fine)
oc -n chrome-service-prod exec "$SOURCE_POD" -- cat /tmp/widget_migration.meta > ./widget_migration_prod.meta
```

> **Note:** On Linux, use `base64 -d` instead of `base64 -D`.

Verify the file was transferred completely:

```bash
grep -c "^INSERT" ./widget_migration_prod.sql
cat ./widget_migration_prod.meta
# INSERT count must match the count in the .meta file
```

**Recommended:** Compare prod row count with staging to ensure they are in the expected ratio. If prod has significantly fewer or more rows than expected, investigate.

---

## 11. Step 8: Launch Debug-Container in Widget-Layout Namespace (Target)

```bash
oc process --local \
    -f https://raw.githubusercontent.com/app-sre/container-images/master/debug-container/openshift.yml \
    -p POSTGRES_DB_SECRET_NAME="widget-layout-backend-db" \
| oc -n widget-layout-backend-prod apply -f -
```

Wait for the deployment to roll out and get the pod name:

```bash
oc -n widget-layout-backend-prod rollout status deployment/debug-container --timeout=120s

# Get the actual pod name
TARGET_POD=$(oc -n widget-layout-backend-prod get pods -l app=debug-container -o jsonpath='{.items[0].metadata.name}')
echo "Target pod: $TARGET_POD"
```

---

## 12. Step 9: Transfer Files to Target Debug-Container

```bash
# Transfer SQL file via base64 (reliable for large files)
base64 -i ./widget_migration_prod.sql | oc -n widget-layout-backend-prod exec -i "$TARGET_POD" -- bash -c 'base64 -d > /tmp/widget_migration.sql'

# Transfer metadata file
cat ./widget_migration_prod.meta | oc -n widget-layout-backend-prod exec -i "$TARGET_POD" -- bash -c 'cat > /tmp/widget_migration.meta'

# Transfer migration script
cat ./widget-migration-script.py | oc -n widget-layout-backend-prod exec -i "$TARGET_POD" -- bash -c 'cat > /tmp/widget-migration-script.py'
```

> **Note:** On Linux, use `base64` (no `-i` flag) instead of `base64 -i`.

Verify INSERT count on the target pod:

```bash
oc -n widget-layout-backend-prod exec "$TARGET_POD" -- grep -c "^INSERT" /tmp/widget_migration.sql
# Must match the count from the export step
```

---

## 13. Step 10: Run Preflight Check on Target DB

**This check is mandatory for production.** Do not skip.

```bash
oc -n widget-layout-backend-prod exec -it "$TARGET_POD" -- bash
```

```bash
python3 /tmp/widget-migration-script.py preflight-import
```

This checks:
- DB connectivity and `dashboard_templates` table exists
- All expected columns present (`user_id`, `dashboard_name`, `sm`, `md`, `lg`, `xl`, etc.)
- INSERT permission (inserts a test row and rolls back)
- Target table is empty (warns about duplicate risk if not)
- SQL file exists and has INSERT statements

**STOP if any check shows `[FAIL]`.** Investigate before proceeding.

> **If "Target table empty" fails:** Inspect existing rows to determine if they are test data or real user data. In production, be cautious about deleting existing rows — verify with the team first:
> ```bash
> psql -c "SELECT id, user_id, dashboard_name, created_at FROM dashboard_templates"
> ```

```bash
exit
```

---

## 14. Step 11: Import Data into Widget-Layout DB

### 14.1 Exec into the Target Debug-Container

```bash
oc -n widget-layout-backend-prod exec -it "$TARGET_POD" -- bash
```

### 14.2 Document Baseline Row Count

**Document the current row count.** This is your baseline for rollback verification.

### 14.3 Run the Import

```bash
python3 /tmp/widget-migration-script.py import
```

Or skip the interactive confirmation prompt:

```bash
python3 /tmp/widget-migration-script.py import --yes
```

The script will:
1. Show the number of INSERT statements found
2. Ask for confirmation (unless `--yes` is passed): `Proceed with import? (y/N):`
3. **Double-check the count matches your export count before typing `y`**
4. Type `y` and press Enter
5. Execute all INSERTs inside a transaction (auto-rollback on any error)
6. Print the final row count in the target table

Expected output:
```
Found N INSERT statements in /tmp/widget_migration.sql
Proceed with import? (y/N): y
Running import...
Target DB now has N dashboard_templates rows
Done.
```

### 14.4 Exit the Container

```bash
exit
```

---

## 15. Step 12: Verify the Migration

### 15.1 Via Debug-Container (primary method)

```bash
oc -n widget-layout-backend-prod exec -it "$TARGET_POD" -- bash
```

The debug-container sets `PGHOST`, `PGUSER`, `PGPASSWORD`, `PGDATABASE` automatically, so `psql` works without flags:

```bash
# Total count — should equal pre-migration count + exported rows
psql -c "SELECT COUNT(*) FROM dashboard_templates"

# Sample rows — verify user_id and dashboard_name are populated
psql -c "SELECT id, user_id, dashboard_name, name, display_name, \"default\" FROM dashboard_templates LIMIT 10"

# Check no NULL/empty user_ids (should return 0)
psql -c "SELECT COUNT(*) FROM dashboard_templates WHERE user_id IS NULL OR user_id = ''"

# Check dashboard_name matches display_name for all migrated rows (should be 0)
psql -c "SELECT COUNT(*) FROM dashboard_templates WHERE dashboard_name != display_name"

# Verify JSON columns are populated
psql -c "SELECT COUNT(*) FROM dashboard_templates WHERE sm IS NOT NULL AND sm::text != '[]'"

# Verify widget IDs are in FEO format (should show "landing-./RhelWidget" etc., not "rhel#rhel")
psql -c "SELECT id, sm->0->>'i' AS sm_first, md->0->>'i' AS md_first, lg->0->>'i' AS lg_first, xl->0->>'i' AS xl_first FROM dashboard_templates LIMIT 5"

# Confirm no old-format IDs remain in any JSONB column (should return 0)
psql -c "SELECT COUNT(*) FROM dashboard_templates WHERE sm::text LIKE '%#%' OR md::text LIKE '%#%' OR lg::text LIKE '%#%' OR xl::text LIKE '%#%'"
```

```bash
exit
```

### 15.2 Via the Widget-Layout API

Test the API to confirm the application can read migrated data:

```bash
curl -H "x-rh-identity: <base64-encoded-identity>" \
    https://<prod-api>/api/widget-layout/v1/
```

### 15.3 Via GABI (if configured)

```bash
TOKEN="sha256~<your-token>"
GABI="https://gabi-chrome-service-prod.apps.crcp01ue1.o9m8.p1.openshiftapps.com"

# Verify source data is unchanged
curl -H "Authorization: Bearer $TOKEN" "$GABI/query" \
    -d '{"query": "SELECT COUNT(*) FROM dashboard_templates"}' -s | jq
```

---

## 16. Step 13: Clean Up

### 16.1 Delete Debug-Containers (MANDATORY)

Production debug-containers must be removed immediately after use. Delete the Deployment (not just the pod — the Deployment would recreate it):

```bash
oc -n chrome-service-prod delete deployment debug-container
oc -n widget-layout-backend-prod delete deployment debug-container
```

Verify they are gone:

```bash
oc -n chrome-service-prod get deployment debug-container 2>&1 | grep "not found"
oc -n widget-layout-backend-prod get deployment debug-container 2>&1 | grep "not found"
```

### 16.2 Clean Up Local Files

```bash
rm ./widget_migration_prod.sql ./widget_migration_prod.meta
```

### 16.3 Remove Breakglass Role

Submit an MR to remove the breakglass access:

1. Remove `- $ref: /teams/insights/roles/widget-migration-breakglass.yml` from all 4 user files
2. Delete `data/teams/insights/roles/widget-migration-breakglass.yml`
3. Delete `resources/insights-prod/chrome-service-prod/widget-migration-breakglass.role.yaml`
4. Delete `resources/insights-prod/chrome-service-prod/widget-migration-breakglass.rolebinding.yaml`
5. Delete `resources/insights-prod/widget-layout-backend-prod/widget-migration-breakglass.role.yaml`
6. Delete `resources/insights-prod/widget-layout-backend-prod/widget-migration-breakglass.rolebinding.yaml`
7. Remove `widget-migration-breakglass-group` from `managedGroups` in `data/openshift/crcp01ue1/cluster.yml`
8. Remove the `provider: resource` entries for the breakglass role/rolebinding from both namespace files
9. Remove the breakglass resource paths from `resource-owner` in `platform-experience-services.yml`

Or let it expire on `2026-07-31`. However, removing promptly is preferred for production breakglass.

---

## 17. Rollback Procedure

### Automatic Rollback

The migration script wraps all INSERTs in a `BEGIN`/`COMMIT` transaction. If any INSERT fails, the entire transaction is rolled back automatically. No partial data will be written.

### Manual Rollback (after successful import)

If the data was imported successfully but needs to be removed:

**Option A: Delete all migrated rows (if target DB was empty before migration)**

```bash
oc -n widget-layout-backend-prod exec -it "$TARGET_POD" -- bash

psql -c "DELETE FROM dashboard_templates"
```

**Option B: Delete only migrated rows (if target DB had pre-existing data)**

Use a timestamp-based filter. The migration preserves original `created_at` timestamps from chrome-service, so you cannot filter by insertion time. Instead, if you recorded the max `id` before migration:

```bash
psql -c "DELETE FROM dashboard_templates WHERE id > <last_pre_migration_id>"
```

**Option C: Restore from RDS snapshot**

If a full rollback is needed and manual deletion is insufficient, restore the widget-layout-backend-prod RDS instance from a pre-migration snapshot via AWS console or app-interface.

**IMPORTANT:** After any rollback, verify the application is functioning correctly.

---

## 18. Reference

| Resource | Location |
|----------|----------|
| Migration script | `widget-migration-script.py` (local) |
| Breakglass MR branch | `widget-migration-prod-breakglass` (to be created) |
| Breakglass role file | `data/teams/insights/roles/widget-migration-breakglass.yml` |
| Chrome-service breakglass RBAC Role | `resources/insights-prod/chrome-service-prod/widget-migration-breakglass.role.yaml` |
| Chrome-service breakglass RoleBinding | `resources/insights-prod/chrome-service-prod/widget-migration-breakglass.rolebinding.yaml` |
| Widget-layout breakglass RBAC Role | `resources/insights-prod/widget-layout-backend-prod/widget-migration-breakglass.role.yaml` |
| Widget-layout breakglass RoleBinding | `resources/insights-prod/widget-layout-backend-prod/widget-migration-breakglass.rolebinding.yaml` |
| Cluster config (managedGroups) | `data/openshift/crcp01ue1/cluster.yml` |
| Debug-container docs | `docs/app-sre/sops/general/debug-container.md` |
| Self-SRE access docs | `docs/platform-users/access/self-sre-openshift-access.md` |
| DB connection docs | `docs/platform-users/external-resources/aws/rds/connect-to-postgres-mysql-database.md` |
| Chrome-service prod namespace | `data/services/insights/chrome-service/namespaces/chrome-service-prod.yml` |
| Widget-layout prod namespace | `data/services/insights/widget-layout-backend/namespaces/crcp01ue1-widget-layout-backend-prod.yml` |
| Chrome-service DB secret name | `chrome-service-db` |
| Widget-layout DB secret name | `widget-layout-backend-db` |
| Cluster | `crcp01ue1` |
| Cluster console | `https://console-openshift-console.apps.crcp01ue1.o9m8.p1.openshiftapps.com` |
| GABI endpoint (prod) | `https://gabi-chrome-service-prod.apps.crcp01ue1.o9m8.p1.openshiftapps.com` |
| Staging migration plan | `MIGRATION_STAGING.md` |
| JIRA ticket | RHCLOUD-40883 |
