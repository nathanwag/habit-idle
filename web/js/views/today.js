import { api } from '../api.js';
import { html, nav, hours } from '../ui.js';

export async function view(ctx) {
  const t = await api.get('today');
  const missions = t.items.filter((i) => i.schedule.kind !== 'weekly');
  const done = missions.filter((i) => i.done).length;
  const rows = t.items.map((i) => html`
    <div class="row">
      <button type="button" class="chk" data-click="toggle" data-id="${i.habit.id}" data-done="${i.done}" aria-pressed="${i.done}"
        aria-label="${i.done ? 'Desmarcar' : 'Marcar'} ${i.habit.name}"><span class="box"></span></button>
      <a href="#/habito?id=${i.habit.id}" style="flex:1;min-width:0;color:var(--text);text-decoration:none">
        <span style="display:block;font-weight:600">${i.habit.name}</span>
        <span class="small muted">${i.schedule.kind === 'weekly' ? `${i.doneThisWeek}/${i.schedule.times} na semana` : 'missão do dia'}</span>
      </a>
      <span class="pf" style="color:${i.done ? 'var(--good)' : 'var(--muted)'}">+${hours(i.habit.effort)}</span>
    </div>`);
  return {
    html: html`<div class="content">
      <div class="head"><h1 class="title">Hoje</h1><span class="pf muted">${done}/${missions.length}</span></div>
      <div class="pad">
        ${t.items.length ? html`<div class="card">${rows}</div>` : html`<p class="empty">Nenhum hábito para hoje.</p>`}
        <p style="margin-top:12px"><a href="#/habito/novo">+ Novo hábito</a></p>
      </div></div>${nav('#/')}`,
    actions: {
      async toggle(el) {
        await api.put('checks', { habitId: Number(el.dataset.id), day: t.day, done: el.dataset.done !== 'true' });
        ctx.refresh();
      },
    },
  };
}
