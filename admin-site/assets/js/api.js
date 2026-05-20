const ADMIN_API_BASE = '/admin-api/v1';

const adminApi = {
  async request(method, path, data = null) {
    const token = localStorage.getItem('admin_token');
    const options = {
      method,
      headers: { 'Content-Type': 'application/json' },
    };
    if (token) options.headers['Authorization'] = `Bearer ${token}`;
    if (data) options.body = JSON.stringify(data);
    try {
      const res = await fetch(ADMIN_API_BASE + path, options);
      if (res.status === 401) {
        localStorage.removeItem('admin_token');
        localStorage.removeItem('admin_info');
        window.location.href = '/admin/login.html';
        return null;
      }
      return await res.json();
    } catch (e) {
      console.error('Admin API Error:', e);
      throw e;
    }
  },
  get:    (path)       => adminApi.request('GET', path),
  post:   (path, data) => adminApi.request('POST', path, data),
  put:    (path, data) => adminApi.request('PUT', path, data),
  delete: (path)       => adminApi.request('DELETE', path),
};

const adminAuth = {
  isLoggedIn()  { return !!localStorage.getItem('admin_token'); },
  getAdmin()    { try { return JSON.parse(localStorage.getItem('admin_info')); } catch(e) { return null; } },
  isSuperAdmin() { const a = this.getAdmin(); return a && a.role >= 2; },
  hasPermission(perm) {
    const a = this.getAdmin();
    if (!a) return false;
    if (a.role >= 2) return true;
    return a.permissions && a.permissions[perm];
  },
  requireLogin() { if (!this.isLoggedIn()) window.location.href = '/admin/login.html'; },
  logout() {
    localStorage.removeItem('admin_token');
    localStorage.removeItem('admin_info');
    window.location.href = '/admin/login.html';
  }
};
