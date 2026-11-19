import { html, nav } from '../ui.js';

export async function view(ctx) {
  return {
    html: html`<div class="content"><div class="head"><h1 class="title">${ctx.path === '/torre' ? 'Torre' : 'Heróis'}</h1></div>
      <p class="empty">Em construção.</p></div>${nav('#' + ctx.path)}`,
  };
}
