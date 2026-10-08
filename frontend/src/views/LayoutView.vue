<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Button from 'primevue/button'
import Avatar from 'primevue/avatar'
import Menu from 'primevue/menu'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const sidebarVisible = ref(false)
const darkMode = ref(false)

const menuItems = [
  { separator: true },
  { label: 'MENU UTAMA', items: [
    { label: 'Dashboard', icon: 'pi pi-home', to: '/' },
    { label: 'Barang', icon: 'pi pi-box', to: '/items' },
    { label: 'Kategori', icon: 'pi pi-tags', to: '/categories' },
    { label: 'Lokasi', icon: 'pi pi-map-marker', to: '/locations' },
  ]},
  { separator: true },
  { label: 'TRANSAKSI', items: [
    { label: 'Mutasi', icon: 'pi pi-arrows-h', to: '/movement' },
    { label: 'Riwayat', icon: 'pi pi-clock', to: '/history' },
  ]},
  { separator: true },
  { label: 'ALAT', items: [
    { label: 'QR Generator', icon: 'pi pi-qrcode', to: '/barcode' },
    { label: 'Opname Massal', icon: 'pi pi-clipboard', to: '/adjust/bulk' },
  ]},
]

const adminItems = [
  { separator: true },
  { label: 'ADMIN', items: [
    { label: 'Pengguna', icon: 'pi pi-users', to: '/users' },
    { label: 'Audit Log', icon: 'pi pi-shield', to: '/audit' },
  ]},
]

const allItems = computed(() => {
  return auth.isAdmin ? [...menuItems, ...adminItems] : menuItems
})

const pageTitle = computed(() => {
  const map: Record<string, string> = {
    '/': 'Dashboard', '/items': 'Daftar Barang', '/categories': 'Kategori',
    '/locations': 'Lokasi', '/movement': 'Mutasi Barang', '/history': 'Riwayat',
    '/barcode': 'QR Generator', '/users': 'Pengguna', '/audit': 'Audit Log',
    '/password': 'Ganti Password', '/adjust/bulk': 'Opname Massal',
  }
  return map[route.path] || 'Inventaris Kantor'
})

function toggleDarkMode() {
  darkMode.value = !darkMode.value
  document.documentElement.classList.toggle('dark', darkMode.value)
}

function logout() {
  auth.logout()
  router.push('/login')
}

function navigate(to: string) {
  router.push(to)
  sidebarVisible.value = false
}

onMounted(() => {
  if (auth.isLoggedIn && !auth.user) {
    auth.fetchMe()
  }
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 flex">
    <!-- Sidebar -->
    <aside
      class="fixed lg:static inset-y-0 left-0 z-40 w-64 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 transform transition-transform duration-200"
      :class="{ '-translate-x-full lg:translate-x-0': !sidebarVisible, 'translate-x-0': sidebarVisible }"
    >
      <!-- Brand -->
      <div class="h-16 flex items-center gap-3 px-5 border-b border-gray-200 dark:border-gray-800">
        <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center text-white font-bold text-sm shadow-lg shadow-emerald-500/30">
          IK
        </div>
        <div>
          <div class="font-bold text-sm text-gray-800 dark:text-gray-100">Inventaris</div>
          <div class="text-xs text-gray-400">Kantor</div>
        </div>
      </div>

      <!-- Menu -->
      <nav class="p-3 overflow-y-auto h-[calc(100vh-8rem)]">
        <template v-for="(section, i) in allItems" :key="i">
          <div v-if="section.separator" class="border-t border-gray-100 dark:border-gray-800 my-2"></div>
          <div v-if="section.label" class="px-3 py-2 text-xs font-bold text-gray-400 uppercase tracking-wider">{{ section.label }}</div>
          <template v-if="section.items">
            <button
              v-for="item in section.items"
              :key="item.to"
              @click="navigate(item.to)"
              class="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-150"
              :class="route.path === item.to
                ? 'bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 border-l-4 border-emerald-500'
                : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 border-l-4 border-transparent'"
            >
              <i :class="item.icon" class="text-base"></i>
              {{ item.label }}
            </button>
          </template>
        </template>
      </nav>

      <!-- User -->
      <div class="absolute bottom-0 left-0 right-0 p-3 border-t border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
        <div class="flex items-center gap-3">
          <Avatar :label="auth.user?.name?.[0] || '?'" shape="circle" class="bg-emerald-500" />
          <div class="flex-1 min-w-0">
            <div class="text-sm font-semibold text-gray-800 dark:text-gray-100 truncate">{{ auth.user?.name }}</div>
            <div class="text-xs text-gray-400 capitalize">{{ auth.user?.role }}</div>
          </div>
          <Button icon="pi pi-sign-out" severity="danger" text rounded size="small" @click="logout" />
        </div>
      </div>
    </aside>

    <!-- Overlay mobile -->
    <div v-if="sidebarVisible" class="fixed inset-0 bg-black/30 z-30 lg:hidden" @click="sidebarVisible = false"></div>

    <!-- Main -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Topbar -->
      <header class="h-16 bg-white/80 dark:bg-gray-900/80 backdrop-blur border-b border-gray-200 dark:border-gray-800 flex items-center px-4 lg:px-6 sticky top-0 z-20">
        <Button icon="pi pi-bars" text rounded @click="sidebarVisible = !sidebarVisible" class="lg:hidden" />
        <h1 class="text-lg font-bold text-gray-800 dark:text-gray-100 ml-2 lg:ml-0">{{ pageTitle }}</h1>
        <div class="ml-auto flex items-center gap-2">
          <Button icon="pi pi-moon" text rounded @click="toggleDarkMode" />
          <Button icon="pi pi-key" text rounded @click="router.push('/password')" />
        </div>
      </header>

      <!-- Content -->
      <main class="flex-1 p-4 lg:p-6 overflow-y-auto">
        <RouterView />
      </main>
    </div>
  </div>
</template>
