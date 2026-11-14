.mode csv

.print "Tables:"
.tables

.print "Dropping temporary tables..."
drop table if exists agency_;
drop table if exists calendar_dates_;
drop table if exists routes_;
drop table if exists stop_times_;
drop table if exists stops_;
drop table if exists translations_;
drop table if exists trips_;

.print "Loading..."
.import agency.txt agency_
.import calendar_dates.txt calendar_dates_
.import routes.txt routes_
.import stops.txt stops_
.import stop_times.txt stop_times_
.import trips.txt trips_
.import translations.txt translations_

.print "Fin."

.print "Tables:"
.tables

-- vi: set expandtab ts=4 sw=4 ft=sql:
