function authHeaders() {
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function request(method, path, body) {
  const res = await fetch(path, {
    method,
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  if (res.status === 401) {
    localStorage.removeItem('token');
    localStorage.removeItem('email');
    window.location.href = '/login';
    return res;
  }
  return res;
}

export const api = {
  register: (email, password) => request('POST', '/auth/register', { email, password }),
  login:    (email, password) => request('POST', '/auth/login', { email, password }),
  getLinks: ()                => request('GET',  '/api/links'),
  createLink: (url)           => request('POST', '/api/links', { url }),
  deleteLink: (code)          => request('DELETE', `/api/links/${code}`),
};
