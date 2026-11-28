-- CAUTION: we consider *all* dates
-- CAUTION: it have to be in same transuction

drop table if exists derived_stops;
create table if not exists derived_stops (
    stop_id integer primary key,
    route_short_names text,
    stop_name text,
    stop_lat real,
    stop_lon real,
    stop_alt_name text,
    distances text
);

with
    distances_all as (
        select
            a.stop_id as stop_id,
            b.stop_id as related_stop_id,
            cast(111.2*1000*sqrt(pow(0.823*(a.stop_lon-b.stop_lon),2)+pow(a.stop_lat-b.stop_lat,2)) as integer) as distance
        from
            stops a join stops b
        where
            a.stop_id <> b.stop_id and
            distance < 1000
    ),
    distances as (
        select
            stop_id,
            json_group_array(
                json_object(
                    'id', related_stop_id,
                    'm', distance
                ) order by distance
            ) as distance
        from
            distances_all
        group by stop_id
    ),
    route_short_names as (
        select
            stop_times.stop_id,
            json_group_array(
                distinct routes.route_short_name
                order by cast(routes.route_short_name as integer), routes.route_short_name
            ) as route_short_name
        from
            stop_times inner join
            trips on stop_times.trip_id = trips.trip_id inner join
            routes on routes.route_id = trips.route_id
        group by stop_times.stop_id
    ),
    translations_stops as (
        select
            translations.record_id as stop_id,
            group_concat(distinct translations.translation order by translations.translation) as stop_alt_name
        from
            translations
        where
            translations.table_name = 'stops' and
            translations.field_name = 'stop_name'
        group by translations.record_id
    )
insert into derived_stops
select
    stops.stop_id,
    route_short_names.route_short_name,
    stops.stop_name,
    stops.stop_lat,
    stops.stop_lon,
    ifnull(translations_stops.stop_alt_name, ''),
    ifnull(distances.distance, '[]')
from
    stops left join
    route_short_names on stops.stop_id = route_short_names.stop_id left join
    translations_stops on stops.stop_id = translations_stops.stop_id left join
    distances on distances.stop_id = stops.stop_id;

.print 'Check counts are equal'
.mode line
select count(stop_id) as 'Original stops' from stops;
select count(stop_id) as 'Derived stops ' from derived_stops;
select count(distinct stop_id) as 'Unique stops  ' from stop_times;

-- vi: set expandtab ts=4 sw=4 ft=sql:
