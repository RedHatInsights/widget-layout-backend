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
| **Access type** | `edit-no-secrets` (staging dev role) | Breakglass role (scoped RBAC Role) |
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

---

## 3. Schema Translation

Same as staging. The migration script handles these translations automatically:

| Chrome-Service Column | Widget-Layout Column | Translation |
|-----------------------|----------------------|-------------|
| `id` (uint, PK) | `id` (uint, PK) | Auto-generated in target (not copied) |
| `user_identity_id` (uint, FK) | `user_id` (string) | JOIN `user_identities` table to resolve `account_id` |
| _(does not exist)_ | `dashboard_name` (string) | Copied from `display_name` |
| `default` (bool) | `default` (bool) | Direct copy |
| `name` (string, embedded) | `name` (string, embedded) | Direct copy |
| `display_name` (string, embedded) | `display_name` (string, embedded) | Direct copy |
| `sm` (JSON) | `sm` (JSON) | Direct copy |
| `md` (JSON) | `md` (JSON) | Direct copy |
| `lg` (JSON) | `lg` (JSON) | Direct copy |
| `xl` (JSON) | `xl` (JSON) | Direct copy |
| `created_at` (timestamp) | `created_at` (timestamp) | Direct copy |
| `updated_at` (timestamp) | `updated_at` (timestamp) | Direct copy |
| `deleted_at` (timestamp) | `deleted_at` (timestamp) | Direct copy |

> **Note on `x`/`y` vs `cx`/`cy` coordinates:** The widget grid items stored in the `sm`/`md`/`lg`/`xl` JSONB columns use `x`/`y` keys. The widget-layout-backend API also uses `x`/`y` at runtime. The `cx`/`cy` naming is only required in YAML configuration files (ConfigMaps, env vars) because YAML parsers treat bare `y` as a boolean. Since this migration copies JSONB directly between PostgreSQL databases — never passing through a YAML parser — no coordinate key renaming is needed.

---

## 4. Step 1: Obtain Production Breakglass Access (app-interface MR)

Production requires a **breakglass role** — `edit-no-secrets` is prohibited on production namespaces per Self-SRE OpenShift Access Standards. Breakglass uses a scoped Kubernetes RBAC Role (not ClusterRole) with explicit permissions.

### 4.1 Create the Breakglass RBAC Role Resource

File: `resources/services/insights/chrome-service/rbac/widget-migration-breakglass.role.yaml`

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: widget-migration-breakglass
rules:
# Create and manage debug pods
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["create", "list", "get", "delete"]
# Exec into pods for running migration
- apiGroups: [""]
  resources: ["pods/exec"]
  verbs: ["create", "get"]
# Read DB secret (named explicitly)
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["chrome-service-db"]
  verbs: ["get"]
```

Create a similar file for widget-layout-backend:

File: `resources/services/insights/widget-layout-backend/rbac/widget-migration-breakglass.role.yaml`

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: widget-migration-breakglass
rules:
# Create and manage debug pods
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["create", "list", "get", "delete"]
# Exec into pods for running migration
- apiGroups: [""]
  resources: ["pods/exec"]
  verbs: ["create", "get"]
# Read DB secret (named explicitly)
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["widget-layout-backend-db"]
  verbs: ["get"]
```

### 4.2 Create the Breakglass Role File

File: `data/teams/insights/roles/platform-experience-breakglass.yml`

```yaml
---

$schema: /access/role-1.yml

labels: {}
name: platform-experience-breakglass

description: |
  Breakglass role for production dashboard_templates data migration
  from chrome-service to widget-layout-backend. Scoped to debug pod
  creation, exec, and named DB secret access only.

permissions: []

expirationDate: '2026-06-08'

access:
- namespace:
    $ref: /services/insights/chrome-service/namespaces/chrome-service-prod.yml
  role: widget-migration-breakglass
- namespace:
    $ref: /services/insights/widget-layout-backend/namespaces/crcp01ue1-widget-layout-backend-prod.yml
  role: widget-migration-breakglass

self_service:
- change_type:
    $ref: /app-interface/changetype/breakglass-role-manager.yml
  datafiles:
  - $ref: /teams/insights/roles/platform-experience-breakglass.yml
```

### 4.3 Add Role to User Files

Add the following line to the `roles:` section of each user file in `data/teams/insights/users/`:

```yaml
- $ref: /teams/insights/roles/platform-experience-breakglass.yml
```

Users to update:
- `tefaz.yml`
- `khala.yml`
- `bflorkie.yml`
- `mmarosi.yml`

### 4.4 Add Resource Paths to Self-Service

Ensure the breakglass RBAC resource files are listed in the team's `resource-owner` self-service section so changes to these resources can be self-approved. Check the `platform-experience` or `platform-experience-services` role file and add:

```yaml
- change_type:
    $ref: /app-interface/changetype/resource-owner.yml
  resources:
  - /services/insights/chrome-service/rbac/widget-migration-breakglass.role.yaml
  - /services/insights/widget-layout-backend/rbac/widget-migration-breakglass.role.yaml
```

### 4.5 Commit, Push, and Create MR

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

Wait for the pod to be ready:

```bash
oc -n chrome-service-prod get pod debug-container -w
```

Wait until `STATUS` shows `Running`. Press `Ctrl+C` to stop watching.

---

## 9. Step 6: Export Data from Chrome-Service DB

### 9.1 Copy the Migration Script into the Debug-Container

```bash
oc -n chrome-service-prod cp ./widget-migration-script.py debug-container:/tmp/widget-migration-script.py
```

### 9.2 Exec into the Debug-Container

```bash
oc -n chrome-service-prod exec -it debug-container -- bash
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
env | grep DB_
```

### 9.4 Run the Export

```bash
python3 /tmp/widget-migration-script.py export
```

Expected output:
```
Found N active dashboard_templates rows
Fetched N rows (with user_id resolved)
Generated N INSERT statements in /tmp/widget_migration.sql
```

### 9.5 Verify Export Data

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

### 9.6 Exit the Container

```bash
exit
```

---

## 10. Step 7: Transfer Files to Local Machine

```bash
oc -n chrome-service-prod cp debug-container:/tmp/widget_migration.sql ./widget_migration_prod.sql
```

Verify the file was copied:

```bash
head -20 ./widget_migration_prod.sql
grep -c "^INSERT" ./widget_migration_prod.sql
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

Wait for the pod to be ready:

```bash
oc -n widget-layout-backend-prod get pod debug-container -w
```

Wait until `STATUS` shows `Running`. Press `Ctrl+C` to stop watching.

---

## 12. Step 9: Transfer Files to Target Debug-Container

```bash
oc -n widget-layout-backend-prod cp ./widget_migration_prod.sql debug-container:/tmp/widget_migration.sql
oc -n widget-layout-backend-prod cp ./widget-migration-script.py debug-container:/tmp/widget-migration-script.py
```

---

## 13. Step 10: Run Preflight Check on Target DB

**This check is mandatory for production.** Do not skip.

```bash
oc -n widget-layout-backend-prod exec -it debug-container -- bash
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

```bash
exit
```

---

## 14. Step 11: Import Data into Widget-Layout DB

### 14.1 Exec into the Target Debug-Container

```bash
oc -n widget-layout-backend-prod exec -it debug-container -- bash
```

**Document the current row count.** This is your baseline for rollback verification.

### 14.3 Run the Import

```bash
python3 /tmp/widget-migration-script.py import
```

The script will:
1. Show the number of INSERT statements found
2. Ask for confirmation: `Proceed with import? (y/N):`
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

### 15.1 Via Debug-Container

```bash
oc -n widget-layout-backend-prod exec -it debug-container -- bash
```

```bash
# Total count — should equal pre-migration count + exported rows
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT COUNT(*) FROM dashboard_templates"

# Sample rows — verify user_id and dashboard_name are populated
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT id, user_id, dashboard_name, name, display_name, \"default\" FROM dashboard_templates LIMIT 10"

# Check no NULL/empty user_ids (should return 0)
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT COUNT(*) FROM dashboard_templates WHERE user_id IS NULL OR user_id = ''"

# Check dashboard_name matches display_name for all migrated rows
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT COUNT(*) FROM dashboard_templates WHERE dashboard_name != display_name"
# Should be 0 for migrated rows
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

Production debug-containers must be removed immediately after use.

```bash
oc -n chrome-service-prod delete pod debug-container
oc -n widget-layout-backend-prod delete pod debug-container
```

Verify they are gone:

```bash
oc -n chrome-service-prod get pod debug-container 2>&1 | grep "not found"
oc -n widget-layout-backend-prod get pod debug-container 2>&1 | grep "not found"
```

### 16.2 Clean Up Local Files

```bash
rm ./widget_migration_prod.sql
```

### 16.3 Remove Breakglass Role

Submit an MR to remove the breakglass access:

1. Remove `- $ref: /teams/insights/roles/platform-experience-breakglass.yml` from all 4 user files
2. Delete `data/teams/insights/roles/platform-experience-breakglass.yml`
3. Delete `resources/services/insights/chrome-service/rbac/widget-migration-breakglass.role.yaml`
4. Delete `resources/services/insights/widget-layout-backend/rbac/widget-migration-breakglass.role.yaml`

Or let it expire on `2026-06-08`. However, removing promptly is preferred for production breakglass.

---

## 17. Rollback Procedure

### Automatic Rollback

The migration script wraps all INSERTs in a `BEGIN`/`COMMIT` transaction. If any INSERT fails, the entire transaction is rolled back automatically. No partial data will be written.

### Manual Rollback (after successful import)

If the data was imported successfully but needs to be removed:

**Option A: Delete all migrated rows (if target DB was empty before migration)**

```bash
oc -n widget-layout-backend-prod exec -it debug-container -- bash

PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" -d "$DB_NAME" \
    -c "DELETE FROM dashboard_templates"
```

**Option B: Delete only migrated rows (if target DB had pre-existing data)**

Use a timestamp-based filter. The migration preserves original `created_at` timestamps from chrome-service, so you cannot filter by insertion time. Instead, if you recorded the max `id` before migration:

```bash
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" -d "$DB_NAME" \
    -c "DELETE FROM dashboard_templates WHERE id > <last_pre_migration_id>"
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
| Breakglass role file | `data/teams/insights/roles/platform-experience-breakglass.yml` |
| Chrome-service breakglass RBAC | `resources/services/insights/chrome-service/rbac/widget-migration-breakglass.role.yaml` |
| Widget-layout breakglass RBAC | `resources/services/insights/widget-layout-backend/rbac/widget-migration-breakglass.role.yaml` |
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
