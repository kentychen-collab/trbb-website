import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' }
})

// Request interceptor – 自動帶入 JWT token
api.interceptors.request.use(config => {
  const token = localStorage.getItem('trbb_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor – auto handle 401
api.interceptors.response.use(
  res => res,
  err => {
    if (err.response?.status === 401) {
      localStorage.removeItem('trbb_token')
      localStorage.removeItem('trbb_user')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export const thirdApi = axios.create({
  baseURL: import.meta.env.VITE_THIRD_BASE_URL,
  timeout: 30000,
})

export default api
