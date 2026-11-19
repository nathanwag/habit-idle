import { api } from '../api.js';
import { html, nav, icons, scheduleText, hours } from '../ui.js';

export async function view() {
  const list = await api.get('habits');
  const active = list.filter((h) => !h.current.archived);
  const archived = list.filter((h) => h.current.archived);
  const row = (h) => html`<a class="row" href="#/habito?id=${h.id}">
      <span style="flex:1;min-width:0"><span style="display:block;font-weight:600">${h.name}</span>
      <span class="small muted">${scheduleText(h.current.schedule)} · +${hours(h.effort)}</span></span>
      <span class="pf" style="font-size:16px;color:${h.streak >= 7 ? 'var(--accent)' : 'var(--text)'}">${h.streak}${h.current.schedule.kind === 'weekly' ? ' sem' : 'd'}</span></a>`;
  return {
    html: html`<div class="content">
      <div class="head"><h1 class="title">Hábitos</h1><a class="btn sm" href="#/habito/novo">${icons.plus}Novo</a></div>
      <div class="pad stack">
        ${active.length ? html`<div><div class="sec" style="display:flex;justify-content:space-between"><span>${active.length} ATIVOS</span><span>SEQUÊNCIA</span></div><div class="card">${active.map(row)}</div></div>`
          : html`<p class="empty">Crie o primeiro hábito: cada vez que você cumprir, o esquadrão ganha fôlego.</p>`}
        ${archived.length ? html`<div><div class="sec">ARQUIVADOS</div><div class="card">${archived.map(row)}</div></div>` : ''}
      </div></div>${nav('#/habitos')}`,
  };
}
