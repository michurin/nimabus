.bail on
.echo off
.timer on
.changes on
.timeout 5000
.mode list

begin immediate;

-- {{{1 Cleanup

.print 'Cleanup'

.print 'Cleaning calendar_dates (based on new data)'
select count(*) from calendar_dates;
delete from calendar_dates
where service_id in (
    select calendar_dates.service_id
    from
        calendar_dates left join
        calendar_dates_ on
            calendar_dates.date = calendar_dates_.date and
            calendar_dates.service_id != calendar_dates_.service_id
    where calendar_dates_.service_id is not null
);
select count(*) from calendar_dates;

.print 'Cleaning calendar_dates (keep last days only) [TODO]'

.print 'Cleaning trips'
select count(*) from trips;
delete from trips
where trip_id in (
    select trips.trip_id
    from
        trips left join
        calendar_dates on trips.service_id = calendar_dates.service_id
    where calendar_dates.service_id is null
);
select count(*) from trips;

.print 'Cleaning routes'
select count(*) from routes;
delete from routes
where route_id in (
    select distinct routes.route_id
    from
        routes left join
        trips on routes.route_id = trips.route_id
    where trips.route_id is null
);
select count(*) from routes;

.print 'Cleaning stop_times'
select count(*) from stop_times;
delete from stop_times
where stop_id in (
    select distinct stop_times.stop_id
    from
        stop_times left join
        trips on stop_times.trip_id = trips.trip_id
    where trips.trip_id is null
);
select count(*) from stop_times;

.print 'Cleaning stops'
select count(*) from stops;
delete from stops
where stop_id in (
    select stops.stop_id
    from
        stops left join
        stop_times on stops.stop_id = stop_times.stop_id
    where stop_times.stop_id is null
);
select count(*) from stops;

.print 'Create generation'

insert into system_generations (generation_id) values (null);

.print 'Upserting'

-- {{{1 Upserting: agency

.print 'Upserting: agency'
select count(*) from agency;
with g as (select max(generation_id) as id from system_generations)
replace into agency (
    agency_id,
    agency_name,
    agency_url,
    agency_timezone,
    agency_lang,
    agency_phone,
    generation_id
) select
    cast(agency_id as integer),
    agency_name,
    agency_url,
    agency_timezone,
    agency_lang,
    agency_phone,
    g.id
from agency_ inner join g;
drop table agency_;
select count(*) from agency;

-- {{{1 Upserting: calendar_dates

.print 'Upserting: calendar_dates'
select count(*) from calendar_dates;
with g as (select max(generation_id) as id from system_generations)
replace into calendar_dates (
    service_id,
    date,
    exception_type,
    generation_id
) select
    cast(service_id as integer),
    date,
    cast(exception_type as integer),
    g.id
from calendar_dates_ inner join g;
drop table calendar_dates_;
select count(*) from calendar_dates;

-- {{{1 Upserting: routes

.print 'Upserting: routes'
select count(*) from routes;
with g as (select max(generation_id) as id from system_generations)
replace into routes (
    route_id,
    agency_id,
    route_short_name,
    route_long_name,
    route_desc,
    route_type,
    route_color,
    route_text_color,
    generation_id
) select
    cast(route_id as integer),
    cast(agency_id as integer),
    route_short_name,
    route_long_name,
    route_desc,
    cast(route_type as integer),
    route_color,
    route_text_color,
    g.id
from routes_ inner join g;
drop table routes_;
select count(*) from routes;

-- {{{1 Upserting: stop_times

.print 'Upserting: stop_times'
select count(*) from stop_times;
with g as (select max(generation_id) as id from system_generations)
replace into stop_times (
    trip_id,
    arrival_time,
    departure_time,
    stop_id,
    stop_sequence,
    pickup_type,
    drop_off_type,
    generation_id
) select
    cast(trip_id as integer),
    arrival_time,
    departure_time,
    cast(stop_id as integer),
    cast(stop_sequence as integer),
    cast(pickup_type as integer),
    cast(drop_off_type as integer),
    g.id
from stop_times_ inner join g;
drop table stop_times_;
select count(*) from stop_times;

-- {{{1 Upserting: stops

.print 'Upserting: stops'
select count(*) from stops;
with g as (select max(generation_id) as id from system_generations)
replace into stops (
    stop_id,
    stop_code,
    stop_name,
    stop_desc,
    stop_lat,
    stop_lon,
    zone_id,
    generation_id
) select
    cast(stop_id as integer),
    cast(stop_code as integer),
    stop_name,
    stop_desc,
    cast(stop_lat as real),
    cast(stop_lon as real),
    cast(zone_id as integer),
    g.id
from stops_ inner join g;
drop table stops_;
select count(*) from stops;

-- {{{1 Upserting: trips

.print 'Upserting: trips'
select count(*) from trips;
with g as (select max(generation_id) as id from system_generations)
replace into trips (
    route_id,
    service_id,
    trip_id,
    trip_headsign,
    shape_id,
    direction_id,
    generation_id
) select
    cast(route_id as integer),
    cast(service_id as integer),
    cast(trip_id as integer),
    trip_headsign,
    cast(shape_id as integer),
    cast(direction_id as integer),
    g.id
from trips_ inner join g;
drop table trips_;
select count(*) from trips;

-- {{{1 Upserting: translations

.print 'Upserting: translations'
select count(*) from translations;
with g as (select max(generation_id) as id from system_generations)
replace into translations (
    table_name,
    field_name,
    language,
    translation,
    record_id,
    generation_id
) select
    table_name,
    field_name,
    language,
    translation,
    cast(record_id as integer),
    g.id
from translations_ inner join g;
drop table translations_;
select count(*) from translations;

commit;

-- {{{1 Vacuum/analyze

.print 'Vacuum/analyze'
analyze;
vacuum;

.print 'Done.'

-- vi: set expandtab ts=4 sw=4 ft=sql fdm=marker fdl=1:
