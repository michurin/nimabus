#!/bin/sh

set -ex

DATA_DIR=../data
DB_DIR=../db
DB_FILE=routes.db
ZIP_FILE="emel_limassol_gtfs_$(date +%Y%m%d).zip"
GTFS_URL='https://www.motionbuscard.org.cy/opendata/downloadfile?file=GTFS%5C6_google_transit.zip&rel=True'

mkdir -p "$DATA_DIR"
mkdir -p "$DB_DIR"

curl -qs -L -A 'Mozilla/5.0' -o "$DATA_DIR/$ZIP_FILE" "$GTFS_URL"

unzip -o -j "$DATA_DIR/$ZIP_FILE" -d "$DATA_DIR"

sqlite3 "$DB_DIR/$DB_FILE" < sql/99_drop.sql
sqlite3 "$DB_DIR/$DB_FILE" < sql/00_init.sql
sqlite3 "$DB_DIR/$DB_FILE" -cmd ".cd $DATA_DIR" < sql/10_load.sql
sqlite3 "$DB_DIR/$DB_FILE" < sql/20_merge.sql
sqlite3 "$DB_DIR/$DB_FILE" < sql/30_derived.sql
