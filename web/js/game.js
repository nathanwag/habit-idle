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
