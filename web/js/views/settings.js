import { api } from '../api.js';
import { html, nav, toast } from '../ui.js';

export async function view(ctx) {
  const s = await api.get('settings');
  return {
    html: html`<div class="content">
      <div class="head"><h1 class="title">Ajustes</h1></div>
      <div class="pad stack">
        <div><div class="sec">DIA</div>
          <div class="card" style="padding:12px">
            <label class="lab" for="virada">VIRADA DO DIA</label>
            <input id="virada" class="inp" type="time" value="${s.dayStart}" data-change="dayStart">
            <p class="small muted" style="margin:6px 0 0">Antes desse horário a madrugada ainda conta como o dia anterior. Na virada, metade do fôlego que sobrou passa para o dia seguinte.</p>
          </div></div>
        <div class="card small muted" style="padding:12px;line-height:1.6">
          <div style="display:flex;justify-content:space-between"><span class="pf" style="font-size:15px;color:var(--text)">habit-idle</span><span>protótipo</span></div>
          <div>Arte pixel: 0x72 DungeonTileset II (CC0). Paleta Apollo.</div>
        </div>
      </div></div>${nav('#/ajustes')}`,
    actions: {
      async dayStart(el) {
        await api.put('settings', { dayStart: el.value });
        toast('Virada salva');
        ctx.refresh();
      },
    },
  };
}
