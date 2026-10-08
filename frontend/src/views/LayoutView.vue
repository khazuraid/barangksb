<script setup lang="ts">
import { RouterView, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const menu = [
  { path: '/', label: 'Dashboard', icon: '📊' },
  { path: '/items', label: 'Barang', icon: '📦' },
  { path: '/categories', label: 'Kategori', icon: '🏷️' },
  { path: '/locations', label: 'Lokasi', icon: '📍' },
  { path: '/movement', label: 'Mutasi', icon: '🔄' },
  { path: '/history', label: 'Riwayat', icon: '🕐' },
  { path: '/barcode', label: 'QR Code', icon: '📱' },
]

const adminMenu = [
  { path: '/users', label: 'Pengguna', icon: '👥' },
  { path: '/audit', label: 'Audit', icon: '🛡️' },
]

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="drawer lg:drawer-open">
    <input id="sidebar" type="checkbox" class="drawer-toggle" />
    <div class="drawer-content flex flex-col">
      <!-- Topbar -->
      <div class="navbar bg-base-100 border-b border-base-200 sticky top-0 z-10 px-4">
        <label for="sidebar" class="btn btn-square btn-ghost lg:hidden">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" /></svg>
        </label>
        <h1 class="text-lg font-bold ml-2">{{ $route.meta.title || 'Inventaris Kantor' }}</h1>
        <div class="ml-auto flex items-center gap-2">
          <div class="avatar avatar-placeholder">
            <div class="bg-primary text-primary-content w-8 rounded-full">
              <span class="text-xs">{{ auth.user?.name?.[0] || '?' }}</span>
            </div>
          </div>
          <button class="btn btn-ghost btn-sm" @click="logout">Keluar</button>
        </div>
      </div>
      <!-- Content -->
      <main class="flex-1 p-4 lg:p-6 max-w-full">
        <RouterView />
      </main>
    </div>
    <!-- Sidebar -->
    <div class="drawer-side">
      <label for="sidebar" class="drawer-overlay"></label>
      <aside class="min-h-full w-64 bg-base-100 border-r border-base-200">
        <div class="p-4 border-b border-base-200">
          <div class="flex items-center gap-2">
            <div class="w-10 h-10 rounded-xl bg-primary text-primary-content flex items-center justify-center font-bold">IK</div>
            <div>
              <div class="font-bold text-sm">Inventaris</div>
              <div class="text-xs text-base-content/50">Kantor</div>
            </div>
          </div>
        </div>
        <nav class="p-2">
          <div class="text-xs font-bold text-base-content/40 px-3 py-2">MENU</div>
          <RouterLink v-for="m in menu" :key="m.path" :to="m.path"
            class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm hover:bg-base-200 transition"
            :class="{ 'bg-primary text-primary-content': $route.path === m.path }">
            <span>{{ m.icon }}</span> {{ m.label }}
          </RouterLink>
          <template v-if="auth.isAdmin">
            <div class="text-xs font-bold text-base-content/40 px-3 py-2 mt-2">ADMIN</div>
            <RouterLink v-for="m in adminMenu" :key="m.path" :to="m.path"
              class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm hover:bg-base-200 transition"
              :class="{ 'bg-primary text-primary-content': $route.path === m.path }">
              <span>{{ m.icon }}</span> {{ m.label }}
            </RouterLink>
          </template>
        </nav>
        <div class="p-3 border-t border-base-200 mt-auto">
          <div class="flex items-center gap-2">
            <div class="avatar avatar-placeholder">
              <div class="bg-neutral text-neutral-content w-8 rounded-full">
                <span class="text-xs">{{ auth.user?.name?.[0] || '?' }}</span>
              </div>
            </div>
            <div class="min-w-0">
              <div class="text-sm font-medium truncate">{{ auth.user?.name }}</div>
              <div class="text-xs text-base-content/50 capitalize">{{ auth.user?.role }}</div>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>
