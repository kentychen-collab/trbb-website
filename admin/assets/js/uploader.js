// 共用圖片上傳元件
// 用法: const url = await uploadImage(file, 'products')
async function uploadImage(file, type = 'images') {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('type', type);

  const token = localStorage.getItem('admin_token');
  const res = await fetch('/admin-api/v1/upload/image', {
    method: 'POST',
    headers: token ? { 'Authorization': 'Bearer ' + token } : {},
    body: formData,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || '上傳失敗');
  }
  const data = await res.json();
  return data.url;
}

// Alpine.js 圖片上傳元件（單張封面）
function imageUploader(initialUrl = '', type = 'images') {
  return {
    url: initialUrl,
    uploading: false,
    error: '',
    async pick(event) {
      const file = event.target.files[0];
      if (!file) return;
      this.uploading = true;
      this.error = '';
      try {
        this.url = await uploadImage(file, type);
      } catch(e) {
        this.error = e.message;
      }
      this.uploading = false;
    },
    clear() { this.url = ''; }
  };
}
