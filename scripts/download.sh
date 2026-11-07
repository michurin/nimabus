#!/bin/sh

set -ex

DATA_DIR=../data
DB_DIR=../db
DB_FILE=routes.db
ZIP_FILE="emel_limassol_gtfs_$(date +%Y%m%d).zip"
GTFS_URL="https://www.motionbuscard.org.cy/opendata/downloadfile?file=GTFS%5C6_google_transit.zip&rel=True"

mkdir -p "$DATA_DIR"
mkdir -p "$DB_DIR"

cp *.sqlight "$DATA_DIR" # just prepare working directory in naive way

pushd "$DATA_DIR"

curl -qs -L -A "Mozilla/5.0" -o "$ZIP_FILE" "$GTFS_URL"
unzip -o "$ZIP_FILE"

# naive drop->init->upload->merge. It has to be upload->merge
sqlite3 "$DB_FILE" <sql_drop.sqlight
sqlite3 "$DB_FILE" <sql_init.sqlight
sqlite3 "$DB_FILE" <sql_upload.sqlight
sqlite3 "$DB_FILE" <sql_merge.sqlight

popd

mv "$DATA_DIR/$DB_FILE" "$DB_DIR/$DB_FILE"

echo 'FIN.'