#!/bin/sh
# Backup Postgres inventaris + rotasi 30 hari.
# Cron: 0 2 * * * /path/to/backup.sh
set -e

BACKUP_DIR="${BACKUP_DIR:-./backups}"
PGHOST="${PGHOST:-localhost}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-postgres}"
PGDATABASE="${PGDATABASE:-inventaris}"
KEEP_DAYS="${KEEP_DAYS:-30}"

mkdir -p "$BACKUP_DIR"
STAMP=$(date +%Y%m%d_%H%M%S)
FILE="$BACKUP_DIR/${PGDATABASE}_$STAMP.sql.gz"

PGPASSWORD="$PGPASSWORD" pg_dump -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" "$PGDATABASE" | gzip > "$FILE"
echo "backup written: $FILE"

find "$BACKUP_DIR" -name "${PGDATABASE}_*.sql.gz" -mtime +"$KEEP_DAYS" -delete
echo "rotated: kept last $KEEP_DAYS days"
