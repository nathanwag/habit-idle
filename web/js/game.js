import { html } from './ui.js';

export const STATS = [
  { key: 'atk', label: 'ATK', name: 'Ataque', color: 'var(--atk)' },
  { key: 'def', label: 'DEF', name: 'Defesa', color: 'var(--def)' },
  { key: 'hp', label: 'HP', name: 'Vida', color: 'var(--hp)' },
  { key: 'vel', label: 'VEL', name: 'Velocidade', color: 'var(--vel)' },
  { key: 'sor', label: 'SORTE', name: 'Sorte', color: 'var(--sor)' },
];

export const BLURB = {
  oren: 'Bate mais forte. Bom para vencer chefes.',
  kael: 'Segura os golpes na frente do esquadrão.',
  brom: 'Muita vida: aguenta quando a torre aperta.',
  rasko: 'Ataca rápido, golpe atrás de golpe.',
  lyra: 'Vive de crítico: golpes que valem o dobro.',
};

export const statOf = (key) => STATS.find((s) => s.key === key);

export const sprite = (id, size = 'x3', mode = 'idle') => html`<span class="spr h-${id} ${size} ${mode}" aria-hidden="true"></span>`;

export const coin = html`<span class="coin" aria-hidden="true"></span>`;

export const num = (n) => Math.round(n).toLocaleString('pt-BR');

export function hm(minutes) {
  const neg = minutes < 0;
  const m = Math.abs(Math.round(minutes));
  return `${neg ? '−' : ''}${Math.floor(m / 60)}h${String(m % 60).padStart(2, '0')}`;
}

export const clock = (iso) => new Date(iso).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });

export const RARITY = ['comum', 'rara', 'épica', 'lendária'];
export const SLOTS = [['arma', 'Arma'], ['armadura', 'Armadura'], ['amuleto', 'Amuleto']];

export function itemDesc(it) {
  return STATS.filter((s, i) => it.pts[i] > 0).map((s) => `${s.label} +${it.pts[STATS.indexOf(s)]}`).join(' · ');
}

// Cena do andar com o que está acontecendo de verdade (g.enc): caminhando,
// lutando contra o inimigo atual ou descansando. Sem fôlego, todos parados.
export function battleScene(g, exploring, { height = 270, bottom = 52, top = '' } = {}) {
  const squad = g.squad.map((id) => g.heroes.find((h) => h.id === id));
  const enc = g.enc;
  const phase = exploring ? enc.phase : 'stop';
  // squad[0] vai na frente: é quem apanha.
  const xs = [128, 74, 22];
  const lifePct = (h) => Math.max(0, Math.round(100 * (enc.life[h.id] ?? 1) / (enc.maxLife[h.id] || 1)));
  const lifeBar = (h) => {
    const pct = lifePct(h);
    return html`<span class="life ${pct < 30 ? 'low' : ''}" aria-hidden="true"><span style="width:${pct}%"></span></span>`;
  };
  const tag = { walk: 'caminhando', fight: `inimigo ${enc.enemy}/${enc.of}`, rest: 'descansando', stop: 'sem fôlego' }[phase];
  const heroes = squad.map((h, i) => {
    const down = lifePct(h) <= 0;
    const mode = phase === 'walk' ? 'run' : 'idle';
    return html`<div class="fh ${phase === 'fight' && i === 0 && !down ? 'atk' : ''}" style="left:${xs[i]}px;bottom:${bottom + (i === 1 ? -8 : 0)}px;${down ? 'opacity:.35;filter:grayscale(1)' : ''}">
      ${phase === 'stop' ? '' : lifeBar(h)}${sprite(h.id, 'x3', mode)}</div>`;
  });
  const k = 1 - g.fatigue / 100;
  const front = squad.find((h) => lifePct(h) > 0) ?? squad[0];
  const foe = phase === 'fight' ? html`<div class="foe" style="left:186px;bottom:${bottom - 10}px">
      <div class="hpb"><span style="width:${Math.max(0, Math.round(100 * enc.enemyLife / enc.enemyMax))}%"></span></div><div class="fb"></div></div>
    <span class="dmg" style="left:220px;bottom:${bottom + 80}px">-${Math.round(4 * front.points[0] * k)}</span>` : '';
  const label = {
    walk: 'O esquadrão caminha pelo andar',
    fight: `O esquadrão luta contra o inimigo ${enc.enemy} de ${enc.of}`,
    rest: 'O esquadrão descansa depois de cair',
    stop: 'O esquadrão descansa: sem fôlego',
  }[phase];
  const zz = phase === 'rest' || phase === 'stop'
    ? html`<span class="zz" style="left:${xs[0] + 30}px;bottom:${bottom + 80}px">z</span><span class="zz" style="left:${xs[0] + 50}px;bottom:${bottom + 90}px;animation-delay:1.2s">z</span>`
    : '';
  return html`<div class="scn ${phase === 'walk' ? 'go' : phase === 'fight' ? '' : 'rest'}" role="img" aria-label="${label}" style="height:${height}px">${top}
    ${heroes}${foe}${zz}<span class="state-tag">${tag}</span>
  </div>`;
}
