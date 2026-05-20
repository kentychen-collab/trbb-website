// 統一 API 請求封裝
const API_BASE = '/api/v1';

const api = {
  async request(method, path, data = null) {
    const token = localStorage.getItem('access_token');
    const options = {
      method,
      headers: { 'Content-Type': 'application/json' },
    };
    if (token) options.headers['Authorization'] = `Bearer ${token}`;
    if (data) options.body = JSON.stringify(data);

    try {
      const res = await fetch(API_BASE + path, options);
      if (res.status === 401) {
        localStorage.removeItem('access_token');
        window.location.href = '/member/login.html';
        return;
      }
      return await res.json();
    } catch (e) {
      console.error('API Error:', e);
      throw e;
    }
  },
  get: (path) => api.request('GET', path),
  post: (path, data) => api.request('POST', path, data),
  put: (path, data) => api.request('PUT', path, data),
  delete: (path) => api.request('DELETE', path),
};
