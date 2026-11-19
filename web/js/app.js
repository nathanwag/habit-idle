import { api } from './api.js';
import { toString, toast } from './ui.js';
import * as today from './views/today.js';
import * as habits from './views/habits.js';
import * as habit from './views/habit.js';
import * as habitForm from './views/habit-form.js';
import * as settings from './views/settings.js';
import * as start from './views/start.js';
import * as retorno from './views/retorno.js';
import * as torre from './views/torre.js';
import * as chefe from './views/chefe.js';
import * as herois from './views/herois.js';
import * as heroi from './views/heroi.js';
import * as itens from './views/itens.js';

const routes = {
  '/': today,
  '/inicio': start,
  '/retorno': retorno,
  '/torre': torre,
  '/chefe': chefe,
  '/herois': herois,
  '/heroi': heroi,
  '/itens': itens,
  '/habitos': habits,
  '/habito': habit,
  '/habito/novo': habitForm,
  '/habito/editar': habitForm,
  '/ajustes': settings,
};

const view = document.getElementById('view');
let actions = {};

async function render() {
  const [path, query = ''] = location.hash.slice(1).split('?');
  const route = routes[path || '/'] ?? today;
  const params = Object.fromEntries(new URLSearchParams(query));
  const ctx = { path: path || '/', params, refresh: render, go: (h) => { if (location.hash === h) render(); else location.hash = h; } };
  try {
    const out = await route.view(ctx);
    actions = out.actions ?? {};
    view.innerHTML = toString(out.html);
    out.mount?.(view);
  } catch (e) {
    view.innerHTML = `<p class="empty">Não deu para carregar: ${e.message}</p>`;
  }
}

// Ao abrir (ou voltar ao app): depois de um tempo fora, o resumo do que
// aconteceu; senão a visita já conta como vista.
async function arrive() {
  try {
    const s = await api.get('state');
    if (retorno.worthShowing(s)) {
      location.hash = '#/retorno';
      return;
    }
    await api.put('seen');
  } catch { /* sem servidor: a tela mostra o erro */ }
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
document.addEventListener('visibilitychange', async () => {
  if (document.hidden) return;
  await arrive();
  render();
});
arrive().then(render);

// Hoje e Torre mostram a luta em andamento: atualizam sozinhas a cada 30 s.
setInterval(() => {
  const path = location.hash.slice(1).split('?')[0] || '/';
  if (!document.hidden && (path === '/' || path === '/torre')) render();
}, 30000);
