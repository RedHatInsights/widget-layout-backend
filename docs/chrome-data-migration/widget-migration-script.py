#!/usr/bin/env python3
"""
Widget Data Migration Script
Migrates dashboard_templates from chrome-service DB to widget-layout-backend DB.

Schema translation:
  chrome-service.user_identity_id (uint FK) -> widget-layout.user_id (string)
    resolved via JOIN on user_identities.account_id
  chrome-service.display_name -> widget-layout.dashboard_name (copy)

Usage:
  1. Copy this script into chrome-service debug-container
  2. Run: python3 widget-migration-script.py preflight-export
  3. Run: python3 widget-migration-script.py export
  4. Copy /tmp/widget_migration.sql out, then into widget-layout debug-container
  5. Run: python3 widget-migration-script.py preflight-import
  6. Run: python3 widget-migration-script.py import
"""

import json
import os
import sys

import psycopg2
import psycopg2.extras

OUTPUT_FILE = "/tmp/widget_migration.sql"


def get_connection():
    # Debug-containers mount secrets as PG* vars; the script originally used DB_*.
    # Accept both formats: try PG* first (debug-container default), fall back to DB_*.
    host = os.environ.get("PGHOST") or os.environ.get("DB_HOST")
    port = os.environ.get("PGPORT") or os.environ.get("DB_PORT", "5432")
    user = os.environ.get("PGUSER") or os.environ.get("DB_USER")
    password = os.environ.get("PGPASSWORD") or os.environ.get("DB_PASSWORD")
    dbname = os.environ.get("PGDATABASE") or os.environ.get("DB_NAME")

    missing = []
    if not host:
        missing.append("PGHOST/DB_HOST")
    if not user:
        missing.append("PGUSER/DB_USER")
    if not password:
        missing.append("PGPASSWORD/DB_PASSWORD")
    if not dbname:
        missing.append("PGDATABASE/DB_NAME")

    if missing:
        print(f"ERROR: Missing env vars: {', '.join(missing)}")
        print("Is the DB secret mounted? Run: env | grep -E 'PG|DB_'")
        sys.exit(1)

    return psycopg2.connect(
        host=host,
        port=port,
        user=user,
        password=password,
        dbname=dbname,
    )


def sql_value(val):
    if val is None:
        return "NULL"
    if isinstance(val, bool):
        return "true" if val else "false"
    if isinstance(val, (dict, list)):
        return "'" + json.dumps(val).replace("'", "''") + "'::jsonb"
    s = str(val).replace("'", "''")
    return f"'{s}'"


def check(label, ok, detail=""):
    status = "PASS" if ok else "FAIL"
    msg = f"  [{status}] {label}"
    if detail:
        msg += f" — {detail}"
    print(msg)
    return ok


def do_preflight_export():
    print("=== Preflight: Source DB (chrome-service) ===\n")
    all_ok = True

    # 1. Connectivity
    try:
        conn = get_connection()
        cur = conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)
        all_ok &= check("DB connectivity", True)
    except Exception as e:
        check("DB connectivity", False, str(e))
        sys.exit(1)

    # 2. SELECT permission on dashboard_templates
    try:
        cur.execute("SELECT COUNT(*) as cnt FROM dashboard_templates")
        total = cur.fetchone()["cnt"]
        all_ok &= check("SELECT on dashboard_templates", True, f"{total} total rows")
    except Exception as e:
        all_ok &= check("SELECT on dashboard_templates", False, str(e))
        cur.close()
        conn.close()
        print(f"\n{'READY' if all_ok else 'NOT READY'} for export.")
        sys.exit(0 if all_ok else 1)

    # 3. SELECT permission on user_identities
    try:
        cur.execute("SELECT COUNT(*) as cnt FROM user_identities")
        ui_total = cur.fetchone()["cnt"]
        all_ok &= check("SELECT on user_identities", True, f"{ui_total} total rows")
    except Exception as e:
        all_ok &= check("SELECT on user_identities", False, str(e))

    # 4. Active (non-deleted) rows
    cur.execute("SELECT COUNT(*) as cnt FROM dashboard_templates WHERE deleted_at IS NULL")
    active = cur.fetchone()["cnt"]
    all_ok &= check("Active dashboard_templates", active > 0, f"{active} rows (deleted_at IS NULL)")

    # 5. Orphaned active rows (no matching user_identity)
    cur.execute("""
        SELECT COUNT(*) as cnt FROM dashboard_templates dt
        LEFT JOIN user_identities ui ON dt.user_identity_id = ui.id
        WHERE dt.deleted_at IS NULL AND ui.id IS NULL
    """)
    orphaned = cur.fetchone()["cnt"]
    all_ok &= check("Orphaned rows (no user_identity)", orphaned == 0,
                     f"{orphaned} active rows will be dropped by JOIN" if orphaned > 0 else "none")

    # 6. Active users with NULL/empty account_id
    cur.execute("""
        SELECT COUNT(*) as cnt FROM dashboard_templates dt
        JOIN user_identities ui ON dt.user_identity_id = ui.id
        WHERE dt.deleted_at IS NULL AND (ui.account_id IS NULL OR ui.account_id = '')
    """)
    null_acct = cur.fetchone()["cnt"]
    all_ok &= check("NULL/empty account_id", null_acct == 0,
                     f"{null_acct} rows would get NULL user_id in target" if null_acct > 0 else "none")

    # 7. Row count after JOIN (what export will actually produce)
    cur.execute("""
        SELECT COUNT(*) as cnt FROM dashboard_templates dt
        JOIN user_identities ui ON dt.user_identity_id = ui.id
        WHERE dt.deleted_at IS NULL
    """)
    join_count = cur.fetchone()["cnt"]
    all_ok &= check("Exportable rows (after JOIN)", join_count > 0, f"{join_count} rows")

    cur.close()
    conn.close()

    print(f"\n{'READY' if all_ok else 'NOT READY'} for export.")
    if not all_ok:
        sys.exit(1)


def do_preflight_import():
    print("=== Preflight: Target DB (widget-layout-backend) ===\n")
    all_ok = True

    # 1. Connectivity
    try:
        conn = get_connection()
        cur = conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)
        all_ok &= check("DB connectivity", True)
    except Exception as e:
        check("DB connectivity", False, str(e))
        sys.exit(1)

    # 2. dashboard_templates table exists
    try:
        cur.execute("""
            SELECT COUNT(*) as cnt FROM information_schema.tables
            WHERE table_schema = 'public' AND table_name = 'dashboard_templates'
        """)
        exists = cur.fetchone()["cnt"] > 0
        all_ok &= check("dashboard_templates table exists", exists)
    except Exception as e:
        all_ok &= check("dashboard_templates table exists", False, str(e))

    # 3. Expected columns exist
    expected_cols = {"user_id", "dashboard_name", "default", "name", "display_name",
                     "sm", "md", "lg", "xl", "created_at", "updated_at", "deleted_at"}
    try:
        cur.execute("""
            SELECT column_name FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'dashboard_templates'
        """)
        actual_cols = {row["column_name"] for row in cur.fetchall()}
        missing = expected_cols - actual_cols
        all_ok &= check("Expected columns present", len(missing) == 0,
                         f"missing: {', '.join(sorted(missing))}" if missing else "all found")
    except Exception as e:
        all_ok &= check("Expected columns present", False, str(e))

    # 4. INSERT permission (dry test)
    try:
        cur.execute("BEGIN")
        cur.execute("""
            INSERT INTO dashboard_templates (user_id, dashboard_name, "default", name, display_name, sm, md, lg, xl, created_at, updated_at, deleted_at)
            VALUES ('__preflight_test__', '__test__', false, '__test__', '__test__', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, NOW(), NOW(), NULL)
        """)
        conn.rollback()
        all_ok &= check("INSERT permission", True, "test row inserted and rolled back")
    except Exception as e:
        conn.rollback()
        all_ok &= check("INSERT permission", False, str(e))

    # 5. Current row count (duplicate risk)
    try:
        cur.execute("SELECT COUNT(*) as cnt FROM dashboard_templates")
        existing = cur.fetchone()["cnt"]
        if existing > 0:
            all_ok &= check("Target table empty", False, f"{existing} rows already exist — risk of duplicates")
        else:
            all_ok &= check("Target table empty", True, "0 rows")
    except Exception as e:
        all_ok &= check("Target table empty", False, str(e))

    # 6. SQL file exists
    sql_exists = os.path.exists(OUTPUT_FILE)
    if sql_exists:
        with open(OUTPUT_FILE) as f:
            insert_count = f.read().count("INSERT INTO")
        all_ok &= check(f"SQL file exists ({OUTPUT_FILE})", True, f"{insert_count} INSERT statements")
    else:
        all_ok &= check(f"SQL file exists ({OUTPUT_FILE})", False, "copy it from source debug-container first")

    cur.close()
    conn.close()

    print(f"\n{'READY' if all_ok else 'NOT READY'} for import.")
    if not all_ok:
        sys.exit(1)


def do_export():
    conn = get_connection()
    cur = conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)

    cur.execute("SELECT COUNT(*) as cnt FROM dashboard_templates WHERE deleted_at IS NULL")
    total = cur.fetchone()["cnt"]
    print(f"Found {total} active dashboard_templates rows")

    if total == 0:
        print("No records to migrate.")
        return

    cur.execute("""
        SELECT
            ui.account_id AS user_id,
            dt.display_name AS dashboard_name,
            dt."default",
            dt.name,
            dt.display_name,
            dt.sm,
            dt.md,
            dt.lg,
            dt.xl,
            dt.created_at,
            dt.updated_at,
            dt.deleted_at
        FROM dashboard_templates dt
        JOIN user_identities ui ON dt.user_identity_id = ui.id
        WHERE dt.deleted_at IS NULL
    """)

    rows = cur.fetchall()
    print(f"Fetched {len(rows)} rows (with user_id resolved)")

    bad_rows = [r for r in rows if not r["user_id"]]
    if bad_rows:
        print(f"ERROR: {len(bad_rows)} rows have NULL/empty user_id:")
        for r in bad_rows:
            print(f"  name={r['name']}, display_name={r['display_name']}")
        print("Aborting — fix user_identities data before retrying.")
        cur.close()
        conn.close()
        sys.exit(1)

    with open(OUTPUT_FILE, "w") as f:
        f.write("-- Widget Migration: chrome-service -> widget-layout-backend\n")
        f.write("-- Auto-generated. Review before running.\n\n")
        f.write("BEGIN;\n\n")

        for row in rows:
            cols = "user_id, dashboard_name, \"default\", name, display_name, sm, md, lg, xl, created_at, updated_at, deleted_at"
            vals = ", ".join([
                sql_value(row["user_id"]),
                sql_value(row["dashboard_name"]),
                sql_value(row["default"]),
                sql_value(row["name"]),
                sql_value(row["display_name"]),
                sql_value(row["sm"]),
                sql_value(row["md"]),
                sql_value(row["lg"]),
                sql_value(row["xl"]),
                sql_value(row["created_at"]),
                sql_value(row["updated_at"]),
                sql_value(row["deleted_at"]),
            ])
            f.write(f"INSERT INTO dashboard_templates ({cols})\nVALUES ({vals});\n\n")

        f.write("COMMIT;\n")

    insert_count = sum(1 for line in open(OUTPUT_FILE) if line.startswith("INSERT"))
    print(f"Generated {insert_count} INSERT statements in {OUTPUT_FILE}")
    print()
    print("Next steps:")
    print(f"  1. Review {OUTPUT_FILE}")
    print(f"  2. Copy out: oc -n chrome-service-stage cp debug-container:{OUTPUT_FILE} ./widget_migration.sql")
    print(f"  3. Copy in:  oc -n widget-layout-backend-stage cp ./widget_migration.sql debug-container:{OUTPUT_FILE}")
    print("  4. Run import: python3 widget-migration-script.py import")

    cur.close()
    conn.close()


def do_import():
    if not os.path.exists(OUTPUT_FILE):
        print(f"ERROR: {OUTPUT_FILE} not found.")
        print("Copy it from the export debug-container first.")
        sys.exit(1)

    with open(OUTPUT_FILE) as f:
        sql = f.read()

    insert_count = sql.count("INSERT INTO")
    print(f"Found {insert_count} INSERT statements in {OUTPUT_FILE}")

    confirm = input("Proceed with import? (y/N): ").strip().lower()
    if confirm != "y":
        print("Aborted.")
        return

    conn = get_connection()
    cur = conn.cursor()

    print("Running import...")
    try:
        cur.execute(sql)
        conn.commit()
    except psycopg2.Error as e:
        conn.rollback()
        print(f"ERROR: Import failed, transaction rolled back: {e}")
        cur.close()
        conn.close()
        sys.exit(1)

    cur.execute("SELECT COUNT(*) FROM dashboard_templates")
    count = cur.fetchone()[0]
    print(f"Target DB now has {count} dashboard_templates rows")
    print("Done.")

    cur.close()
    conn.close()


if __name__ == "__main__":
    commands = ("preflight-export", "preflight-import", "export", "import")
    if len(sys.argv) != 2 or sys.argv[1] not in commands:
        print("Usage: python3 widget-migration-script.py <command>")
        print()
        print("  preflight-export  - Run inside chrome-service debug-container")
        print("                      Checks connectivity, permissions, data quality")
        print()
        print("  preflight-import  - Run inside widget-layout-backend debug-container")
        print("                      Checks connectivity, permissions, schema, target state")
        print()
        print("  export            - Run inside chrome-service debug-container")
        print("                      Generates SQL file from source DB")
        print()
        print("  import            - Run inside widget-layout-backend debug-container")
        print("                      Executes SQL file against target DB")
        sys.exit(1)

    cmd = sys.argv[1]
    if cmd == "preflight-export":
        do_preflight_export()
    elif cmd == "preflight-import":
        do_preflight_import()
    elif cmd == "export":
        do_export()
    else:
        do_import()
