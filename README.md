# NimaBus — ΝήμαBus

Web UI for viewing public bus data in Limassol (or any other GTFS-compatible dataset).

> This project started almost by accident.
>
> At my day job, where I earn my living, I use AI extensively in combination with approaches like BMAD.
> It's a very powerful tool, but until then, I had only used it on a huge, existing codebase.
>
> I really wanted to try this approach to build a project from scratch.
> And not just a simple one: I wanted something that involved integrations with external services,
> a fairly sophisticated UI, a frontend, a backend, database, scheduled jobs, and so on.
>
> That's when I came up with the idea of building a website/app for viewing bus timetables and planning routes.
>
> It was a fantastic experience! I discovered that theory alone isn't enough, and I had to learn quite a few lessons the hard way.
> But, as you can see, I got it done. And at a very low cost, too!

## Quick start

```sh
cd scripts
./download.sh
```

```sh
cd backend
export NIMABUS_BIND=':8112'
export NIMABUS_DSN='../db/routes.db?mode=ro'
go run .
```

Open <http://localhost:8112> in your browser.

## Known issues

### Beware

- `download.sh` **replaces** existing GTFS data. Additionally, it's currently focused on Limassol only.

### Fundamental limitations

- **HTML over the wire.** This approach makes certain UI features difficult to implement, such as lists of recently viewed stops. Switching to a JSON API with client-side rendering would address this.
- **Incremental UI updates.** These can lead to inconsistencies when the underlying data changes.

### Nice to have

- List of recently viewed stops
- ~List of nearby stops~ `DONE`
- ~Show routes for each stop~ `DONE`
- Limit the date range
- ~Use a shared date list~ `DONE`
- ~Search~ `DONE`
- Localization and multilingual support
- Display errors
- Handle edge cases (e.g. today being outside the available date range)
- UI has to be more mobile friendly

### Integration with online maps

Services like this are usually closely integrated with maps.
However, I haven't seen a UI that I find particularly convenient.
At the same time, schematic diagrams of routes and transfers can be very useful.

I'd prioritize schematic diagrams and take that approach first.
That's why I'm in no hurry to integrate maps — I don't want to let them cloud my vision of what the UI should look like.

### Updates

- The `updated_at` field is not a reliable indicator. We need a more robust way to track data versions or upload generations.

## Limitations to keep in mind

Some bus operators seem to maintain denormalized data on their side.
They generate new foreign keys with every data export, which can make updates difficult.
This is why I've separated the *upload* and *merge* phases in `download.sh`.
The *merge* phase may not be implemented carefully enough yet.

### Development

- It nice to have tests
- Write proper routing on client side using true history API

### Instructions

- Add `systemd` unit files: service, updater unit and updater timer

## Links

- Docs: <https://gtfs.org/documentation/schedule/reference/>
- Data: <https://motionbuscard.org.cy/opendata> (Cyprus)
- Demo (on my personal home Linux box):
    - <http://nimabus.duckdns.org/>
    - <https://nimabus.duckdns.org:8443/>
