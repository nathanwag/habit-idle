import { api } from '../api.js';
import { html } from '../ui.js';
import { sprite, BLURB, statOf } from '../game.js';

let chosen = 'kael';

export async function view(ctx) {
  const s = await api.get('state');
  if (s.game.started) { ctx.go('#/'); return { html: '' }; }
  return {
    html: html`<div class="content pad stack" style="padding-top:22px">
      <div><h1 class="title">Escolha seu herói</h1>
        <p class="muted" style="margin:6px 0 0">Ele começa a torre sozinho. Os outros entram quando você vencer os chefes dos andares 10, 20, 30 e 40.</p></div>
      ${s.catalog.map((h) => {
        const st = statOf(h.spec);
        return html`<button type="button" class="pick" aria-pressed="${chosen === h.id}" data-click="pick" data-id="${h.id}">
          <span class="pad-box">${sprite(h.id, 'x3', 'idle')}</span>
          <span style="flex:1;min-width:0"><span style="display:flex;justify-content:space-between;align-items:baseline"><span class="pf" style="font-size:18px">${h.name}</span><span class="pf" style="color:${st.color}">${st.label}</span></span>
          <span class="small muted" style="display:block">${h.class}</span>
          <span style="display:block;margin-top:4px">${BLURB[h.id]}</span></span></button>`;
      })}
      <button type="button" class="btn" data-click="start">Começar com ${s.catalog.find((h) => h.id === chosen).name}</button>
    </div>`,
    actions: {
      pick(el) { chosen = el.dataset.id; ctx.refresh(); },
      async start() {
        await api.post('events', { kind: 'start', data: { hero: chosen } });
        await api.put('seen');
        ctx.go('#/');
      },
    },
  };
}
