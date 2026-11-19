import { toString, toast } from './ui.js';
import * as today from './views/today.js';
import * as habits from './views/habits.js';
import * as habit from './views/habit.js';
import * as habitForm from './views/habit-form.js';
import * as settings from './views/settings.js';
import * as soon from './views/soon.js';

const routes = {
  '/': today,
  '/habitos': habits,
  '/habito': habit,
  '/habito/novo': habitForm,
  '/habito/editar': habitForm,
  '/ajustes': settings,
  '/torre': soon,
  '/herois': soon,
};

const view = document.getElementById('view');
let actions = {};

async function render() {
  const [path, query = ''] = location.hash.slice(1).split('?');
  const route = routes[path || '/'] ?? today;
  const params = Object.fromEntries(new URLSearchParams(query));
  const ctx = { path: path || '/', params, refresh: render, go: (h) => { location.hash = h; } };
  try {
    const out = await route.view(ctx);
    actions = out.actions ?? {};
    view.innerHTML = toString(out.html);
    out.mount?.(view);
  } catch (e) {
    view.innerHTML = `<p class="empty">Não deu para carregar: ${e.message}</p>`;
  }
}

// Um handler só no #view: os filhos são recriados a cada render.
for (const type of ['click', 'change', 'submit']) {
  view.addEventListener(type, async (ev) => {
    const el = ev.target.closest(`[data-${type}]`);
    if (!el) return;
    const fn = actions[el.dataset[type]];
    if (!fn) return;
    if (type === 'submit') ev.preventDefault();
    try {
      await fn(el, ev);
    } catch (e) {
      toast(e.message);
    }
  });
}

window.addEventListener('hashchange', render);
document.addEventListener('visibilitychange', () => { if (!document.hidden) render(); });
render();
