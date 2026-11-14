/*eslint indent: ["error", 2]*/
/*eslint eqeqeq: ["error", "always"]*/
/*eslint prefer-const: "error"*/
/*eslint no-var: "error"*/
/*eslint one-var: ["error", "never"]*/
/*eslint semi: ["error", "never"]*/
/*eslint quotes: ["error", "single"]*/
/*eslint prefer-arrow-callback: "error"*/
/*eslint arrow-body-style: "error"*/

// constants

const titleEl = document.getElementById('title')
const datesEl = document.getElementById('dates')
const rootEl = document.getElementById('root')
const loaderEl = document.getElementById('loader')

// state

let scope = 'root'
let current_stop = 0

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
  const r = await fetch(url)
  if (!r.ok) {
    throw new Error(`Response is not ok: ${r.status}`)
  }
  return await r.json()
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

function stop_line(pfx, id, m, t) {
  const v = stops_info[id]
  const lab = `${pfx}_${id}`
  return '<li>' +
    (m ? '' : `<button onclick="act_walk(event, '${lab}', ${id})">🚶</button> `) +
    `<button onclick="act_timetable(event, '${lab}', ${id})">🚍</button> ` +
    (t ? `<tt>🚍 ${t}</tt> ` : '') +
    (m ? `<tt>🚶 ${m}m</tt> &rarr; ` : '') +
    `<a onclick="act_stop(event, ${id})">` +
    esc(v.name) +
    '</a>' +
    `<a target="_blank" href="https://google.com/maps?q=${v.lat},${v.lon}">📍</a><br>` +
    v.routes.map((x) => `<span class="r${esc(x)}">${esc(x)}</span>`).join('\n') +
    `<ul id="${lab}"></ul></li>`
}

function time_line(pfx, id, t, r, d) {
  const lab = `${pfx}_${id}`
  return '<li>' +
    `<button onclick="act_trip(event, '${lab}', ${id})"><tt>${esc(t)}</tt> ` +
    `<span class="r${esc(r)}">${esc(r)} ${d === 0 ? '◀' : '▶'}</span></button>` +
    `<ul id="${lab}"></ul></li>`
}

async function time_lines_all(l, id) {
  const sh = await fetchJson(`/api/timetable/${current_date}/${id}.json`)
  return sh.map((x) => time_line(l, x.id, x.time, x.route, x.dir)).join('\n')
}

// render pages (in fact, parts of page)

function root_page() {
  scope = 'root'
  titleEl.innerText = 'All stops'
  rootEl.innerHTML = '<ul>' +
    Object.keys(stops_info).sort((a, b) => {
      const parse = s => {
        const m = s.match(/^(\d*)\s*(.*)$/)
        return { n: Number(m[1] || 0), t: m[2] }
      }
      const x = parse(stops_info[a].name)
      const y = parse(stops_info[b].name)
      return x.t.localeCompare(y.t, undefined, {
        sensitivity: 'base',
        numeric: true
      }) || x.number - y.number
    }).map((x) => stop_line('label', x)).join('\n') +
    '</ul>'
}

async function stop_page(id) {
  show_loader()
  scope = 'stop'
  current_stop = id
  const si = stops_info[id]
  titleEl.innerHTML = esc(si.name) + (si.alt_name ? '<br>' + esc(si.alt_name) : '')
  rootEl.innerHTML = '<button onclick="act_root(event)">all</button><h2>Nearby</h2><ul>' +
    si.distances.map((x) => stop_line('walk', x.id, x.m)).join('\n') +
    '</ul><h2>Timetable</h2><ul>' +
    (await time_lines_all('bus', id)) +
    '</ul>'
  hide_loader()
}

function date_page() {
  datesEl.innerHTML = dates.map((x) => `<button onclick="act_date(event, '${esc(x.id)}')""${x.id === current_date ? ' style="color:red"' : ''}><tt>${esc(x.date)}</tt></button>`).join('\n')
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
  // TODO update other parts of page
  if (scope === 'root') {
    root_page()
  } else {
    stop_page(current_stop)
  }
}

function act_root(ev) {
  ev.stopPropagation()
  ev.preventDefault()
  history.pushState({}, '', '/')
  root_page()
}

async function act_timetable(ev, pfx, id) {
  ev.stopPropagation()
  ev.preventDefault()
  const e = document.getElementById(`${pfx}`)
  if (e.childElementCount === 0) {
    show_loader()
    e.innerHTML = await time_lines_all(`${pfx}_bus`, id)
    hide_loader()
  } else {
    e.innerHTML = ''
  }
}

async function act_trip(ev, pfx, id) {
  ev.stopPropagation()
  ev.preventDefault()
  const e = document.getElementById(`${pfx}`)
  if (e.childElementCount === 0) {
    show_loader()
    const sh = await fetchJson(`/api/trip/${id}.json`)
    e.innerHTML = sh.map((x) => stop_line(pfx, x.id, false, x.t)).join('\n')
    hide_loader()
  } else {
    e.innerHTML = ''
  }
}

function act_walk(ev, pfx, id) {
  ev.stopPropagation()
  ev.preventDefault()
  const l = `${pfx}_walk`
  const e = document.getElementById(`${pfx}`)
  if (e.childElementCount === 0) {
    e.innerHTML = (stops_info[id]?.distances || []).map((x) => stop_line(
      l, x.id, x.m
    )).join('\n')
  } else {
    e.innerHTML = ''
  }
}
