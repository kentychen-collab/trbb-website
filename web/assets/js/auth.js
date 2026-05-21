// JWT Auth 管理
const auth = {
  isLoggedIn() {
    return !!localStorage.getItem('access_token');
  },
  getToken() {
    return localStorage.getItem('access_token');
  },
  setToken(token) {
    localStorage.setItem('access_token', token);
  },
  logout() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('member_info');
    window.location.href = '/member/login.html';
  },
  getMember() {
    const info = localStorage.getItem('member_info');
    return info ? JSON.parse(info) : null;
  },
  setMember(member) {
    localStorage.setItem('member_info', JSON.stringify(member));
  },
  requireLogin() {
    if (!this.isLoggedIn()) {
      window.location.href = '/member/login.html?redirect=' + encodeURIComponent(location.href);
    }
  }
};
