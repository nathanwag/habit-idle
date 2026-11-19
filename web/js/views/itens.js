import { api } from '../api.js';
import { html, top, toast } from '../ui.js';
import { coin, num, SLOTS, RARITY, itemDesc } from '../game.js';

export async function view(ctx) {
  const { hero, slot } = ctx.params;
  const s = await api.get('state');
  const g = s.game;
  const h = g.heroes.find((x) => x.id === hero);
  const label = SLOTS.find(([k]) => k === slot)?.[1];
  const list = g.items
    .filter((it) => !slot || it.slot === slot)
    .sort((a, b) => b.rarity - a.rarity || b.floor - a.floor);
  const pair = (it) => g.items.find((o) => o.id !== it.id && o.name === it.name && o.rarity === it.rarity && !o.hero);
  const owner = (it) => g.heroes.find((x) => x.id === it.hero)?.name;
  const back = h ? `#/heroi?id=${h.id}` : '#/herois';
  return {
    html: html`${top(h ? `${label} · ${h.name}` : 'Inventário', back, html`<span class="pf" style="color:var(--accent)">${num(g.coins)} ${coin}</span>`)}
    <div class="content pad stack" style="padding-top:12px">
      ${list.length ? list.map((it) => html`<div class="card" style="padding:10px 12px">
        <div style="display:flex;justify-content:space-between;gap:8px;align-items:baseline">
          <span style="font-weight:600">${it.name}${it.refine ? html` <span class="pf r3">+${it.refine}</span>` : ''}</span>
          <span class="small r${it.rarity}">${RARITY[it.rarity]}</span></div>
        <div class="small muted">${itemDesc(it)} · andar ${it.floor}${it.hero ? ` · com ${owner(it)}` : ''}</div>
        <div style="display:flex;gap:8px;margin-top:8px">
          ${h && it.hero !== h.id ? html`<button type="button" class="ghost" style="height:40px;font-size:14px" data-click="equip" data-id="${it.id}">Equipar</button>` : ''}
          ${h && it.hero === h.id ? html`<button type="button" class="ghost" style="height:40px;font-size:14px" data-click="unequip" data-id="${it.id}">Tirar</button>` : ''}
          ${pair(it) && it.refine < 10 ? html`<button type="button" class="ghost" style="height:40px;font-size:14px;color:var(--accent);border-color:var(--accent)" data-click="refine" data-id="${it.id}" data-with="${pair(it).id}">Refinar</button>` : ''}
          <button type="button" class="ghost" style="height:40px;font-size:14px" data-click="sell" data-id="${it.id}">Vender · ${it.value} ${coin}</button>
        </div></div>`)
      : html`<p class="empty">Nenhum item${label ? ` de ${label.toLowerCase()}` : ''} ainda. Baús caem na torre e quando você completa todas as missões do dia.</p>`}
    </div>`,
    actions: {
      async equip(el) {
        await api.post('events', { kind: 'equip', data: { hero: h.id, item: Number(el.dataset.id) } });
        ctx.go(back);
      },
      async unequip(el) {
        await api.post('events', { kind: 'unequip', data: { item: Number(el.dataset.id) } });
        ctx.refresh();
      },
      async refine(el) {
        await api.post('events', { kind: 'refine', data: { item: Number(el.dataset.id), with: Number(el.dataset.with) } });
        toast('Refinado!');
        ctx.refresh();
      },
      async sell(el) {
        await api.post('events', { kind: 'sell', data: { item: Number(el.dataset.id) } });
        ctx.refresh();
      },
    },
  };
}
