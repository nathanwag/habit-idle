import { api } from '../api.js';
import { html } from '../ui.js';
import { sprite, coin, num, hm, clock, RARITY, itemDesc } from '../game.js';

// Só vale a pena abrir o resumo depois de um tempo fora.
const AWAY_MS = 2 * 60 * 60 * 1000;

export function since(s) {
  const from = s.lastSeen ? new Date(s.lastSeen) : null;
  return from ? s.game.log.filter((e) => new Date(e.at) > from) : [];
}

export function worthShowing(s) {
  if (!s.game.started || !s.lastSeen) return false;
  return new Date(s.now) - new Date(s.lastSeen) > AWAY_MS && since(s).length > 0;
}

const RCOLOR = ['var(--muted)', 'var(--def)', 'var(--sor)', 'var(--accent)'];

export async function view(ctx) {
  const s = await api.get('state');
  const g = s.game;
  const log = since(s);
  const floors = log.filter((e) => e.kind === 'floor');
  const coins = log.reduce((t, e) => t + (e.coins || 0), 0);
  const items = log.filter((e) => e.kind === 'chest').map((e) => g.items.find((it) => it.id === e.item)).filter(Boolean);
  const levels = log.filter((e) => e.kind === 'level');
  const bosses = log.filter((e) => e.kind === 'boss');
  const squad = g.squad.map((id) => g.heroes.find((h) => h.id === id));
  const hour = new Date(s.now).getHours();
  const hello = hour < 12 ? 'BOM DIA' : hour < 18 ? 'BOA TARDE' : 'BOA NOITE';
  const row = (color, title, sub) => html`<div class="row" style="min-height:56px"><span class="slot" style="border-color:${color}"></span>
    <span style="flex:1;min-width:0"><span style="display:block;font-weight:600">${title}</span><span class="small muted">${sub}</span></span></div>`;

  return {
    html: html`<div class="scn rest" role="img" aria-label="O esquadrão descansa" style="height:220px">
      <div style="position:absolute;left:0;right:0;top:18px;text-align:center;z-index:2">
        <div class="label">${hello} · ${clock(s.now)}</div>
        <div class="pf" style="font-size:26px;margin-top:2px">Enquanto você estava fora</div></div>
      <div class="team" style="left:50%;transform:translateX(-50%);gap:22px">${squad.map((h) => sprite(h.id, 'x3', 'idle'))}</div>
    </div>
    <div class="content pad stack" style="padding-top:14px">
      <div class="grid3" style="text-align:center">
        <div class="card stat"><b style="color:var(--accent)">+${floors.length}</b><span class="small muted">andares</span></div>
        <div class="card stat"><b style="color:var(--accent)">+${num(coins)} ${coin}</b><span class="small muted">moedas</span></div>
        <div class="card stat"><b style="color:var(--good)">${num(g.power)}</b><span class="small muted">poder</span></div>
      </div>
      ${items.length || levels.length || bosses.length ? html`<div class="card">
        ${bosses.map((b) => row(b.win ? 'var(--good)' : 'var(--bad)',
          b.win ? `Chefe do andar ${b.floor} vencido!` : `Derrota para o chefe do andar ${b.floor}`,
          b.win ? `+${num(b.coins)} moedas` : 'O esquadrão treina no andar anterior.'))}
        ${items.map((it) => row(RCOLOR[it.rarity], it.name, `${RARITY[it.rarity]} · ${itemDesc(it)}`))}
        ${levels.map((l) => row('var(--good)', `${g.heroes.find((h) => h.id === l.hero)?.name} subiu para o nível ${l.floor}`, 'XP da torre'))}
      </div>` : ''}
      <div class="card" style="padding:12px;border-style:dashed;border-color:var(--line2);display:flex;gap:12px;align-items:center">
        <div style="flex:1"><div class="pf" style="font-size:15px">${s.folego.exploring ? 'Ainda explorando' : 'Volte a explorar'}</div>
          <div class="small muted">${s.folego.exploring ? 'O fôlego de antes ainda está rendendo.' : 'Cada missão de hoje dá de 1h a 3h de fôlego.'}</div></div>
        <div class="pf" style="font-size:22px;color:${s.folego.balance > 0 ? 'var(--good)' : 'var(--muted)'}">${hm(s.folego.balance)}</div>
      </div>
      <div class="stack" style="margin-top:auto;padding-bottom:16px">
        <button type="button" class="btn" data-click="collect">Coletar e ver missões</button>
        <a href="#/torre" class="small" style="text-align:center;padding:8px">Ver o registro da torre</a>
      </div>
    </div>`,
    actions: {
      async collect() { await api.put('seen'); ctx.go('#/'); },
    },
  };
}
