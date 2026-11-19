import { api } from '../api.js';
import { html, top, scheduleText, hours } from '../ui.js';

const MONTHS = ['jan', 'fev', 'mar', 'abr', 'mai', 'jun', 'jul', 'ago', 'set', 'out', 'nov', 'dez'];

export async function view(ctx) {
  const d = await api.get(`habits/${ctx.params.id}`);
  const s = d.stats;
  const weekly = d.current.schedule.kind === 'weekly';
  const months = [];
  for (let w = 0; w < s.grid.length; w += 7) {
    const m = MONTHS[Number(s.grid[w].day.slice(5, 7)) - 1];
    if (months.at(-1) !== m) months.push(m);
  }
  return {
    html: html`${top(d.habit.name, '#/habitos', html`<a href="#/habito/editar?id=${d.habit.id}" style="padding:12px 0">Editar</a>`)}
    <div class="content pad stack" style="padding-top:14px">
      <div class="grid4">
        <div class="card stat"><b style="color:var(--accent)">${s.streak}${weekly ? ' sem' : 'd'}</b><span class="small muted">sequência</span></div>
        <div class="card stat"><b>${s.best}${weekly ? ' sem' : 'd'}</b><span class="small muted">recorde</span></div>
        <div class="card stat"><b>${s.total}</b><span class="small muted">vezes</span></div>
        <div class="card stat"><b>${s.rate < 0 ? '—' : `${s.rate}%`}</b><span class="small muted">feito</span></div>
      </div>
      <div class="card" style="padding:12px">
        <div class="pf" style="font-size:15px;margin-bottom:8px">Últimas 12 semanas</div>
        <div class="hm" role="img" aria-label="Calendário das últimas 12 semanas">${s.grid.map((g) => html`<i class="${g.state}" title="${g.day}"></i>`)}</div>
        <div class="small muted" style="display:flex;justify-content:space-between;margin-top:6px">${months.map((m) => html`<span>${m}</span>`)}</div>
      </div>
      <div class="card" style="padding:2px 12px">
        <div class="row" style="padding:0"><span class="muted" style="flex:1">Frequência</span><span>${scheduleText(d.current.schedule)}</span></div>
        <div class="row" style="padding:0"><span class="muted" style="flex:1">Fôlego por check</span><span class="pf" style="color:var(--good)">+${hours(d.habit.effort)}</span></div>
        ${d.current.archived ? html`<div class="row" style="padding:0"><span class="muted" style="flex:1">Situação</span><span>arquivado</span></div>` : ''}
      </div>
    </div>`,
  };
}
