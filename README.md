# NimaBus — ΝήμαBus

Web UI for viewing public bus data in Limassol (or any other GTFS-compatible dataset).

## Quick start

```sh
cd scripts
./download.sh
```

```sh
cd backend
go run . :8112 'file:../db/routes.db?mode=ro'
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
- List of nearby stops
- Show routes for each stop
- Limit the date range
- Use a shared date list
- Search
- Localization and multilingual support
- Display errors
- Handle edge cases (e.g. today being outside the available date range)
- UI has to be more mobile friendly

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
