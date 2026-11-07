// state

var targetDate = 'today'
var currentHash = ''

// helpers

async function fetchHTML(url) {
  const r = await fetch(url);
  if (!r.ok) {
    throw new Error(`Response is not ok: ${r.status}`);
  }
  return await r.text();
}

function updateElement(elementId, html) {
  document.getElementById(elementId).innerHTML = html
}

function cleanByClass(className) {
  Array.from(document.getElementsByClassName(className)).forEach(x => { x.innerHTML = '' })
}

function jumpToId(elementId) {
  currentHash = `#${elementId}`
  document.location.href = currentHash
}

function cleanIfNotEmpty(elementId) {
  const e = document.getElementById(elementId)
  if (e && e.innerHTML !== '') {
    e.innerHTML = ''
    return true
  }
  return false
}

// init

(async () => { updateElement('main', await fetchHTML('/stops')) })()

// actions

async function open_stop(id, toggle) {
  const tag = `stop_${id}_body`
  if (cleanIfNotEmpty(tag) && toggle) {
    return
  }
  const html = await fetchHTML(`/schedule/${id}/${targetDate}`)
  cleanByClass('stop_body')
  updateElement(tag, html)
  jumpToId(`stop_${id}`)
}

async function select_date(id, dt) {
  targetDate = dt
  const html = await fetchHTML(`/schedule/${id}/${dt}`)
  cleanByClass('stop_body')
  updateElement(`stop_${id}_body`, html)
}

async function open_trip(id) {
  const tag = `trip_${id}_body`
  if (cleanIfNotEmpty(tag)) {
    return
  }
  const html = await fetchHTML(`/trip/${id}`)
  cleanByClass('trip_body')
  updateElement(tag, html)
}

// events

window.addEventListener('hashchange', () => {
  const h = window.location.hash;
  if (h === currentHash) {
    return
  }
  open_stop(+h.match(/\d+/), false)
});
