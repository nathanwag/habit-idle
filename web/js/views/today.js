import { api } from '../api.js';
import { html, nav, hours } from '../ui.js';
import { sprite, coin, num, hm, clock, battleScene } from '../game.js';

function scene(g, f) {
  const squad = g.squad.map((id) => g.heroes.find((h) => h.id === id));
  const hud = html`<div class="hud" style="justify-content:space-between">
      <a href="#/torre" class="pf" style="font-size:13px;color:var(--muted);text-decoration:none;padding:6px 0;white-space:nowrap">ANDAR <b style="font-size:20px;color:var(--text)">${g.floor}</b></a>
      <a href="#/chefe" class="pf" style="font-size:12px;color:var(--bad);text-decoration:none;padding:10px 0;white-space:nowrap">CHEFE ${g.nextBoss}</a>
    </div>`;
  return html`${battleScene(g, f.exploring, { height: 290, bottom: 58, top: hud })}
  <div class="cards">${squad.map((h) => html`<a href="#/heroi?id=${h.id}" class="card hc">${sprite(h.id, 'x2', 'idle')}
    <span style="min-width:0;flex:1"><span class="pf" style="display:block;font-size:14px">${h.name}</span>
    <span style="display:block;font-size:11px;color:var(--muted)">Nv ${h.level}</span>
    <span class="bar" style="margin-top:4px" title="vida"><span style="width:${Math.max(0, Math.round(100 * (g.enc.life[h.id] ?? 1) / (g.enc.maxLife[h.id] || 1)))}%;background:var(--good)"></span></span>
    <span class="bar" style="margin-top:2px;height:2px" title="XP"><span style="width:${Math.round(100 * h.xp / h.xpNext)}%"></span></span></span></a>`)}</div>`;
}

function folegoLine(f, g) {
  if (f.balance < 0) return html`<span style="color:var(--bad)">dívida: a próxima missão paga primeiro</span>`;
  if (f.exploring && f.until) return html`explorando até ${clock(f.until)}`;
  if (!g.started) return 'escolha um herói para explorar';
  return 'parado: faça uma missão';
}

export async function view(ctx) {
  const s = await api.get('state');
  const { game: g, folego: f } = s;
  if (!g.started) { ctx.go('#/inicio'); return { html: '' }; }
  const ratio = s.missions ? s.done / s.missions : 0;
  const rows = s.items.map((i) => html`
    <div class="row" style="padding:2px 12px 2px 2px">
      <button type="button" class="chk" data-click="toggle" data-id="${i.habit.id}" data-done="${i.done}" aria-pressed="${i.done}"
        aria-label="${i.done ? 'Desmarcar' : 'Marcar'} ${i.habit.name}"><span class="box"></span></button>
      <a href="#/habito?id=${i.habit.id}" style="flex:1;min-width:0;color:var(--text);text-decoration:none">
        <span style="display:block;font-weight:600">${i.habit.name}</span>
        <span class="small muted">${i.schedule.kind === 'weekly' ? `semanal · ${i.doneThisWeek}/${i.schedule.times}` : 'missão do dia'}</span></a>
      <span class="pf" style="color:${i.done ? 'var(--good)' : 'var(--muted)'}">+${hours(i.habit.effort)}</span>
    </div>`);
  return {
    html: html`<div class="content">
      ${scene(g, f)}
      <div style="padding:14px 14px 0;display:flex;align-items:center;gap:14px">
        <svg class="ring" role="img" aria-label="${s.done} de ${s.missions} missões feitas hoje" width="64" height="64" viewBox="0 0 64 64">
          <circle cx="32" cy="32" r="26" fill="none" style="stroke:var(--line)" stroke-width="7"></circle>
          <circle cx="32" cy="32" r="26" fill="none" style="stroke:var(--accent)" stroke-width="7" stroke-dasharray="${(163.4 * ratio).toFixed(1)} 163.4" transform="rotate(-90 32 32)"></circle>
          <text x="32" y="37" text-anchor="middle" font-family="Pixelify Sans, monospace" font-size="16" style="fill:var(--text)">${s.done}/${s.missions}</text>
        </svg>
        <div style="flex:1;min-width:0"><div class="label">FÔLEGO</div>
          <div class="big-num" style="color:${f.balance < 0 ? 'var(--bad)' : 'var(--text)'}">${hm(f.balance)}</div>
          <div class="small muted">${folegoLine(f, g)}</div></div>
        <div style="text-align:right"><div class="label">PODER</div>
          <div class="pf" style="font-size:22px;line-height:1">${num(g.power)}</div>
          ${g.fatigue ? html`<div class="pf small" style="color:var(--warn)">cansado −${g.fatigue}%</div>` : html`<div class="pf small" style="color:var(--accent)">${num(g.coins)} ${coin}</div>`}</div>
      </div>
      <div style="margin:10px 14px 0;display:flex;justify-content:space-between;align-items:center">
        <span class="pf" style="font-size:17px">Missões de hoje</span>
        <a href="#/habitos" class="small" style="padding:8px 0">Gerenciar</a>
      </div>
      <div class="pad">${s.items.length ? html`<div class="card">${rows}</div>`
        : html`<p class="empty">Nenhum hábito hoje. <a href="#/habito/novo">Crie um</a> para dar fôlego ao esquadrão.</p>`}
        ${s.missions >= 2 && s.done < s.missions ? html`<p class="small muted" style="margin:8px 2px 0">Completar todas dá +1h e um baú.</p>` : ''}</div>
    </div>${nav('#/')}`,
    actions: {
      async toggle(el) {
        await api.put('checks', { habitId: Number(el.dataset.id), day: s.day, done: el.dataset.done !== 'true' });
        ctx.refresh();
      },
    },
  };
}
