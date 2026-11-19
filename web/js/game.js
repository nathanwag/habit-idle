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

// Cena da torre. Explorando: o ciclo de luta do protótipo (só visual, com o
// dano de verdade de cada herói). Parado: o esquadrão descansa.
export function battleScene(g, exploring, { height = 270, bottom = 52, top = '' } = {}) {
  const squad = g.squad.map((id) => g.heroes.find((h) => h.id === id));
  // squad[0] vai na frente (é quem apanha do chefe).
  const xs = [128, 74, 22];
  if (!exploring) {
    return html`<div class="scn rest" role="img" aria-label="O esquadrão descansa: sem fôlego" style="height:${height}px">${top}
      ${squad.map((h, i) => html`<span class="fh" style="left:${xs[i]}px;bottom:${bottom}px">${sprite(h.id, 'x3', 'idle')}</span>`)}
      <span class="zz" style="left:${xs[0] + 30}px;bottom:${bottom + 80}px">z</span><span class="zz" style="left:${xs[0] + 50}px;bottom:${bottom + 90}px;animation-delay:1.2s">z</span>
    </div>`;
  }
  const k = 1 - g.fatigue / 100;
  const dmg = (h) => Math.round(4 * h.points[0] * k);
  const hit = (i) => dmg(squad[i % squad.length]);
  return html`<div class="scn fight" role="img" aria-label="O esquadrão luta no andar ${g.floor}" style="height:${height}px">${top}
    ${squad.map((h, i) => html`<div class="fh ${i === 0 ? 'lunge' : ''}" style="left:${xs[i]}px;bottom:${bottom + (i === 1 ? -8 : 0)}px">
      <div class="${i === 0 ? 'hurt' : ''}" style="position:absolute;inset:0"><i class="r h-${h.id}"></i><i class="d h-${h.id}"></i></div></div>`)}
    <div class="foe" style="left:186px;bottom:${bottom - 10}px"><div class="hpb"><span></span></div><div class="fb"><i class="r"></i><i class="d"></i></div></div>
    <span class="dmg n1" style="left:214px;bottom:${bottom + 70}px">-${hit(0)}</span>
    <span class="dmg n2" style="left:236px;bottom:${bottom + 62}px;color:var(--def)">-${hit(1)}</span>
    <span class="dmg n3" style="left:222px;bottom:${bottom + 74}px">-${hit(2) * 2}!</span>
    <span class="dmg lt" style="left:204px;bottom:${bottom + 30}px">+${10 + 3 * g.floor} ${coin}</span>
  </div>`;
}
