/*eslint indent: ["error", 2]*/
/*eslint eqeqeq: ["error", "always"]*/
/*eslint prefer-const: "error"*/
/*eslint no-var: "error"*/
/*eslint one-var: ["error", "never"]*/
/*eslint semi: ["error", "never"]*/
/*eslint quotes: ["error", "single"]*/
/*eslint prefer-arrow-callback: "error"*/
/*eslint arrow-body-style: "error"*/

import { generation, dates, current_date as current_date_, stops_info } from './vars.js'
import { StringNgramIndex } from './ngramindex.js'

// constants

const titleEl = document.getElementById('title')
const datesEl = document.getElementById('dates')
const staticEl = document.getElementById('static-body')
const dynamicEl = document.getElementById('dynamic-body')
const loaderEl = document.getElementById('loader')
const errorEl = document.getElementById('error')

const baseTitle = 'NimaBus — ΝήμαBus'

const index = new StringNgramIndex({
  ngramLen: 3, normalizer: text => [
    text.toLowerCase().normalize('NFD').replace(/\p{M}/gu, '').replace(/[^\p{L}\p{N}]+/gu, ' ').trim()
  ]
})

function order_stops(ids) {
  ids.sort((a, b) => {
    const parse = s => {
      const m = s.match(/^(\d*)\s*(.*)$/)
      return { n: Number(m[1] || 0), t: m[2] }
    }
    const x = parse(stops_info[a].name)
    const y = parse(stops_info[b].name)
    return x.t.localeCompare(y.t, undefined, {
      sensitivity: 'base',
      numeric: true
    }) || x.n - y.n
  })
  return ids
}

const ordered_stops = order_stops(Object.keys(stops_info))

const index_by_stop = {}

Object.values(stops_info).forEach((e) => e.routes.forEach((p) => {
  if (!index_by_stop[p]) {
    index_by_stop[p] = []
  }
  index_by_stop[p].push(e.id)
}))
Object.keys(index_by_stop).forEach((k) => { index_by_stop[k] = order_stops(index_by_stop[k]) })

const cache = {}

// state

let current_date = current_date_ // global app scope
let scope = 'root'
let current_stop = 0 // on scope != root
let filtered = undefined // on scope == root

// helpers

function esc(x) {
  return x.replace(/[&<>"']/g, c => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    '\'': '&#39;'
  })[c])
}

async function fetchJson(url) {
  const v = cache[url]
  if (v) {
    return v
  }
  try {
    const r = await fetch(url)
    if (!r.ok) {
      throw new Error(`Response is not ok: ${r.status}`)
    }
    const x = await r.json()
    cache[url] = x
    return x
  } catch {
    return false
  }
}

function show_loader() {
  loaderEl.style.display = 'block'
  requestAnimationFrame(() => { requestAnimationFrame(() => { loaderEl.style.opacity = '1' }) })
}

function hide_loader() {
  loaderEl.style.opacity = '0'
  loaderEl.style.display = 'none'
}

// renders

function stop_line(pfx, id, m, t, active) {
  const v = stops_info[id]
  const lab = `${pfx}_stop_${id}`
  return `<li${active ? ' class="current_stop"' : ''}>` +
    (m ? '' : `<button onclick="act_walk(event, '${lab}_walks', ${id}, '${lab}_buses')">🚶</button> `) +
    `<button onclick="act_timetable(event, '${lab}_buses', ${id}, '${lab}_walks')">🚍</button> ` +
    (t ? `<tt>🚍 ${t}</tt> ` : '') +
    (m ? `<tt>🚶 ${m}m</tt> &rarr; ` : '') +
    `<a onclick="act_stop(event, ${id})">` +
    esc(v.name) +
    '</a>' +
    `<a target="_blank" href="https://maps.google.com/maps?q=loc:${v.lat},${v.lon}">📍</a><br>` +
    v.routes.map((x) => `<span class="r${esc(x)}">${esc(x)}</span>`).join('\n') +
    `<ul id="${lab}_walks"></ul><ul id="${lab}_buses"></ul></li>`
}

function time_line(pfx, id, t, r, d, stop_id) { // stop_id just for tracking
  const lab = `${pfx}_trip_${id}`
  return '<li>' +
    `<button onclick="act_trip(event, '${lab}', ${id}, ${stop_id})"><tt>${esc(t)}</tt> ` +
    `<span class="r${esc(r)}">${esc(r)} ${d === 0 ? '◀' : '▶'}</span></button>` +
    `<ul id="${lab}"></ul></li>`
}

async function time_lines_all(l, id) {
  const sh = await fetchJson(`/api/timetable/${current_date}/${id}.json`)
  if (sh) {
    return sh.map((x) => time_line(l, x.id, x.time, x.route, x.dir, id)).join('\n')
  }
  show_error()
  return ''
}

function stop_lines_all(a) { // it's NOT universal, useful in lists that allows filtering
  if (a && a.length > 0) {
    filtered = true
  } else {
    if (filtered === false) { // do not consider undefined
      return // do nothing, list is already full
    }
    a = ordered_stops
    filtered = false
  }
  dynamicEl.innerHTML = `<ul>${a.map((x) => stop_line('root', x)).join('\n')}</li>`
}

// render pages (in fact, parts of page)

function root_page() {
  scope = 'root'
  document.title = baseTitle + ' — Limassol buses'
  titleEl.innerText = 'All Limassol stops'
  staticEl.innerHTML = '<input type="text" oninput="act_search(event)" placeholder="Search..."></input>'
  stop_lines_all()
  window.scrollTo(0, 0)
}

async function stop_page_dynamic(id) {
  show_loader()
  dynamicEl.innerHTML = '<h2>Timetable</h2><ul>' + (await time_lines_all('bus', id)) + '</ul>'
  hide_loader()
}

async function stop_page(id) {
  scope = 'stop'
  current_stop = id
  const si = stops_info[id]
  document.title = esc(si.name) + (si.alt_name ? ' — ' + esc(si.alt_name) : '') + ' — ' + baseTitle
  titleEl.innerHTML = esc(si.name) + (si.alt_name ? '<br>' + esc(si.alt_name) : '')
  staticEl.innerHTML = '<h2>' +
    '<a onclick="act_root(event)">◀ Back to all stops</a> ⸽ ' +
    '<a onclick="toggle_hidding(event, \'stop-page-walking\')">Nearby ▾</a>' +
    '</h2>' +
    '<ul id="stop-page-walking" style="display:none">' +
    si.distances.map((x) => stop_line('walk', x.id, x.m)).join('\n') + '</ul>'
  await stop_page_dynamic(id)
  window.scrollTo(0, 0)
}

function date_page() {
  datesEl.innerHTML = dates.map((x) => `<button onclick="act_date(event, '${esc(x.id)}')""${x.id === current_date ? ' style="filter: invert(1) hue-rotate(180deg);"' : ''}><tt>${esc(x.date).replaceAll(' ', '<br>')}</tt></button>`).join('\n')
}

// events

window.addEventListener('popstate', (ev) => {
  const st = ev.state
  date_page()
  if (st?.stop) {
    stop_page(st.stop)
  } else {
    root_page()
  }
});

// init

(() => {
  date_page()
  const id = +location.pathname.match('^/stop/([0-9]+)$')?.[1]
  if (stops_info[id]) {
    stop_page(id)
  } else {
    root_page()
  }

  // index
  Object.values(stops_info).forEach((v) => { index.add(v.id, v.name, v.alt_name) })

  // localization
  if (navigator.language.toLowerCase().startsWith('el')) {
    document.getElementById('theme-toggle-label').textContent = 'Αλλαγή θέματος'
    document.getElementById('theme-toggle').setAttribute('aria-label', 'Εναλλαγή μεταξύ φωτεινού και σκοτεινού θέματος')
  }
})()

// actions

function act_stop(ev, id) {
  ev.stopPropagation()
  ev.preventDefault()
  history.pushState({ stop: id }, '', `/stop/${id}`)
  stop_page(id)
}

function act_date(ev, id) {
  ev.stopPropagation()
  ev.preventDefault()
  current_date = id
  date_page()
  if (scope === 'root') {
    root_page()
  } else {
    stop_page_dynamic(current_stop)
  }
}

function act_root(ev) {
  ev.stopPropagation()
  ev.preventDefault()
  history.pushState({}, '', '/')
  root_page()
}

async function act_timetable(ev, pfx, id, clean) {
  ev.stopPropagation()
  ev.preventDefault()
  document.getElementById(`${clean}`).innerHTML = ''
  const e = document.getElementById(`${pfx}`)
  if (e.childElementCount === 0) {
    show_loader()
    e.innerHTML = await time_lines_all(`${pfx}`, id)
    hide_loader()
  } else {
    e.innerHTML = ''
  }
}

async function act_trip(ev, pfx, id, stop_id) {
  ev.stopPropagation()
  ev.preventDefault()
  const e = document.getElementById(`${pfx}`)
  if (e.childElementCount === 0) {
    show_loader()
    const sh = await fetchJson(`/api/trip/${id}.json`)
    if (sh) {
      e.innerHTML = sh.map((x) => stop_line(pfx, x.id, false, x.t, x.id === stop_id)).join('\n')
    } else {
      show_error()
    }
    hide_loader()
  } else {
    e.innerHTML = ''
  }
}

function act_walk(ev, pfx, id, clean) {
  ev.stopPropagation()
  ev.preventDefault()
  document.getElementById(`${clean}`).innerHTML = ''
  const e = document.getElementById(`${pfx}`)
  const l = `${pfx}`
  if (e.childElementCount === 0) {
    e.innerHTML = (stops_info[id]?.distances || []).map((x) => stop_line(l, x.id, x.m)).join('\n')
  } else {
    e.innerHTML = ''
  }
}

function act_search(ev) { // it works on root page only
  ev.stopPropagation()
  ev.preventDefault()
  const v = ev.target.value.trim()
  if (index_by_stop[v]) {
    stop_lines_all(index_by_stop[v])
  } else {
    stop_lines_all(index.search(v))
  }
}

// errors

function show_error() {
  errorEl.style.display = 'block'
}

// universal utilities

function toggle_hidding(ev, id) {
  ev.stopPropagation()
  ev.preventDefault()
  const x = document.getElementById(id)
  if (x) {
    x.style.display = x.style.display === 'none' ? '' : 'none'
  }
}

// hack, however i'm considering it useful as sort of declaration of public actions

window.act_search = act_search
window.act_walk = act_walk
window.act_timetable = act_timetable
window.act_date = act_date
window.act_stop = act_stop
window.act_trip = act_trip
window.act_root = act_root

window.toggle_hidding = toggle_hidding
