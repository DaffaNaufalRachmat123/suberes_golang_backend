#!/usr/bin/env bash
# deploy/backup.sh — PostgreSQL backup script
# ─────────────────────────────────────────────────────────────────────────────
# Creates a compressed, timestamped dump inside ./backup/ and removes dumps
# older than RETENTION_DAYS (default: 30).
#
# Usage (runs inside a cron job on the VPS host):
#   0 3 * * * /opt/suberes/deploy/backup.sh >> /var/log/suberes_backup.log 2>&1
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/opt/suberes/backup}"
CONTAINER="${POSTGRES_CONTAINER:-suberes_postgres}"
DB_NAME="${PROD_DATABASE:-suberes}"
DB_USER="${PROD_USERNAME:-postgres}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/db_${DB_NAME}_${TIMESTAMP}.sql.gz"

mkdir -p "$BACKUP_DIR"

echo "[$(date -u +"%Y-%m-%dT%H:%M:%SZ")] Starting backup → ${BACKUP_FILE}"

# Dump via pg_dump running inside the postgres container, then gzip on-the-fly
docker exec "$CONTAINER" \
    pg_dump -U "$DB_USER" "$DB_NAME" \
    | gzip -9 > "$BACKUP_FILE"

echo "[$(date -u +"%Y-%m-%dT%H:%M:%SZ")] Backup complete. Size: $(du -sh "$BACKUP_FILE" | cut -f1)"

# Remove old backups
find "$BACKUP_DIR" -name "db_${DB_NAME}_*.sql.gz" -mtime +"$RETENTION_DAYS" -delete
echo "[$(date -u +"%Y-%m-%dT%H:%M:%SZ")] Pruned backups older than ${RETENTION_DAYS} days"
