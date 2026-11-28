package main

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed htdocs/style.css
var documentStyle []byte

//go:embed htdocs/llms.txt
var documentLlmsTxt []byte

//go:embed htdocs/apple-touch-icon.png
//go:embed htdocs/favicon.ico
//go:embed htdocs/icon-192.png
//go:embed htdocs/icon-512.png
//go:embed htdocs/index.html
//go:embed htdocs/loader.svg
//go:embed htdocs/manifest.json
//go:embed htdocs/ngramindex.js
//go:embed htdocs/robots.txt
//go:embed htdocs/script.js
var htdocs embed.FS

const htdocsDir = "htdocs/"

var replacerMarkdownEscaper = strings.NewReplacer(
	`{`, `\{`, `}`, `\}`, `[`, `\[`, `]`, `\]`, `(`, `\(`, `)`, `\)`,
	`<`, `&lt;`, `>`, `&gt;`, `+`, `\+`, `-`, `\-`, `\`, `\\`, "`", "\\`",
	`*`, `\*`, `_`, `\_`, `#`, `\#`, `.`, `\.`, `!`, `\!`, `|`, `\|`, `~`, `\~`,
)

var replacerUnjson = strings.NewReplacer("[", "", "]", "", `"`, "", ",", ", ")

// handlers

func handlerVars(ctx context.Context, db *sql.DB, w io.Writer) error {
	fmt.Fprint(w, "export const generation = 'TODO'\n")
	fmt.Fprint(w, "export const dates = ")
	current, err := varDates(ctx, db, w)
	if err != nil {
		return Wrf("stop info: %w", err)
	}
	fmt.Fprint(w, "export const current_date = '"+current+"'\n")
	fmt.Fprint(w, "export const stops_info = ")
	err = varStops(ctx, db, w)
	if err != nil {
		return Wrf("stop info: %w", err)
	}
	return nil
}

func handlerTimetable(ctx context.Context, db *sql.DB, w io.Writer, p string) error {
	type dto struct {
		TripID         int    `json:"id"`
		Time           string `json:"time"`
		RouteShortName string `json:"route"`
		Direction      int    `json:"dir"`
	}
	p, ok := strings.CutSuffix(p, ".json")
	if !ok {
		return Wrf("invalid suffix: %q", p)
	}
	date, stopStr, ok := strings.Cut(p, "/")
	if !ok {
		return Wrf("invalid format: %q", p)
	}
	stopID, err := strconv.Atoi(stopStr)
	if err != nil {
		return Wrf("invalid stopID: %w", err)
	}
	// TODO we use arrival_time only
	rows, err := db.QueryContext(ctx, `
select
    trips.trip_id,
    routes.route_short_name,
    trips.direction_id,
    stop_times.arrival_time
from
    stop_times inner join
    trips on stop_times.trip_id = trips.trip_id inner join
    routes on trips.route_id = routes.route_id inner join
    calendar_dates on trips.service_id = calendar_dates.service_id
where
    stop_times.stop_id = ? and
    calendar_dates.date = ?
order by stop_times.arrival_time
`, stopID, date)
	if err != nil {
		return Wrf("schedule: trip: db: %w", err)
	}
	defer rows.Close()
	r := []dto(nil)
	for rows.Next() {
		var id, direction int
		var shortName, aTime string
		rows.Scan(&id, &shortName, &direction, &aTime)
		r = append(r, dto{
			TripID:         id,
			Time:           strings.TrimSuffix(aTime, ":00"),
			RouteShortName: shortName,
			Direction:      direction,
		})
	}
	if err := rows.Err(); err != nil {
		return Wrf("schedule: trip fetching: db: %w", err)
	}
	err = json.NewEncoder(w).Encode(r)
	if err != nil {
		return Wrf("json encoder: %w", err)
	}
	return nil
}

func handlerTrip(ctx context.Context, db *sql.DB, w io.Writer, p string) error {
	type dto struct {
		ID   int    `json:"id"`
		Time string `json:"t"`
	}
	p, ok := strings.CutSuffix(p, ".json")
	if !ok {
		return Wrf("invalid suffix: %q", p)
	}
	tripID, err := strconv.Atoi(p)
	if err != nil {
		return Wrf("trip: invalid id: %w", err)
	}

	rows, err := db.QueryContext(ctx, `
select
    stop_times.stop_id,
    stop_times.arrival_time
from
    stop_times
where trip_id = ?
order by stop_sequence
`, tripID)
	if err != nil {
		return Wrf("trip: db: %w", err)
	}
	defer rows.Close()
	r := []dto{}
	for rows.Next() {
		var stopID int
		var aTime string
		rows.Scan(&stopID, &aTime)
		r = append(r, dto{
			ID:   stopID,
			Time: strings.TrimSuffix(aTime, ":00"),
		})
	}
	if err := rows.Err(); err != nil {
		return Wrf("trip: db fetching: %w", err)
	}
	err = json.NewEncoder(w).Encode(r)
	if err != nil {
		return Wrf("marshal trip: %w", err)
	}
	return nil
}

func handlerStyle(ctx context.Context, db *sql.DB, w io.Writer) error {
	rows, err := db.QueryContext(ctx, `
select
    route_short_name,
    max(route_text_color),
    max(route_color)
from routes
group by route_short_name
order by route_short_name
`)
	if err != nil {
		return Wrf("enrichCSS: db: %w", err)
	}
	defer rows.Close()
	_, err = io.Copy(w, bytes.NewReader(documentStyle))
	if err != nil {
		return Wrf("copy static: %w", err)
	}
	_, err = io.Copy(w, bytes.NewReader([]byte("\n\n// ** generated **\n\n")))
	if err != nil {
		return Wrf("copy comment: %w", err)
	}
	for rows.Next() {
		var name, fg, bg string
		rows.Scan(&name, &fg, &bg)
		fmt.Fprintf(w, ".r%s {color:#%s;background-color:#%s}\n", name, fg, bg)
	}
	if err := rows.Err(); err != nil {
		return Wrf("db fetching: %w", err)
	}
	return nil
}

func handlerLlmsTxt(ctx context.Context, db *sql.DB, w io.Writer) error {
	rows, err := db.QueryContext(ctx, `
select
    stop_id,
    route_short_names,
    stop_name,
    stop_alt_name
from derived_stops
order by stop_id
`)
	if err != nil {
		return Wrf("db: %w", err)
	}
	defer rows.Close()
	buf := []byte(nil)
	for rows.Next() {
		var id int
		var routes, name, altName string
		rows.Scan(&id, &routes, &name, &altName)
		if altName != "" {
			altName = " (" + replacerMarkdownEscaper.Replace(altName) + ")"
		}
		name = replacerMarkdownEscaper.Replace(name)
		routes = replacerMarkdownEscaper.Replace(replacerUnjson.Replace(routes))
		buf = fmt.Appendf(buf, "- [%[2]s%[3]s](/stop/%[1]d): routes: %[4]s\n", id, name, altName, routes)
	}
	if err := rows.Err(); err != nil {
		return Wrf("db fetching: %w", err)
	}
	_, err = io.Copy(w, bytes.NewReader(bytes.Replace(documentLlmsTxt, []byte("{{}}"), buf, 1)))
	if err != nil {
		return Wrf("write: %w", err)
	}
	return nil
}

// helpers for var handler

func varStops(ctx context.Context, db *sql.DB, w io.Writer) error {
	type dto struct {
		ID        int             `json:"id"`
		Routes    json.RawMessage `json:"routes"`
		Name      string          `json:"name"`
		Lat       float64         `json:"lat"`
		Lon       float64         `json:"lon"`
		AltName   string          `json:"alt_name"`
		Distances json.RawMessage `json:"distances"`
	}
	rows, err := db.QueryContext(ctx, `
select
    stop_id,
    route_short_names,
    stop_name,
    stop_lat,
    stop_lon,
    stop_alt_name,
    distances
from derived_stops
`)
	if err != nil {
		return Wrf("writeVars: db: %w", err)
	}
	defer rows.Close()
	data := map[string]dto{}
	for rows.Next() {
		var id int
		var routes, name, altName, distances string
		var lat, lon float64
		rows.Scan(&id, &routes, &name, &lat, &lon, &altName, &distances)
		data[strconv.Itoa(id)] = dto{
			ID:        id,
			Routes:    json.RawMessage(routes),
			Name:      name,
			Lat:       lat,
			Lon:       lon,
			AltName:   altName,
			Distances: json.RawMessage(distances),
		}
	}
	if err := rows.Err(); err != nil {
		return Wrf("writeVars: db fetching: %w", err)
	}
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		return Wrf("marshal stops_info: %w", err)
	}
	return nil
}

func varDates(ctx context.Context, db *sql.DB, w io.Writer) (string, error) {
	const yymd = "20060102"
	type dto struct {
		ID   string `json:"id"`
		Date string `json:"date"`
	}
	rows, err := db.QueryContext(ctx, `select date from calendar_dates order by date`)
	if err != nil {
		return "", Wrf("db: %w", err)
	}
	defer rows.Close()
	current := time.Now().Format(yymd)
	ok := false
	data := []dto(nil)
	for rows.Next() {
		var date string
		rows.Scan(&date)
		if date == current {
			ok = true
		}
		t, err := time.Parse(yymd, date)
		if err != nil {
			return "", Wrf("invalid date: %q: %w", date, err)
		}
		data = append(data, dto{ID: date, Date: t.Format("Jan 02 Mon")})
	}
	if err := rows.Err(); err != nil {
		return "", Wrf("db fetching: %w", err)
	}
	if len(data) == 0 {
		return "", Wrf("no data")
	}
	if !ok {
		current = data[0].ID // just force first slot
	}
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		return "", Wrf("marshal: %w", err)
	}
	return current, nil
}

// static handlers

func buildStaticHandlers() (map[string]http.Handler, error) {
	m := map[string]http.Handler{}
	now := time.Now()
	l, err := htdocs.ReadDir("htdocs")
	if err != nil {
		return nil, Wrf("static listing: %w", err)
	}
	for _, e := range l {
		filename := e.Name()
		slog.Info("Loading", slog.String("file", filename))
		faviconBytes, err := htdocs.ReadFile(htdocsDir + filename)
		if err != nil {
			return nil, Wrf("static loading: %w", err)
		}
		m["/"+filename] = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeContent(w, r, filename, now, bytes.NewReader(faviconBytes))
		})
	}
	return m, nil
}

// router

func mux(db *sql.DB, staticMap map[string]http.Handler, indexHandler http.Handler) http.Handler {
	stopRe := regexp.MustCompile(`^/stop/[0-9]{1,6}$`)
	return http.HandlerFunc(func(wo http.ResponseWriter, r *http.Request) {
		w := wrapResponseWriter(wo)
		ctx := r.Context()
		p := r.URL.Path
		ua := r.Header.Get("User-Agent")
		ip := r.Header.Get("X-Real-IP") // proxy_set_header X-Real-IP $remote_addr;
		var err error
		defer func() {
			a := []any{
				slog.Int("code", w.code),
				slog.Int("size", w.size),
				slog.String("method", r.Method),
				slog.String("path", p),
			}
			if len(ua) > 0 {
				a = append(a, slog.String("user_agent", ua))
			}
			if len(ip) > 0 {
				a = append(a, slog.String("real_ip", ip))
			}
			l := slog.LevelInfo
			if err != nil {
				a = append(a, slog.String("error", err.Error()))
				l = slog.LevelError
			}
			slog.Log(ctx, l, "Request", a...)
		}()
		if h, ok := staticMap[p]; ok {
			h.ServeHTTP(w, r)
			return
		}
		switch {
		case strings.HasPrefix(p, "/api/timetable/"):
			err = Wrc(handlerTimetable(ctx, db, w, p[15:]))
		case strings.HasPrefix(p, "/api/trip/"):
			err = Wrc(handlerTrip(ctx, db, w, p[10:]))
		case p == "/vars.js":
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			err = Wrc(handlerVars(ctx, db, w))
		case p == "/style.css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			err = Wrc(handlerStyle(ctx, db, w))
		case p == "/llms.txt":
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			err = Wrc(handlerLlmsTxt(ctx, db, w))
		case p == "/" || p == "/index.html" || stopRe.MatchString(p):
			indexHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// main

func main() {
	params, err := getParams("NIMABUS_BIND", "NIMABUS_DSN")
	if err != nil {
		slog.Error(err.Error())
		return
	}
	addr := params[0]
	dbfile := params[1]
	slog.Info("Opening database: " + dbfile)
	db, err := Wrc2(sql.Open("sqlite3", dbfile))
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer db.Close()

	staticMap, err := buildStaticHandlers()
	if err != nil {
		slog.Error(err.Error())
		return
	}
	indexHandler := staticMap["/index.html"]

	http.Handle("/", mux(db, staticMap, indexHandler))

	slog.Info("Starting at " + addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		slog.Error(err.Error())
	}
}

// env utils

func getParams(names ...string) ([]string, error) {
	r := make([]string, len(names))
	e := make([]string, 0, len(names))
	for i, n := range names {
		v, ok := os.LookupEnv(n)
		if !ok {
			e = append(e, n)
		}
		r[i] = v
		slog.Info("Using variable", slog.String("var", n), slog.String("value", v))
	}
	if len(e) > 0 {
		return nil, Wrf("no variable: %s", strings.Join(e, ", "))
	}
	return r, nil
}

// http utils

func wrapResponseWriter(h http.ResponseWriter) *responseWriter {
	return &responseWriter{
		next: h,
		code: http.StatusOK,
		size: 0,
		err:  nil,
	}
}

type responseWriter struct {
	next http.ResponseWriter
	code int
	size int
	err  error
}

func (rw *responseWriter) Header() http.Header {
	return rw.next.Header()
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	n, err := rw.next.Write(data)
	if err != nil {
		err = Wrf("middleware: %w", err)
		rw.err = err
	} else {
		rw.size += n
	}
	return n, err
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
	rw.next.WriteHeader(code)
}

// error handling utilities

func wr(err error) error {
	name := "nofunc"
	pc, _, no, ok := runtime.Caller(2)
	if ok {
		details := runtime.FuncForPC(pc)
		name = details.Name()
	}
	return fmt.Errorf("[%s:%d]: %w", name, no, err)
}

func Wr(err error) error {
	return wr(err)
}

func Wrf(format string, a ...any) error {
	return wr(fmt.Errorf(format, a...))
}

func Wrc(err error) error {
	if err != nil {
		err = wr(err)
	}
	return err
}

func Wrc2[T any](x T, err error) (T, error) {
	if err != nil {
		err = wr(err)
	}
	return x, err
}
