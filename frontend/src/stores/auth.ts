import { defineStore } from 'pinia'
import api from '@/api'

interface User {
  id: string
  name: string
  email: string
  role: string
}

let cachedUser: User | null = null
try {
  const raw = localStorage.getItem('user')
  if (raw) cachedUser = JSON.parse(raw)
} catch {}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: cachedUser,
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
    isAdmin: (state) => state.user?.role === 'admin',
  },
  actions: {
    async login(email: string, password: string) {
      const res = await api.post('/auth/login', { email, password })
      this.token = res.data.token
      this.user = res.data.user
      localStorage.setItem('token', this.token)
      localStorage.setItem('user', JSON.stringify(this.user))
    },
    async fetchMe() {
      try {
        const res = await api.get('/auth/me')
        this.user = res.data
        localStorage.setItem('user', JSON.stringify(this.user))
      } catch {
        this.logout()
      }
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    },
  },
})
