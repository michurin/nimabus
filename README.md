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

## Installation using systemd

Create dedicated user:

```sh
sudo groupadd nima
sudo useradd -g nima -d /nohome -M -l -s /usr/sbin/nologin nima
```

Systemd unit example:

```ini
# /etc/systemd/system/nimabus.service
[Unit]
Description=Nimabus service
Documentation=http://nimabus.duckdns.org/
After=network-online.target nss-lookup.target
Wants=network-online.target nss-lookup.target

[Service]
Type=simple
Restart=always
RestartSec=600
TimeoutStartSec=600
User=nima
Group=nima
Environment=NIMABUS_BIND=127.0.0.1:7567
Environment=NIMABUS_DSN=/var/lib/nimabus/var/nimabus.db?mode=ro
ExecStart=/opt/nimabus/bin/server
StateDirectory=nimabus/var
# StandardError=null
# StandardOutput=null

[Install]
WantedBy=multi-user.target
```

Start unit:

```sh
sudo systemctl daemon-reload
sudo systemctl start nimabus
sudo systemctl status nimabus
sudo systemctl enable nimabus
```

Initialize database:

```
sudo -u nima sh -c 'sqlite3 /var/lib/nimabus/var/nimabus.db <sql/00_init.sql'
```

One manual run of downloading (just to check everything is OK):

```
sudo -u nima sh -c 'DATA_DIR=/var/lib/nimabus/tmp DB_DIR=/var/lib/nimabus/var DB_FILE=nimabus.db /opt/nimabus/bin/download.sh'
```

Systemd timer file:

```ini
# /etc/systemd/system/nimabus-update.timer
[Unit]
Description=Nimabus: update every 30 minutes

[Timer]
OnCalendar=*:0/30
Persistent=true

[Install]
WantedBy=timers.target
```

And updater itself:

```ini
# /etc/systemd/system/nimabus-update.service
[Unit]
Description=cnbot: update message

[Service]
Type=oneshot
User=nima
Group=nima
Environment=DATA_DIR=/var/lib/nimabus/tmp
Environment=DB_DIR=/var/lib/nimabus/var
Environment=DB_FILE=nimabus.db
ExecStart=/opt/nimabus/bin/download.sh
StateDirectory=nimabus/tmp
StateDirectory=nimabus/var
```

I'm showing my setup and my paths, of cause you are free to tune them as you like.

## Known issues and ideas

### Grand Redevelopment required

At the moment, the backend data structure almost entirely mirrors the GTFS structure.

This causes two problems:

- Queries become unnecessarily complicated.
- Merging data from different bus operators — or even data from the same operator downloaded on different days — becomes risky and unpredictable due to inconsistent foreign keys.

There is really only one solution: design and implement data model that can support multiple updates from different sources.

I think this can wait until we have some real-world experience with the site and a clearer understanding of all the requirements.

### Integration with online maps

Services like this are usually closely integrated with maps.
However, I haven't seen a UI that I find particularly convenient.
At the same time, schematic diagrams of routes and transfers can be very useful.

Another problem can be seen right on the transport layer in Google Maps:

A single physical bus stop can be represented as multiple separate stops located very close to each other.

This happens because different data sources (for different routes) maintain their own stop lists,
with their own GPS coordinates and sometimes even different names, despite referring to the same physical bus stop.

This looks really awkward on a map.
However, on schematic diagrams, these can simply be represented as regular stop-to-stop or route-to-route connections.

I'd prioritize schematic diagrams and take that approach first.
That's why I'm in no hurry to integrate maps — I don't want to let them cloud my vision of what the UI should look like.

## Limitations to keep in mind

Some bus operators seem to maintain denormalized data on their side.
They generate new foreign keys with every data export, which can make updates difficult.
This is why I've separated the *upload* and *merge* phases in `download.sh`.
The *merge* phase may not be implemented carefully enough yet.

### Questions to answer

If we decide to support more operators (not just ΕΜΕΛ), we'll need to reconsider a few UI elements:

- **The home page:** We won't be able to display all stops as we do now, as the list would become too long. What should we show instead?
- **Search:** By route and operator? Or across all operators, with labels to indicate which operator each result belongs to?
- **Provider filtering:** Do we need some kind of provider filter? If so, where should it appear — in the nearby stops list, for example?

### Nice to have

- List of recently viewed stops
- Sort of bookmarks
- Handle edge cases (e.g. today being outside the available date range)

### Development

- It nice to have tests

## Links

- Docs: <https://gtfs.org/documentation/schedule/reference/>
- Data: <https://motionbuscard.org.cy/opendata> (Cyprus)
- Demo (on my personal home Linux box):
    - <http://nimabus.duckdns.org/>
    - <https://nimabus.duckdns.org/>
