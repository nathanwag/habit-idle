import { api } from '../api.js';
import { html, nav } from '../ui.js';
import { coin, num, hm, clock, battleScene } from '../game.js';

const NAMES = ['Salão de pedra', 'Cripta úmida', 'Galeria dos ossos', 'Cripta de gelo', 'Forja abandonada'];
const zone = (n) => NAMES[Math.floor((n - 1) / 10) % NAMES.length];

function describe(e, heroes) {
  const name = (id) => heroes.find((h) => h.id === id)?.name;
  switch (e.kind) {
    case 'floor': return html`andar ${e.floor} · +${e.coins} ${coin}`;
    case 'boss': return e.win ? `venceu o chefe do andar ${e.floor}` : `perdeu para o chefe do andar ${e.floor}`;
    case 'level': return `${name(e.hero)} subiu para o nível ${e.floor}`;
    case 'recruit': return `${name(e.hero)} entrou no grupo`;
    case 'death': return `o esquadrão caiu no inimigo ${e.enemy} do andar ${e.floor}`;
    default: return '';
  }
}

export async function view() {
  const s = await api.get('state');
  const g = s.game;
  const moving = s.folego.exploring;
  // Para cima só se vê o próximo andar, e só o chefe tem cara antes de chegar.
  const rows = [g.nextBoss === g.floor + 1
    ? html`<a href="#/chefe" class="fl" style="min-height:52px"><span class="num" style="background:var(--bad);color:var(--bg);height:46px;flex-direction:column;line-height:1.05;font-weight:700"><span style="font-size:10px">CHEFE</span>${g.nextBoss}</span>
      <span><span style="display:block;font-weight:600">Zumbi colosso</span><span class="small muted">pede <b style="color:var(--text)">${num(g.recommended)}</b> de poder · você tem ${num(g.power)}</span></span><span class="pf" style="color:var(--bad)">›</span></a>`
    : html`<div class="fl"><span class="num pf" style="background:var(--surface2);color:var(--muted);border:2px dashed var(--line2);font-size:18px" aria-label="Próximo andar">?</span><span class="small muted">Próximo andar</span><span></span></div>`];
  for (let n = g.floor; n >= Math.max(1, g.floor - 3); n--) {
    if (n === g.floor) {
      const e = g.enc;
      let state = 'Caminhando';
      if (g.training) state = 'Treinando aqui';
      else if (!moving) state = 'Parado aqui';
      else if (e.phase === 'fight') state = 'Lutando';
      else if (e.phase === 'rest') state = 'Descansando';
      const cells = Array.from({ length: e.of }, (_, i) => html`<i class="${i + 1 < e.enemy ? 'done' : i + 1 === e.enemy ? 'now' : ''}"></i>`);
      rows.push(html`<div class="fl" style="min-height:60px"><span class="num" style="background:var(--accent);color:var(--bg);height:56px;box-shadow:0 0 12px var(--accent)"><b>${n}</b></span>
        <span><span style="display:flex;justify-content:space-between"><span style="font-weight:600">${state}</span><span class="small muted">inimigo ${e.enemy} de ${e.of}</span></span>
        <span class="foes" aria-hidden="true">${cells}</span></span><span></span></div>`);
    } else {
      rows.push(html`<div class="fl"><span class="num" style="background:var(--line2)">${n}</span><span class="small muted">${zone(n)}</span><span class="pf small" style="color:var(--accent)">+${10 + 3 * n}</span></div>`);
    }
  }
  const recent = g.log.filter((e) => e.kind !== 'chest').slice(-5).reverse();
  return {
    html: html`<div class="content">
      ${battleScene(g, moving, { height: 170, bottom: 8, top: html`<div class="hud" style="justify-content:space-between"><span class="pf" style="font-size:24px">Torre</span><span class="pf" style="color:var(--accent)">${num(g.coins)} ${coin}</span></div>` })}
      <div class="grid3" style="padding:10px 14px 6px;text-align:center">
        <div><div class="pf" style="font-size:20px">${g.floor}</div><div class="small muted">andar atual</div></div>
        <div><div class="pf" style="font-size:20px">${g.beaten}</div><div class="small muted">chefes vencidos</div></div>
        <div><div class="pf" style="font-size:20px;color:var(--good)">${hm(Math.max(0, s.folego.balance))}</div><div class="small muted">de fôlego</div></div>
      </div>
      <div role="list" aria-label="Andares" style="position:relative;padding:4px 0 0 12px;display:flex;flex-direction:column;gap:4px">
        <div aria-hidden="true" style="position:absolute;left:37px;top:14px;bottom:14px;width:2px;background:var(--line)"></div>
        ${rows}
      </div>
      ${recent.length ? html`<div class="card small muted" style="margin:10px 14px 0;padding:10px 12px;line-height:1.7">
        <div class="pf" style="font-size:13px;color:var(--text)">Registro</div>
        ${recent.map((e) => html`<div><span class="pf" style="color:var(--text)">${clock(e.at)}</span> ${describe(e, g.heroes)}</div>`)}</div>` : ''}
    </div>${nav('#/torre')}`,
  };
}
