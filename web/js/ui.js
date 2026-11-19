// Tagged template que escapa os valores. Template aninhado (ou lista deles)
// entra cru. Booleano vira "true"/"false": aria-pressed="${done}" funciona.
const RAW = Symbol('raw');

export function html(strings, ...values) {
  let out = strings[0];
  values.forEach((v, i) => { out += piece(v) + strings[i + 1]; });
  return { [RAW]: out };
}

function piece(v) {
  if (v == null) return '';
  if (Array.isArray(v)) return v.map(piece).join('');
  if (typeof v === 'object' && RAW in v) return v[RAW];
  return String(v).replace(/[&<>"']/g, (c) => `&#${c.charCodeAt(0)};`);
}

export const toString = (tpl) => piece(tpl);

export const icons = {
  back: html`<svg aria-hidden="true" width="20" height="20" viewBox="0 0 20 20"><path d="M12 4 L6 10 L12 16" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="square"></path></svg>`,
  close: html`<svg aria-hidden="true" width="18" height="18" viewBox="0 0 18 18"><path d="M3 3 L15 15 M15 3 L3 15" stroke="currentColor" stroke-width="2.5"></path></svg>`,
  lock: html`<svg aria-hidden="true" width="22" height="22" viewBox="0 0 22 22"><path d="M6 10 V7 a5 5 0 0 1 10 0 V10 M4 10 H18 V20 H4 Z" fill="none" stroke="currentColor" stroke-width="2"></path></svg>`,
  plus: html`<svg aria-hidden="true" width="14" height="14" viewBox="0 0 14 14"><path d="M7 1 V13 M1 7 H13" stroke="currentColor" stroke-width="3"></path></svg>`,
};

export function top(title, backTo, extra = '') {
  return html`<header class="top"><a class="back" href="${backTo}" aria-label="Voltar">${icons.back}</a><h1>${title}</h1>${extra}</header>`;
}

const TABS = [['#/', 'Hoje'], ['#/torre', 'Torre'], ['#/herois', 'Heróis'], ['#/habitos', 'Hábitos'], ['#/ajustes', 'Ajustes']];

export function nav(current) {
  return html`<nav class="nav" aria-label="Principal">${TABS.map(([href, label]) =>
    html`<a href="${href}" ${href === current ? html`aria-current="page"` : ''}>${label}</a>`)}</nav>`;
}

export function toast(msg) {
  const el = document.createElement('div');
  el.className = 'toast';
  el.setAttribute('role', 'status');
  el.textContent = msg;
  document.body.append(el);
  setTimeout(() => el.remove(), 2200);
}

export const EFFORTS = [[60, 'leve'], [120, 'médio'], [180, 'pesado']];
export const WEEKDAYS = ['D', 'S', 'T', 'Q', 'Q', 'S', 'S'];
export const WEEKDAY_NAMES = ['domingo', 'segunda', 'terça', 'quarta', 'quinta', 'sexta', 'sábado'];

export function scheduleText(s) {
  if (s.kind === 'daily') return 'todo dia';
  if (s.kind === 'weekly') return `${s.times}x por semana`;
  const names = ['dom', 'seg', 'ter', 'qua', 'qui', 'sex', 'sáb'];
  return s.days.map((d) => names[d]).join(', ');
}

export const hours = (min) => `${min / 60}h`;
