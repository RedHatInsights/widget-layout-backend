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
13. [Step 10: Review the Generated SQL](#13-step-10-review-the-generated-sql)
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

## 3. Schema Translation

The two databases use different schemas for the same conceptual table. The migration script handles these translations automatically:

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

---

## 4. Step 1: Obtain Staging Access (app-interface MR)

The `platform-experience` and `platform-experience-services` roles only grant `view` access to the staging namespaces. To run debug-containers (which requires pod create/exec and indirect secret access), a temporary `edit-no-secrets` role is needed.

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

expirationDate: '2026-06-08'

access:
- namespace:
    $ref: /services/insights/chrome-service/namespaces/chrome-service-stage.yml
  clusterRole: edit-no-secrets
- namespace:
    $ref: /services/insights/widget-layout-backend/namespaces/crcs02ue1-widget-layout-backend-stage.yml
  clusterRole: edit-no-secrets
```

Key details:
- `edit-no-secrets` grants pod create/exec but **not** direct secret read
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

Wait for the pod to be ready:

```bash
oc -n chrome-service-stage get pod debug-container -w
```

Wait until `STATUS` shows `Running`. Press `Ctrl+C` to stop watching.

---

## 9. Step 6: Export Data from Chrome-Service DB

### 9.1 Copy the Migration Script into the Debug-Container

```bash
oc -n chrome-service-stage cp ./widget-migration-script.py debug-container:/tmp/widget-migration-script.py
```

### 9.2 Exec into the Debug-Container

```bash
oc -n chrome-service-stage exec -it debug-container -- bash
```

### 9.3 Verify DB Connectivity

Inside the container, the DB secret is available as environment variables. Test the connection:

```bash
echo "Host: $DB_HOST"
echo "User: $DB_USER"
echo "DB:   $DB_NAME"
echo "Port: ${DB_PORT:-5432}"

PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT COUNT(*) FROM dashboard_templates"
```

You should see a row count. If this fails, verify the secret name is correct:
```bash
env | grep DB_
```

### 9.4 Run the Export

```bash
python3 /tmp/widget-migration-script.py export
```

This will:
- Query `dashboard_templates` JOIN `user_identities` from chrome-service DB
- Translate `user_identity_id` (uint FK) to `user_id` (string account_id)
- Copy `display_name` to `dashboard_name`
- Generate `/tmp/widget_migration.sql` with INSERT statements wrapped in a transaction

Expected output:
```
Found N active dashboard_templates rows
Fetched N rows (with user_id resolved)
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

Copy the generated SQL and the migration script from the source debug-container to your local machine:

```bash
oc -n chrome-service-stage cp debug-container:/tmp/widget_migration.sql ./widget_migration.sql
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

Wait for the pod to be ready:

```bash
oc -n widget-layout-backend-stage get pod debug-container -w
```

Wait until `STATUS` shows `Running`. Press `Ctrl+C` to stop watching.

---

## 12. Step 9: Transfer Files to Target Debug-Container

```bash
oc -n widget-layout-backend-stage cp ./widget_migration.sql debug-container:/tmp/widget_migration.sql
oc -n widget-layout-backend-stage cp ./widget-migration-script.py debug-container:/tmp/widget-migration-script.py
```

---

## 13. Step 10: Review the Generated SQL

Before importing, review the SQL to ensure correctness.

```bash
oc -n widget-layout-backend-stage exec -it debug-container -- bash
```

Inside the container:

```bash
# Check first few INSERT statements
head -30 /tmp/widget_migration.sql

# Count total inserts
grep -c "^INSERT" /tmp/widget_migration.sql

# Check for any NULL user_ids (would indicate failed JOIN)
grep "VALUES (NULL" /tmp/widget_migration.sql || echo "No NULL user_ids - good"

# Exit for now
exit
```

---

## 14. Step 11: Import Data into Widget-Layout DB

### 14.1 Exec into the Target Debug-Container

```bash
oc -n widget-layout-backend-stage exec -it debug-container -- bash
```

### 14.2 Verify Target DB Connectivity

```bash
echo "Host: $DB_HOST"
echo "User: $DB_USER"
echo "DB:   $DB_NAME"

PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT COUNT(*) FROM dashboard_templates"
```

Note the current row count (should be 0 or whatever exists before migration).

### 14.3 Run the Import

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

### 15.1 Via Debug-Container

```bash
oc -n widget-layout-backend-stage exec -it debug-container -- bash
```

```bash
# Total count
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT COUNT(*) FROM dashboard_templates"

# Sample a few rows
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT id, user_id, dashboard_name, name, display_name, \"default\" FROM dashboard_templates LIMIT 5"

# Check user_id is populated (should return 0)
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT COUNT(*) FROM dashboard_templates WHERE user_id IS NULL OR user_id = ''"
```

```bash
exit
```

### 15.2 Via GABI (if available for widget-layout-backend)

If you have a GABI instance for widget-layout-backend:

```bash
TOKEN="sha256~<your-token>"
GABI="https://<widget-layout-gabi-endpoint>"

curl -H "Authorization: Bearer $TOKEN" "$GABI/query" \
    -d '{"query": "SELECT COUNT(*) FROM dashboard_templates"}' -s | jq
```

### 15.3 Via the Widget-Layout API (if deployed in stage)

```bash
curl -H "x-rh-identity: <base64-encoded-identity>" \
    https://<stage-api>/api/widget-layout/v1/
```

---

## 16. Step 13: Clean Up

### 16.1 Delete Debug-Containers

```bash
oc -n chrome-service-stage delete pod debug-container
oc -n widget-layout-backend-stage delete pod debug-container
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
oc -n widget-layout-backend-stage exec -it debug-container -- bash

PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" -d "$DB_NAME" \
    -c "DELETE FROM dashboard_templates WHERE created_at < '2026-05-08'"
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
