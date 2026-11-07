package main

import (
	"bytes"
	"database/sql"
	_ "embed"
	"fmt"
	"html"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed htdocs/index.html
var page string

//go:embed htdocs/script.js
var script string

//go:embed htdocs/style.css
var style string

//go:embed htdocs/favicon.ico
var favicon []byte

const yymd = "20060102"

func writeStops(db *sql.DB, w io.Writer) error {
	fmt.Fprint(w, `<h2>stops</h2>`)
	rows, err := db.Query(`
select
    stop_id,
    stop_name,
    stop_lat,
    stop_lon
from stops
order by stop_name, stop_id
`)
	if err != nil {
		return fmt.Errorf("stops list: db: %w", err)
	}
	defer rows.Close()
	fmt.Fprintln(w, `<ul>`)
	for rows.Next() {
		var id int
		var lat, lon float64
		var name string
		rows.Scan(&id, &name, &lat, &lon)
		fmt.Fprintf(w, `<li><a id="stop_%[1]d"></a><button onclick="open_stop(%[1]d, true)">🚍 %[2]s</button> %[3]s<ul class="stop_body" id="stop_%[1]d_body"></ul></li>`, id, html.EscapeString(name), mapUrl(lat, lon))
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("stops list fetching: db: %w", err)
	}
	fmt.Fprintln(w, `</ul>`)
	return nil
}

func writeSchedule(db *sql.DB, w io.Writer, stopIDDate string) error {
	stopIDStr, targetDate, ok := strings.Cut(stopIDDate, "/")
	if !ok {
		return fmt.Errorf("schedule: invalid URL: %q", stopIDDate)
	}
	stopID, err := strconv.Atoi(stopIDStr)
	if err != nil {
		return fmt.Errorf("schedule: stop id: %w", err)
	}
	var dateA, dateB string
	err = db.QueryRow(`select min(calendar_dates.date), max(calendar_dates.date) from calendar_dates`).Scan(&dateA, &dateB) // TODO for stopID
	if err != nil {
		return fmt.Errorf("schedule: date range: db: %w", err)
	}

	if targetDate == "today" {
		targetDate = time.Now().Format(yymd)
	}

	fmt.Fprint(w, `<li>Dates<ul>`)
	dateR, err := dateRange(dateA, dateB)
	if err != nil {
		return fmt.Errorf("schedule: date range: %w", err)
	}
	for _, v := range dateR {
		p := ""
		if v == targetDate {
			p = ` style="filter: invert(1) hue-rotate(180deg);"`
		}
		wdt, err := time.Parse(yymd, v)
		if err != nil {
			return fmt.Errorf("schedule: time error: %w", err)
		}
		wd := wdt.Weekday()
		fmt.Fprintf(w, `<li><button onclick="select_date(%[3]d, '%[5]s')"%[2]s><tt>📅 %[1]s %[4]s</tt></button></li>`, formatDate(v), p, stopID, wd.String()[:3], v)
	}
	fmt.Fprint(w, `</ul></li>`)

	var stopName string
	var stopLat float64
	var stopLon float64
	err = db.QueryRow(`
select
    stop_name,
    stop_lat,
    stop_lon
from stops
where stop_id = ?
`, stopID).Scan(&stopName, &stopLat, &stopLon)
	if err != nil {
		return fmt.Errorf("schedule: stop info: db: %w", err)
	}

	rows, err := db.Query(`
select
    trips.trip_id,
    routes.route_short_name,
    trips.direction_id,
    stop_times.arrival_time,
    routes.route_color,
    routes.route_text_color
from
    stop_times inner join
    trips on stop_times.trip_id = trips.trip_id inner join
    routes on trips.route_id = routes.route_id inner join
    calendar_dates on trips.service_id = calendar_dates.service_id
where
    stop_times.stop_id = ? and
    calendar_dates.date = ?
order by stop_times.arrival_time
`, stopID, targetDate)
	if err != nil {
		return fmt.Errorf("schedule: trip: db: %w", err)
	}
	defer rows.Close()
	fmt.Fprintf(w, `<li>Trips: <b>🚍 %s</b> %s<ul>`, html.EscapeString(stopName), mapUrl(stopLat, stopLon))
	for rows.Next() {
		var id int
		var shortName string
		var direction int
		var aTime string
		var color string
		var bgColor string
		rows.Scan(&id, &shortName, &direction, &aTime, &color, &bgColor)
		fmt.Fprintf(w, `<li><button onclick="open_trip(%[5]d)"><tt>%[4]s</tt> <span style="color: #%[1]s; background-color: #%[2]s">%[3]s %[6]s</span></button><ul class="trip_body" id="trip_%[5]d_body"></ul></li>`, color, bgColor, html.EscapeString(shortName), removeSeconds(aTime), id, directionSign(direction))
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("schedule: trip fetching: db: %w", err)
	}
	fmt.Fprintln(w, `</ul></li>`)
	return nil
}

func writeTrip(db *sql.DB, w io.Writer, tripIDStr string) error {
	tripID, err := strconv.Atoi(tripIDStr)
	if err != nil {
		return fmt.Errorf("trip: invalid id: %w", err)
	}

	rows, err := db.Query(`
select
    stop_times.stop_id,
    stop_times.arrival_time,
    stops.stop_name,
    stops.stop_lat,
    stops.stop_lon
from
    stop_times inner join
    stops on stops.stop_id = stop_times.stop_id
where trip_id = ?
order by stop_sequence
`, tripID)
	if err != nil {
		return fmt.Errorf("trip: db: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var stopID int
		var aTime string
		var stopName string
		var lat, lon float64
		rows.Scan(&stopID, &aTime, &stopName, &lat, &lon)
		fmt.Fprintf(w, `<li><button onclick="open_stop(%[3]d, false)">🚍 <tt>%[2]s</tt> %[1]s</button> %[4]s</li>`, html.EscapeString(stopName), removeSeconds(aTime), stopID, mapUrl(lat, lon))
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("trip: db fetching: %w", err)
	}
	return nil
}

func dateRange(dateA, dateB string) ([]string, error) {
	a, err := time.Parse(yymd, dateA)
	if err != nil {
		return nil, fmt.Errorf("data range: bound a: %w", err)
	}
	b, err := time.Parse(yymd, dateB)
	if err != nil {
		return nil, fmt.Errorf("data range: bound b: %w", err)
	}
	res := []string(nil)
	for i := b.Add(12 * time.Hour); a.Before(i); i = i.Add(-24 * time.Hour) { // TODO limit
		res = append(res, i.Format(yymd))
	}
	slices.Reverse(res)
	return res, nil
}

func mapUrl(lat, lon float64) string {
	return fmt.Sprintf(`<a href="https://google.com/maps?q=%s,%s" target="_blank">📍</a>`, url.QueryEscape(fmt.Sprintf("%g", lat)), url.QueryEscape(fmt.Sprintf("%g", lon)))
}

func removeSeconds(a string) string {
	return a[0:max(0, len(a)-3)]
}

func formatDate(a string) string {
	return a[4:6] + "/" + a[6:8]
}

func directionSign(a int) string {
	if a == 0 {
		return "◀"
	}
	return "▶"
}

func main() {
	addr := os.Args[1]
	dbfile := os.Args[2]
	fmt.Println("Opening database", dbfile)
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		fmt.Println("GET", p)
		var err error
		switch {
		case p == "/stops":
			err = writeStops(db, w)
		case strings.HasPrefix(p, "/schedule/"):
			err = writeSchedule(db, w, p[10:])
		case strings.HasPrefix(p, "/trip/"):
			err = writeTrip(db, w, p[6:])
		case p == "/script.js":
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			_, err = fmt.Fprint(w, script)
		case p == "/style.css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			_, err = fmt.Fprint(w, style)
		case p == "/favicon.ico":
			w.Header().Set("Content-Type", "image/x-icon")
			_, err = io.Copy(w, bytes.NewReader(favicon))
		case p == "/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, err = fmt.Fprint(w, page)
		default:
			http.NotFound(w, r)
		}
		if err != nil {
			log.Println(err)
		}
	})

	fmt.Println("Starting at", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		slog.Error(err.Error())
	}
}
