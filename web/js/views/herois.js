import { api } from '../api.js';
import { html, nav, toast } from '../ui.js';
import { sprite, coin, num, STATS, BLURB, statOf } from '../game.js';

export async function view(ctx) {
  const s = await api.get('state');
  const g = s.game;
  const free = s.catalog.filter((c) => !g.heroes.some((h) => h.id === c.id));
  const card = (h) => {
    const spec = statOf(h.spec);
    return html`<div class="card" style="display:flex;gap:12px;align-items:center;padding:10px 12px 10px 10px;${h.inSquad ? 'border-color:var(--accent)' : ''}">
      <a href="#/heroi?id=${h.id}" class="pad-box" aria-label="Abrir ${h.name}">${sprite(h.id, 'x3', 'idle')}</a>
      <div style="flex:1;min-width:0">
        <div style="display:flex;justify-content:space-between;align-items:baseline"><a href="#/heroi?id=${h.id}" class="pf" style="font-size:18px;color:var(--text);text-decoration:none">${h.name}</a><span class="pf" style="font-size:14px">${num(h.power)}</span></div>
        <div class="small muted">${h.class} · Nv ${h.level} · <span style="color:${spec.color}">${spec.label}</span></div>
        <div class="st4">${STATS.map((st, i) => html`<span>${st.label}<b>${h.points[i]}</b></span>`)}</div>
        <button type="button" class="ghost" style="height:36px;margin-top:8px;font-size:14px" data-click="toggle" data-id="${h.id}" aria-pressed="${h.inSquad}">${h.inSquad ? 'No esquadrão' : 'Na reserva'}</button>
      </div></div>`;
  };
  return {
    html: html`<div class="content">
      <div class="head"><div><h1 class="title">Esquadrão</h1><div class="small muted">${g.squad.length} de 3 vagas · o primeiro leva os golpes do chefe</div></div>
        <div style="text-align:right"><div class="label">PODER</div><div class="pf" style="font-size:24px;line-height:1">${num(g.power)}</div><div class="pf small" style="color:var(--accent)">${num(g.coins)} ${coin}</div></div></div>
      <div class="pad stack">
        ${g.recruits > 0 ? html`<div class="card" style="padding:12px;border-color:var(--accent)">
          <div class="pf" style="font-size:16px">Escolha quem entra no grupo</div>
          <div class="small muted" style="margin-bottom:8px">Ele entra no nível do herói mais baixo.</div>
          <div class="stack" style="gap:8px">${free.map((c) => html`<button type="button" class="pick" data-click="recruit" data-id="${c.id}">
            <span class="pad-box" style="height:64px">${sprite(c.id, 'x2', 'idle')}</span>
            <span style="flex:1"><span class="pf" style="display:block;font-size:16px">${c.name} <span class="small" style="color:${statOf(c.spec).color}">${statOf(c.spec).label}</span></span>
            <span class="small muted">${BLURB[c.id]}</span></span></button>`)}</div></div>` : ''}
        ${g.heroes.map(card)}
        ${free.length && !g.recruits ? html`<p class="small muted" style="text-align:center">Mais ${free.length} ${free.length === 1 ? 'herói entra' : 'heróis entram'} vencendo os chefes dos andares 10 a 40.</p>` : ''}
      </div></div>${nav('#/herois')}`,
    actions: {
      async toggle(el) {
        const id = el.dataset.id;
        const squad = g.squad.includes(id) ? g.squad.filter((x) => x !== id) : [...g.squad, id];
        if (!squad.length) return toast('O esquadrão precisa de pelo menos um herói');
        if (squad.length > 3) return toast('O esquadrão tem 3 vagas');
        await api.post('events', { kind: 'squad', data: { heroes: squad } });
        ctx.refresh();
      },
      async recruit(el) {
        await api.post('events', { kind: 'recruit', data: { hero: el.dataset.id } });
        ctx.refresh();
      },
    },
  };
}
