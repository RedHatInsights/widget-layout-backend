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
  2. Run: python3 widget-migration-script.py export
  3. Copy /tmp/widget_migration.sql out, then into widget-layout debug-container
  4. Run: python3 widget-migration-script.py import
"""

import json
import os
import sys

import psycopg2
import psycopg2.extras

OUTPUT_FILE = "/tmp/widget_migration.sql"


def get_connection():
    required = ["DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME"]
    missing = [v for v in required if not os.environ.get(v)]
    if missing:
        print(f"ERROR: Missing env vars: {', '.join(missing)}")
        print("Is the DB secret mounted?")
        sys.exit(1)

    return psycopg2.connect(
        host=os.environ["DB_HOST"],
        port=os.environ.get("DB_PORT", "5432"),
        user=os.environ["DB_USER"],
        password=os.environ["DB_PASSWORD"],
        dbname=os.environ["DB_NAME"],
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
    """)

    rows = cur.fetchall()
    print(f"Fetched {len(rows)} rows (with user_id resolved)")

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
    print(f"  4. Run import: python3 widget-migration-script.py import")

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
    cur.execute(sql)
    conn.commit()

    cur.execute("SELECT COUNT(*) FROM dashboard_templates")
    count = cur.fetchone()[0]
    print(f"Target DB now has {count} dashboard_templates rows")
    print("Done.")

    cur.close()
    conn.close()


if __name__ == "__main__":
    if len(sys.argv) != 2 or sys.argv[1] not in ("export", "import"):
        print("Usage: python3 widget-migration-script.py <export|import>")
        print()
        print("  export  - Run inside chrome-service debug-container")
        print("            Generates SQL file from source DB")
        print()
        print("  import  - Run inside widget-layout-backend debug-container")
        print("            Executes SQL file against target DB")
        sys.exit(1)

    if sys.argv[1] == "export":
        do_export()
    else:
        do_import()
