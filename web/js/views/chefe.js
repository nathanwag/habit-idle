import { api } from '../api.js';
import { html, top } from '../ui.js';
import { coin, num } from '../game.js';

export async function view(ctx) {
  const s = await api.get('state');
  const g = s.game;
  const k = g.nextBoss / 10;
  const last = g.log.filter((e) => e.kind === 'boss' && e.floor === g.nextBoss).at(-1);
  const ratio = Math.min(1, g.power / g.recommended);
  const gap = g.recommended - g.power;
  let action;
  if (g.canFight) action = html`<button type="button" class="btn danger" data-click="fight">Desafiar agora</button>`;
  else if (g.training) action = html`<p class="small muted" style="text-align:center">Uma tentativa por dia. Volte amanhã.</p>`;
  else action = html`<p class="small muted" style="text-align:center">A luta começa sozinha quando o esquadrão chegar no andar ${g.nextBoss}.</p>`;
  return {
    html: html`${top(`Andar ${g.nextBoss}`, '#/torre', html`<span class="pf small" style="color:var(--bad);border:2px solid var(--bad);padding:2px 8px">CHEFE</span>`)}
    <div class="scn" role="img" aria-label="Zumbi colosso" style="height:220px;display:flex;align-items:flex-end;justify-content:center;padding-bottom:14px;background-position:-420px 100%">
      <div class="boss" style="position:relative;z-index:1"></div></div>
    <div class="content pad stack" style="padding-top:14px">
      <div><div class="pf" style="font-size:28px;line-height:1">Zumbi colosso</div>
        <div class="small muted" style="margin-top:4px">Bate sempre no primeiro herói do esquadrão. A luta dura até 2 minutos.</div></div>
      <div class="card" style="padding:12px">
        <div style="display:flex;justify-content:space-between"><span class="label">SEU PODER</span><span class="label">RECOMENDADO</span></div>
        <div style="display:flex;justify-content:space-between;align-items:baseline"><span class="pf" style="font-size:24px">${num(g.power)}</span><span class="pf" style="font-size:24px;color:var(--bad)">${num(g.recommended)}</span></div>
        <span class="bar" style="height:10px;margin-top:6px"><span style="width:${Math.round(ratio * 100)}%"></span></span>
        <div class="small" style="margin-top:8px;color:${gap > 0 ? 'var(--warn)' : 'var(--good)'}">${gap > 0 ? `Luta arriscada: faltam ${num(gap)} de poder.` : 'Você está acima do recomendado.'}</div>
      </div>
      <div class="card" style="padding:2px 12px">
        <div class="row" style="padding:0"><span class="muted" style="flex:1">Recompensa</span><span class="pf">${num(250 * k)} ${coin}</span></div>
        <div class="row" style="padding:0"><span class="muted" style="flex:1">Baú</span><span class="r2">épico ou lendário</span></div>
        ${g.nextBoss <= 40 ? html`<div class="row" style="padding:0"><span class="muted" style="flex:1">Libera</span><span>um herói novo</span></div>` : ''}
      </div>
      ${last ? html`<div class="card" role="status" style="padding:12px;border-color:${last.win ? 'var(--good)' : 'var(--bad)'}">
        <div class="pf" style="font-size:18px;color:${last.win ? 'var(--good)' : 'var(--bad)'}">${last.win ? 'Vitória!' : 'Derrota'}</div>
        ${last.win ? '' : html`<div class="small muted">O esquadrão treina no andar anterior até ficar mais forte. Nada se perde.</div>`}</div>` : ''}
      <div class="stack" style="margin-top:auto;padding-bottom:16px">
        ${action}
        <a href="#/herois" class="ghost">Fortalecer heróis</a>
      </div>
    </div>`,
    actions: {
      async fight() { await api.post('events', { kind: 'challenge' }); ctx.refresh(); },
    },
  };
}
