# Widget Dashboard Templates Migration Plan (Staging)

Migrate the `dashboard_templates` table from chrome-service DB to widget-layout-backend DB in staging.

---

## Table of Contents

1. [Overview](#1-overview)
2. [Prerequisites](#2-prerequisites)
3. [Schema Translation](#3-schema-translation)
4. [Step 1: Obtain Staging Access (app-interface MR)](#4-step-1-obtain-staging-access-app-interface-mr)
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
| **Source DB** | chrome-service PostgreSQL (stage) |
| **Target DB** | widget-layout-backend PostgreSQL (stage) |
| **Source table** | `dashboard_templates` (chrome-service) |
| **Target table** | `dashboard_templates` (widget-layout-backend) |
| **Cluster** | `crcs02ue1` — both namespaces are on the same cluster |
| **Source namespace** | `chrome-service-stage` |
| **Target namespace** | `widget-layout-backend-stage` |
| **Source DB secret** | `chrome-service-db` (output_resource_name in app-interface) |
| **Target DB secret** | `widget-layout-backend-db` (output_resource_name in app-interface) |
| **Method** | Python script using `psycopg2` inside debug-containers |
| **Migration script** | `widget-migration-script.py` |
| **Authorized users** | tefaz, khala, bflorkie, mmarosi |

---

## 2. Prerequisites

- `oc` CLI installed locally
- Access to the `crcs02ue1` OpenShift cluster console
- The staging access MR merged (see Step 1)
- The migration script `widget-migration-script.py` available locally

---

## Lessons Learned (from 2026-05-11 staging run)

These issues were encountered during the first staging migration. The instructions below have been updated to account for them, but they are documented here for awareness:

1. **Debug-container env vars use `PG*` prefix, not `DB_*`**: The debug-container template mounts the DB secret as `PGHOST`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`, `PGPORT` — not `DB_HOST`, `DB_USER`, etc. The migration script has been updated to accept both formats automatically.

2. **Debug-container creates a Deployment, not a bare Pod**: The `oc process` command creates a Deployment, which produces pods with a hash suffix (e.g., `debug-container-cff5b7cb9-gfv9r`). You must look up the actual pod name via label selector before using `oc cp` or `oc exec`. Cleanup must delete the Deployment, not the pod.

3. **Target table may have existing rows**: If the widget-layout-backend API is already deployed in stage, users or tests may have created rows. The preflight-import check will `[FAIL]` on "Target table empty". Clear these rows before importing if they are test data.

4. **No external route for stage API**: The `widget-layout-backend-stage` namespace has no Route, so API-level verification (Step 12.3) is not available in staging. Debug-container psql verification is sufficient.

5. **`psql` reads PG* vars natively**: Inside the debug-container, you can run `psql` without `-h`/`-U`/`-d` flags — it picks up `PGHOST`, `PGUSER`, `PGPASSWORD`, `PGDATABASE` automatically.

6. **Widget item `"i"` fields must be transformed during migration**: Chrome-service stores widget item identifiers in `"shortKey#shortKey"` format (e.g., `"rhel#rhel"`), while widget-layout-backend uses `"{scope}-{module}"` format (e.g., `"landing-./RhelWidget"`). The migration script handles this via `WIDGET_ID_MAP` during export. Without this transformation, the frontend cannot match template items to widget definitions from the `/widget-mapping` endpoint.

---

## 3. Schema Translation

The two databases use different schemas for the same conceptual table. The migration script handles these translations automatically:

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

## 4. Step 1: Obtain Staging Access (app-interface MR)

The `platform-experience` and `platform-experience-services` roles only grant `view` access to the staging namespaces. To run debug-containers (which requires pod create/exec), a temporary `edit` role is needed.

### 1.1 Create the Role File

File: `data/teams/insights/roles/platform-experience-staging-dev.yml`

```yaml
---

$schema: /access/role-1.yml

labels: {}
name: platform-experience-staging-dev

description: |
  Temporary staging write access for dashboard_templates data migration
  from chrome-service to widget-layout-backend.

permissions: []

expirationDate: '2026-06-10'

access:
- namespace:
    $ref: /services/insights/chrome-service/namespaces/chrome-service-stage.yml
  clusterRole: edit
- namespace:
    $ref: /services/insights/widget-layout-backend/namespaces/crcs02ue1-widget-layout-backend-stage.yml
  clusterRole: edit
```

Key details:
- `edit` grants pod create/exec (and secret read, which is acceptable for staging)
- DB credentials are accessed indirectly via the debug-container pod mounting the secret
- `expirationDate` is set to 1 month out (2026-06-08); adjust as needed
- Both namespaces are on `crcs02ue1` cluster

### 1.2 Add Role to User Files

Add the following line to the `roles:` section of each user file in `data/teams/insights/users/`:

```yaml
- $ref: /teams/insights/roles/platform-experience-staging-dev.yml
```

Users to update:
- `tefaz.yml`
- `khala.yml`
- `bflorkie.yml`
- `mmarosi.yml`

### 1.3 Commit and Push

```bash
cd /path/to/app-interface
git fetch upstream master
git checkout -b widget-migration-staging-dev upstream/master
# ... make the changes above ...
git add data/teams/insights/roles/platform-experience-staging-dev.yml \
        data/teams/insights/users/tefaz.yml \
        data/teams/insights/users/khala.yml \
        data/teams/insights/users/bflorkie.yml \
        data/teams/insights/users/mmarosi.yml
git commit -m "Add staging dev role for widget migration"
git push origin widget-migration-staging-dev
```

### 1.4 Create Merge Request

Open the MR link provided by git push output. Target: `service/app-interface` master branch.

**Status: MR created and pushed to `origin/widget-migration-staging-dev`.**

---

## 5. Step 2: Wait for MR Merge and Reconciliation

1. Get MR reviewed and approved
2. Once merged, wait for `qontract-reconcile` to apply the RBAC changes to the cluster
3. This typically takes 5-15 minutes after merge
4. You can monitor by checking if your new role is visible:
   ```bash
   oc auth can-i create pods -n chrome-service-stage
   ```
   Should return `yes` once reconciliation is complete.

---

## 6. Step 3: Log In to the Cluster

1. Navigate to the `crcs02ue1` cluster console:
   ```
   https://console-openshift-console.apps.crcs02ue1.urby.p1.openshiftapps.com
   ```

2. Click your username (top right) -> **Copy login command** -> **Display Token**

3. Copy the `oc login` command and run it in your terminal:
   ```bash
   oc login --token=sha256~<your-token> --server=https://api.crcs02ue1.urby.p1.openshiftapps.com:6443
   ```

4. Verify login:
   ```bash
   oc whoami
   ```
   Should show your username.

---

## 7. Step 4: Verify Namespace Access

Confirm you have the required permissions in both namespaces:

```bash
# Check chrome-service-stage
oc auth can-i create pods -n chrome-service-stage
# Expected: yes

oc auth can-i create pods/exec -n chrome-service-stage
# Expected: yes

# Check widget-layout-backend-stage
oc auth can-i create pods -n widget-layout-backend-stage
# Expected: yes

oc auth can-i create pods/exec -n widget-layout-backend-stage
# Expected: yes
```

If any return `no`, the RBAC hasn't reconciled yet. Wait a few more minutes and retry.

---

## 8. Step 5: Launch Debug-Container in Chrome-Service Namespace (Source)

The debug-container template mounts the DB secret as environment variables inside the pod.

```bash
oc process --local \
    -f https://raw.githubusercontent.com/app-sre/container-images/master/debug-container/openshift.yml \
    -p POSTGRES_DB_SECRET_NAME="chrome-service-db" \
| oc -n chrome-service-stage apply -f -
```

Wait for the deployment to roll out and get the pod name:

```bash
oc -n chrome-service-stage rollout status deployment/debug-container --timeout=120s

# Get the actual pod name (it has a hash suffix)
SOURCE_POD=$(oc -n chrome-service-stage get pods -l app=debug-container -o jsonpath='{.items[0].metadata.name}')
echo "Source pod: $SOURCE_POD"
```

---

## 9. Step 6: Export Data from Chrome-Service DB

### 9.1 Copy the Migration Script into the Debug-Container

```bash
oc -n chrome-service-stage cp ./widget-migration-script.py "$SOURCE_POD:/tmp/widget-migration-script.py"
```

### 9.2 Exec into the Debug-Container

```bash
oc -n chrome-service-stage exec -it "$SOURCE_POD" -- bash
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

All checks must show `[PASS]`. If any show `[FAIL]`, investigate before proceeding.

If the DB secret is not mounted correctly:
```bash
env | grep -E 'PG|DB_'
```

> **Note:** The debug-container mounts the secret as `PG*` env vars (`PGHOST`, `PGUSER`, etc.), not `DB_*`. The migration script accepts both formats automatically.

### 9.4 Run the Export

```bash
python3 /tmp/widget-migration-script.py export
```

This will:
- Query `dashboard_templates` JOIN `user_identities` from chrome-service DB
- Translate `user_identity_id` (uint FK) to `user_id` (string account_id)
- Copy `display_name` to `dashboard_name`
- Transform widget item `"i"` fields in JSONB columns from chrome-service format (e.g., `"rhel#rhel"`) to FEO format (e.g., `"landing-./RhelWidget"`)
- Generate `/tmp/widget_migration.sql` with INSERT statements wrapped in a transaction

Expected output:
```
Found N active dashboard_templates rows
Fetched N rows (with user_id resolved)
Transformed widget IDs in JSONB columns (sm/md/lg/xl)
Generated N INSERT statements in /tmp/widget_migration.sql
```

### 9.5 Quick Review Inside the Container

```bash
head -20 /tmp/widget_migration.sql
wc -l /tmp/widget_migration.sql
```

### 9.6 Exit the Container

```bash
exit
```

---

## 10. Step 7: Transfer Files to Local Machine

Copy the generated SQL from the source debug-container to your local machine:

```bash
oc -n chrome-service-stage cp "$SOURCE_POD:/tmp/widget_migration.sql" ./widget_migration.sql
```

Verify the file was copied:

```bash
head -20 ./widget_migration.sql
grep -c "^INSERT" ./widget_migration.sql
```

---

## 11. Step 8: Launch Debug-Container in Widget-Layout Namespace (Target)

```bash
oc process --local \
    -f https://raw.githubusercontent.com/app-sre/container-images/master/debug-container/openshift.yml \
    -p POSTGRES_DB_SECRET_NAME="widget-layout-backend-db" \
| oc -n widget-layout-backend-stage apply -f -
```

Wait for the deployment to roll out and get the pod name:

```bash
oc -n widget-layout-backend-stage rollout status deployment/debug-container --timeout=120s

# Get the actual pod name
TARGET_POD=$(oc -n widget-layout-backend-stage get pods -l app=debug-container -o jsonpath='{.items[0].metadata.name}')
echo "Target pod: $TARGET_POD"
```

---

## 12. Step 9: Transfer Files to Target Debug-Container

```bash
oc -n widget-layout-backend-stage cp ./widget_migration.sql "$TARGET_POD:/tmp/widget_migration.sql"
oc -n widget-layout-backend-stage cp ./widget-migration-script.py "$TARGET_POD:/tmp/widget-migration-script.py"
```

---

## 13. Step 10: Run Preflight Check on Target DB

```bash
oc -n widget-layout-backend-stage exec -it "$TARGET_POD" -- bash
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

All checks must show `[PASS]`. If any show `[FAIL]`, investigate before proceeding.

> **If "Target table empty" fails:** The stage API may have created test rows. Verify they are test data, then clear them:
> ```bash
> psql -c "SELECT id, user_id, dashboard_name, created_at FROM dashboard_templates"
> psql -c "DELETE FROM dashboard_templates"
> ```
> Re-run preflight-import after clearing.

```bash
exit
```

---

## 14. Step 11: Import Data into Widget-Layout DB

### 14.1 Exec into the Target Debug-Container

```bash
oc -n widget-layout-backend-stage exec -it "$TARGET_POD" -- bash
```

### 14.2 Run the Import

```bash
python3 /tmp/widget-migration-script.py import
```

The script will:
1. Show the number of INSERT statements found
2. Ask for confirmation: `Proceed with import? (y/N):`
3. Type `y` and press Enter
4. Execute all INSERTs inside a transaction (auto-rollback on any error)
5. Print the final row count in the target table

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
oc -n widget-layout-backend-stage exec -it "$TARGET_POD" -- bash
```

The debug-container sets `PGHOST`, `PGUSER`, `PGPASSWORD`, `PGDATABASE` automatically, so `psql` works without flags:

```bash
# Total count — must match source export count
psql -c "SELECT COUNT(*) FROM dashboard_templates"

# Sample a few rows
psql -c "SELECT id, user_id, dashboard_name, name, display_name, \"default\" FROM dashboard_templates LIMIT 5"

# Check user_id is populated (should return 0)
psql -c "SELECT COUNT(*) FROM dashboard_templates WHERE user_id IS NULL OR user_id = ''"

# Verify JSON columns are populated
psql -c "SELECT COUNT(*) FROM dashboard_templates WHERE sm IS NOT NULL AND sm::text != '[]'"

# Verify widget IDs are in FEO format (should show "landing-./RhelWidget" etc., not "rhel#rhel")
psql -c "SELECT id, sm->0->>'i' AS first_widget_id FROM dashboard_templates LIMIT 5"

# Confirm no old-format IDs remain (should return 0)
psql -c "SELECT COUNT(*) FROM dashboard_templates WHERE sm::text LIKE '%#%'"
```

```bash
exit
```

### 15.2 Via GABI (not available for widget-layout-backend in staging)

GABI is not yet configured for widget-layout-backend. Debug-container psql verification (15.1) is sufficient.

If GABI becomes available in the future:

```bash
TOKEN="sha256~<your-token>"
GABI="https://<widget-layout-gabi-endpoint>"

curl -H "Authorization: Bearer $TOKEN" "$GABI/query" \
    -d '{"query": "SELECT COUNT(*) FROM dashboard_templates"}' -s | jq
```

### 15.3 Via the Widget-Layout API (not available in staging)

The `widget-layout-backend-stage` namespace has no external Route. This verification method is not available in staging. Skip to cleanup.

---

## 16. Step 13: Clean Up

### 16.1 Delete Debug-Containers

The debug-container is a Deployment — delete the Deployment, not just the pod (deleting the pod alone causes the Deployment to recreate it):

```bash
oc -n chrome-service-stage delete deployment debug-container
oc -n widget-layout-backend-stage delete deployment debug-container
```

### 16.2 Clean Up Local Files

```bash
rm ./widget_migration.sql
```

### 16.3 Remove Staging Dev Role (after migration is verified)

Once the migration is confirmed successful and the role is no longer needed, submit an MR to:

1. Remove `- $ref: /teams/insights/roles/platform-experience-staging-dev.yml` from all 4 user files
2. Delete `data/teams/insights/roles/platform-experience-staging-dev.yml`

Or simply let it expire on `2026-06-08`.

---

## 17. Rollback Procedure

If something goes wrong during import:

### Automatic Rollback

The migration script wraps all INSERTs in a `BEGIN`/`COMMIT` transaction. If any INSERT fails, the entire transaction is rolled back automatically. No partial data will be written.

### Manual Rollback (after successful import)

If the data was imported but needs to be removed:

```bash
oc -n widget-layout-backend-stage exec -it "$TARGET_POD" -- bash

psql -c "DELETE FROM dashboard_templates WHERE created_at < '2026-05-08'"
```

Adjust the `WHERE` clause to target only the migrated rows. Use timestamps or other identifying criteria to avoid deleting post-migration data.

---

## 18. Reference

| Resource | Location |
|----------|----------|
| Migration script | `widget-migration-script.py` (local) |
| Access MR branch | `widget-migration-staging-dev` |
| Role file | `data/teams/insights/roles/platform-experience-staging-dev.yml` |
| Debug-container docs | `docs/app-sre/sops/general/debug-container.md` |
| Self-SRE access docs | `docs/platform-users/access/self-sre-openshift-access.md` |
| DB connection docs | `docs/platform-users/external-resources/aws/rds/connect-to-postgres-mysql-database.md` |
| Chrome-service namespace | `data/services/insights/chrome-service/namespaces/chrome-service-stage.yml` |
| Widget-layout namespace | `data/services/insights/widget-layout-backend/namespaces/crcs02ue1-widget-layout-backend-stage.yml` |
| Chrome-service DB secret name | `chrome-service-db` |
| Widget-layout DB secret name | `widget-layout-backend-db` |
| Cluster | `crcs02ue1` |
| Cluster console | `https://console-openshift-console.apps.crcs02ue1.urby.p1.openshiftapps.com` |
| JIRA ticket | RHCLOUD-40883 |
