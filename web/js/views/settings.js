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
        <div><div class="sec">DADOS</div>
          <div class="card">
            <a class="row" href="api/backup" download><span style="flex:1"><span style="display:block">Exportar backup</span><span class="small muted">hábitos, histórico e jogo num arquivo .json</span></span><span class="pf muted">›</span></a>
            <label class="row" for="importar" style="cursor:pointer"><span style="flex:1"><span style="display:block">Importar backup</span><span class="small muted">substitui tudo o que está aqui</span></span><span class="pf muted">›</span></label>
            <input id="importar" type="file" accept="application/json,.json" data-change="restore" style="position:absolute;width:1px;height:1px;opacity:0">
            <button type="button" class="row" data-click="reset"><span style="flex:1;color:var(--bad)">Recomeçar a torre do zero</span><span class="pf muted">›</span></button>
          </div></div>
        <div class="card small muted" style="padding:12px;line-height:1.6">
          <div style="display:flex;justify-content:space-between"><span class="pf" style="font-size:15px;color:var(--text)">Rasko</span><span>protótipo</span></div>
          <div>Arte pixel: 0x72 DungeonTileset II (CC0). Paleta Apollo.</div>
        </div>
      </div></div>${nav('#/ajustes')}`,
    actions: {
      async dayStart(el) {
        await api.put('settings', { dayStart: el.value });
        toast('Virada salva');
        ctx.refresh();
      },
      async restore(el) {
        const file = el.files[0];
        if (!file || !confirm('Importar substitui todos os hábitos, o histórico e o jogo daqui. Continuar?')) return;
        await api.put('backup', JSON.parse(await file.text()));
        toast('Backup importado');
        ctx.go('#/');
      },
      async reset() {
        if (!confirm('Recomeçar a torre? Heróis, moedas e itens somem; hábitos e histórico ficam.')) return;
        await api.del('game');
        ctx.go('#/inicio');
      },
    },
  };
}
