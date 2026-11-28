#!/bin/sh

set -ex

cd "${0%/*}"

DATA_DIR="${DATA_DIR-"../data"}"
DB_DIR="${DB_DIR-"../db"}"
DB_FILE="${DB_FILE-"routes.db"}"
ZIP_FILE="emel_limassol_gtfs_$(date '+%Y%m%d%H%M%S').zip"
GTFS_URL='https://www.motionbuscard.org.cy/opendata/downloadfile?file=GTFS%5C6_google_transit.zip&rel=True'

mkdir -p "$DATA_DIR"
mkdir -p "$DB_DIR"

touch "$DATA_DIR/headers"
curl -qsI "$GTFS_URL" | env -i grep -i '^Content' | env -i sort >"$DATA_DIR/headers_"

if cmp --silent "$DATA_DIR/headers" "$DATA_DIR/headers_"
then
    echo "Files are equal"
    rm "$DATA_DIR/headers_"
    exit
fi

curl -qs -L -A 'Mozilla/5.0' -o "$DATA_DIR/$ZIP_FILE" "$GTFS_URL"

ZIP_SHA="$(sha256sum -b "$DATA_DIR/$ZIP_FILE" | sed 's/ .*//')"
LAST_GEN="$(sqlite3 "$DB_DIR/$DB_FILE" '.mode list' 'select max(generation_id) from system_generations')"
sqlite3 "$DB_DIR/$DB_FILE" "insert or ignore into system_generations (checksum) values ('$ZIP_SHA')"
GEN="$(sqlite3 "$DB_DIR/$DB_FILE" '.mode list' 'select max(generation_id) from system_generations')"

echo "LAST_GEN=$LAST_GEN"
echo "     GEN=$GEN"

if test "x$LAST_GEN" = "x$GEN"
then
    :
else

    unzip -o -j "$DATA_DIR/$ZIP_FILE" -d "$DATA_DIR"

    # sqlite3 "$DB_DIR/$DB_FILE" < sql/99_drop.sql
    # sqlite3 "$DB_DIR/$DB_FILE" < sql/00_init.sql
    sqlite3 "$DB_DIR/$DB_FILE" -cmd ".cd $DATA_DIR" < sql/10_load.sql
    sqlite3 "$DB_DIR/$DB_FILE" < sql/20_merge.sql
    sqlite3 "$DB_DIR/$DB_FILE" < sql/30_derived.sql
fi

cp "$DATA_DIR/headers_" "$DATA_DIR/headers"
rm "$DATA_DIR/headers_"
