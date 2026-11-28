.bail on
.echo off
.timer on
.changes on
.timeout 5000
.mode list

begin immediate;

-- {{{1 Cleanup

.print 'Cleanup (naive truncate)'

delete from agency;
delete from calendar_dates;
delete from routes;
delete from stop_times;
delete from stops;
delete from translations;
delete from trips;

.print 'Upserting'

-- {{{1 Upserting: agency

.print 'Upserting: agency'
select count(*) from agency;
replace into agency (
    agency_id,
    agency_name,
    agency_url,
    agency_timezone,
    agency_lang,
    agency_phone
) select
    cast(agency_id as integer),
    agency_name,
    agency_url,
    agency_timezone,
    agency_lang,
    agency_phone
from agency_;
drop table agency_;
select count(*) from agency;

-- {{{1 Upserting: calendar_dates

.print 'Upserting: calendar_dates'
select count(*) from calendar_dates;
replace into calendar_dates (
    service_id,
    date,
    exception_type
) select
    cast(service_id as integer),
    date,
    cast(exception_type as integer)
from calendar_dates_;
drop table calendar_dates_;
select count(*) from calendar_dates;

-- {{{1 Upserting: routes

.print 'Upserting: routes'
select count(*) from routes;
replace into routes (
    route_id,
    agency_id,
    route_short_name,
    route_long_name,
    route_desc,
    route_type,
    route_color,
    route_text_color
) select
    cast(route_id as integer),
    cast(agency_id as integer),
    route_short_name,
    route_long_name,
    route_desc,
    cast(route_type as integer),
    route_color,
    route_text_color
from routes_;
drop table routes_;
select count(*) from routes;

-- {{{1 Upserting: stop_times

.print 'Upserting: stop_times'
select count(*) from stop_times;
replace into stop_times (
    trip_id,
    arrival_time,
    departure_time,
    stop_id,
    stop_sequence,
    pickup_type,
    drop_off_type
) select
    cast(trip_id as integer),
    arrival_time,
    departure_time,
    cast(stop_id as integer),
    cast(stop_sequence as integer),
    cast(pickup_type as integer),
    cast(drop_off_type as integer)
from stop_times_;
drop table stop_times_;
select count(*) from stop_times;

-- {{{1 Upserting: stops

.print 'Upserting: stops'
select count(*) from stops;
replace into stops (
    stop_id,
    stop_code,
    stop_name,
    stop_desc,
    stop_lat,
    stop_lon,
    zone_id
) select
    cast(stop_id as integer),
    cast(stop_code as integer),
    stop_name,
    stop_desc,
    cast(stop_lat as real),
    cast(stop_lon as real),
    cast(zone_id as integer)
from stops_;
drop table stops_;
select count(*) from stops;

-- {{{1 Upserting: trips

.print 'Upserting: trips'
select count(*) from trips;
replace into trips (
    route_id,
    service_id,
    trip_id,
    trip_headsign,
    shape_id,
    direction_id
) select
    cast(route_id as integer),
    cast(service_id as integer),
    cast(trip_id as integer),
    trip_headsign,
    cast(shape_id as integer),
    cast(direction_id as integer)
from trips_;
drop table trips_;
select count(*) from trips;

-- {{{1 Upserting: translations

.print 'Upserting: translations'
select count(*) from translations;
replace into translations (
    table_name,
    field_name,
    language,
    translation,
    record_id
) select
    table_name,
    field_name,
    language,
    translation,
    cast(record_id as integer)
from translations_;
drop table translations_;
select count(*) from translations;

commit;

-- {{{1 Vacuum/analyze

.print 'Vacuum/analyze'
analyze;
vacuum;

.print 'Done.'

-- vi: set expandtab ts=4 sw=4 ft=sql fdm=marker fdl=1:
