import { api } from '../api.js';
import { html, top, EFFORTS, WEEKDAYS, WEEKDAY_NAMES } from '../ui.js';

// Rascunho do formulário: sobrevive aos re-renders dos botões de escolha.
let draft = null;

const FREQS = [['daily', 'Todo dia'], ['days', 'Dias fixos'], ['weekly', 'X por semana']];

export async function view(ctx) {
  const id = ctx.params.id;
  const key = id ?? 'novo';
  if (!draft || draft.key !== key) {
    draft = { key, name: '', effort: 120, kind: 'daily', days: [1, 3, 5], times: 3, archived: false };
    if (id) {
      const d = await api.get(`habits/${id}`);
      const s = d.current.schedule;
      Object.assign(draft, {
        name: d.habit.name, effort: d.habit.effort, kind: s.kind, archived: d.current.archived,
        days: s.days ?? draft.days, times: s.times ?? draft.times,
      });
    }
  }
  const f = draft;
  const back = id ? `#/habito?id=${id}` : '#/habitos';
  const body = () => ({
    name: f.name, effort: f.effort, archived: f.archived,
    schedule: f.kind === 'daily' ? { kind: 'daily' } : f.kind === 'days' ? { kind: 'days', days: [...f.days].sort() } : { kind: 'weekly', times: f.times },
  });
  const leave = (to) => { draft = null; ctx.go(to); };

  return {
    html: html`${top(id ? 'Editar hábito' : 'Novo hábito', back)}
    <form class="content pad stack" style="padding-top:14px" data-submit="save">
      <div><label class="lab" for="nome">NOME</label>
        <input id="nome" class="inp" type="text" maxlength="60" placeholder="Ex.: Ler 20 páginas" value="${f.name}" data-change="name" autocomplete="off"></div>
      <div><span class="lab">FREQUÊNCIA</span>
        <div class="seg" role="group" aria-label="Frequência">${FREQS.map(([k, t]) =>
          html`<button type="button" aria-pressed="${f.kind === k}" data-click="kind" data-v="${k}">${t}</button>`)}</div>
        ${f.kind === 'days' ? html`<div class="days" role="group" aria-label="Dias da semana" style="margin-top:6px">${WEEKDAYS.map((t, i) =>
          html`<button type="button" aria-pressed="${f.days.includes(i)}" aria-label="${WEEKDAY_NAMES[i]}" data-click="day" data-v="${i}">${t}</button>`)}</div>` : ''}
        ${f.kind === 'weekly' ? html`<div class="seg" role="group" aria-label="Vezes por semana" style="margin-top:6px">${[1, 2, 3, 4, 5, 6].map((n) =>
          html`<button type="button" aria-pressed="${f.times === n}" data-click="times" data-v="${n}">${n}x</button>`)}</div>` : ''}
      </div>
      <div><span class="lab">ESFORÇO</span>
        <div class="seg" role="group" aria-label="Esforço">${EFFORTS.map(([m, t]) =>
          html`<button type="button" class="pf" aria-pressed="${f.effort === m}" aria-label="${t}, ${m / 60}h de fôlego" data-click="effort" data-v="${m}">${t} +${m / 60}h</button>`)}</div>
        <p class="small muted" style="margin:6px 0 0">${f.kind === 'weekly'
          ? 'Cada vez feita dá esse fôlego. Cada vez que faltar na semana vira dívida no domingo.'
          : 'Cada dia feito dá esse fôlego ao esquadrão.'}</p>
      </div>
      <p class="err" id="err" role="alert"></p>
      <div class="stack" style="margin-top:auto">
        <button type="submit" class="btn">${id ? 'Salvar' : 'Criar hábito'}</button>
        ${id ? html`<div style="display:flex;gap:10px">
          <button type="button" class="ghost" data-click="archive">${f.archived ? 'Reativar' : 'Arquivar'}</button>
          <button type="button" class="ghost bad" data-click="remove">Apagar</button></div>` : ''}
      </div>
    </form>`,
    mount(root) { if (!id) root.querySelector('#nome').focus(); },
    actions: {
      name(el) { f.name = el.value; },
      kind(el) { f.kind = el.dataset.v; ctx.refresh(); },
      day(el) {
        const d = Number(el.dataset.v);
        f.days = f.days.includes(d) ? f.days.filter((x) => x !== d) : [...f.days, d];
        ctx.refresh();
      },
      times(el) { f.times = Number(el.dataset.v); ctx.refresh(); },
      effort(el) { f.effort = Number(el.dataset.v); ctx.refresh(); },
      async save(form) {
        f.name = form.querySelector('#nome').value;
        try {
          if (id) await api.put(`habits/${id}`, body());
          else await api.post('habits', body());
        } catch (e) {
          form.querySelector('#err').textContent = e.message;
          return;
        }
        leave(id ? back : '#/habitos');
      },
      async archive() {
        f.archived = !f.archived;
        await api.put(`habits/${id}`, body());
        leave(back);
      },
      async remove() {
        if (!confirm(`Apagar "${f.name}" e todo o histórico dele?`)) return;
        await api.del(`habits/${id}`);
        leave('#/habitos');
      },
    },
  };
}
