import { defineStore } from 'pinia'
import api from '@/api'

interface User {
  id: string
  name: string
  email: string
  role: string
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: null as User | null,
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
    },
    async fetchMe() {
      const res = await api.get('/auth/me')
      this.user = res.data
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
    },
  },
})
