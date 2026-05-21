
window.API_BASE = '/api/v1';

window.api = async function(path, options = {}) {
    const token = localStorage.getItem('token');

    const headers = {
        'Content-Type': 'application/json',
        ...(options.headers || {})
    };

    if (token) {
        headers['Authorization'] = 'Bearer ' + token;
    }

    const response = await fetch(API_BASE + path, {
        credentials: 'include',
        ...options,
        headers
    });

    return response.json();
};
