import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  const saved = typeof localStorage !== 'undefined' ? localStorage.getItem('theme') : null
  const isDark = ref(saved ? saved !== 'light' : true)

  function toggle() {
    isDark.value = !isDark.value
    try {
      localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
    } catch {
      // ignore storage errors in restricted contexts
    }
    apply()
  }

  function apply() {
    if (isDark.value) {
      document.body.classList.add('dark')
      document.documentElement.classList.add('dark')
      document.documentElement.style.colorScheme = 'dark'
    } else {
      document.body.classList.remove('dark')
      document.documentElement.classList.remove('dark')
      document.documentElement.style.colorScheme = 'light'
    }
  }

  return { isDark, toggle, apply }
})
