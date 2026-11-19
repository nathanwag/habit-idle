import { api } from '../api.js';
import { html, top } from '../ui.js';
import { sprite, coin, num, STATS, SLOTS, RARITY, itemDesc, statOf } from '../game.js';

export async function view(ctx) {
  const s = await api.get('state');
  const g = s.game;
  const h = g.heroes.find((x) => x.id === ctx.params.id);
  if (!h) { ctx.go('#/herois'); return { html: '' }; }
  const spec = statOf(h.spec);
  return {
    html: html`${top(h.name, '#/herois', html`<span class="pf" style="color:var(--accent)">${num(g.coins)} ${coin}</span>`)}
    <div class="scn rest" style="height:170px;display:flex;align-items:flex-end;gap:16px;padding:0 14px 12px">
      <span style="position:relative;z-index:1">${sprite(h.id, 'x4', 'idle')}</span>
      <div style="position:relative;z-index:1;flex:1;padding-bottom:6px">
        <div class="small muted">${h.class} · especialista em <span style="color:${spec.color}">${spec.name.toLowerCase()}</span></div>
        <div class="pf" style="font-size:26px;line-height:1.1">Nível ${h.level}</div>
        <span class="bar" style="height:6px;margin-top:6px"><span style="width:${Math.round(100 * h.xp / h.xpNext)}%"></span></span>
        <div class="small muted" style="margin-top:4px">${h.xp}/${h.xpNext} XP${h.inSquad ? '' : ' · na reserva, não ganha XP'}</div>
      </div>
      <div style="position:relative;z-index:1;text-align:right;padding-bottom:6px"><div class="label">PODER</div><div class="pf" style="font-size:22px">${num(h.power)}</div></div>
    </div>
    <div class="content pad stack" style="padding-top:12px">
      <div><div class="sec">STATUS</div>
        <div class="card">${STATS.map((st, i) => html`<div class="stat-row">
          <span class="pf" style="width:52px;color:${st.color}">${st.label}</span>
          <span class="pf" style="flex:1;font-size:20px">${h.points[i]}${h.bought[i] ? html` <span class="small muted">(${h.bought[i]} comprados)</span>` : ''}</span>
          <button type="button" class="buy" data-click="buy" data-stat="${st.key}" ${g.coins < h.costs[i] ? html`disabled` : ''}
            aria-label="Comprar 1 ponto de ${st.name} por ${h.costs[i]} moedas">+1 · ${num(h.costs[i])} ${coin}</button></div>`)}</div></div>
      <div><div class="sec">EQUIPAMENTO</div>
        <div class="card">${SLOTS.map(([slot, label]) => {
          const it = g.items.find((x) => x.id === h.equip[slot]);
          return html`<a class="row" href="#/itens?hero=${h.id}&slot=${slot}">
            <span class="slot" style="${it ? `border-color:var(--${['muted', 'def', 'sor', 'accent'][it.rarity]})` : 'border-style:dashed'}"></span>
            <span style="flex:1;min-width:0">${it
              ? html`<span style="display:block;font-weight:600">${it.name}${it.refine ? html` <span class="pf r3">+${it.refine}</span>` : ''}</span><span class="small r${it.rarity}">${RARITY[it.rarity]} · ${itemDesc(it)}</span>`
              : html`<span style="display:block">${label}</span><span class="small muted">vazio · toque para escolher</span>`}</span>
            <span class="pf muted">›</span></a>`;
        })}</div></div>
      <p class="small muted" style="text-align:center">Moedas vêm da torre, e a torre anda com o fôlego das suas missões.</p>
    </div>`,
    actions: {
      async buy(el) {
        await api.post('events', { kind: 'buy', data: { hero: h.id, stat: el.dataset.stat } });
        ctx.refresh();
      },
    },
  };
}
