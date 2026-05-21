// 共用 navbar Alpine.js component
function navbar() {
  return {
    loggedIn: false,
    memberName: '',
    init() {
      this.loggedIn = !!localStorage.getItem('access_token');
      const member = localStorage.getItem('member_info');
      if (member) {
        try { this.memberName = JSON.parse(member).name || '會員'; } catch(e) {}
      }
    },
    logout() {
      localStorage.removeItem('access_token');
      localStorage.removeItem('member_info');
      window.location.href = '/';
    }
  }
}
