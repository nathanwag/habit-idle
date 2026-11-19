async function call(method, path, body) {
  const res = await fetch(`api/${path}`, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    let msg = `Erro ${res.status}`;
    try { msg = (await res.json()).error ?? msg; } catch { /* corpo sem JSON */ }
    throw new Error(msg);
  }
  return res.status === 204 ? null : res.json();
}

export const api = {
  get: (p) => call('GET', p),
  post: (p, b) => call('POST', p, b),
  put: (p, b) => call('PUT', p, b),
  del: (p) => call('DELETE', p),
};
