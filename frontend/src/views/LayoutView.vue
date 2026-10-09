<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import CommandPalette from '@/components/CommandPalette.vue'
import AppLogo from '@/components/AppLogo.vue'

const auth = useAuthStore()
const theme = useThemeStore()
const router = useRouter()
const route = useRoute()
const sidebarVisible = ref(false)
const commandPaletteVisible = ref(false)

type Nav = { label: string; icon: string; to: string; admin?: boolean }

const NAV: { section: string; items: Nav[] }[] = [
  { section: 'Operasional', items: [
    { label: 'Dashboard', icon: 'pi pi-th-large', to: '/' },
    { label: 'Barang', icon: 'pi pi-box', to: '/items' },
    { label: 'Log & Riwayat', icon: 'pi pi-history', to: '/history' },
  ]},
  { section: 'Master Data', items: [
    { label: 'Kategori', icon: 'pi pi-tags', to: '/categories' },
    { label: 'Lokasi', icon: 'pi pi-map-marker', to: '/locations' },
  ]},
  { section: 'Perangkat', items: [
    { label: 'QR Generator', icon: 'pi pi-qrcode', to: '/barcode' },
    { label: 'Opname Massal', icon: 'pi pi-clipboard', to: '/adjust/bulk' },
  ]},
  { section: 'Administrasi', items: [
    { label: 'Pengguna', icon: 'pi pi-users', to: '/users', admin: true },
    { label: 'Audit Log', icon: 'pi pi-shield', to: '/audit', admin: true },
    { label: 'Telegram', icon: 'pi pi-send', to: '/telegram', admin: true },
  ]},
]

const sections = computed(() =>
  NAV.map(s => ({ ...s, items: s.items.filter(i => !i.admin || auth.isAdmin) }))
    .filter(s => s.items.length)
)

const TITLES: Record<string, [string, string]> = {
  '/': ['Dashboard', 'Ringkasan stok dan pergerakan terkini'],
  '/items': ['Daftar Barang', 'Master inventaris kantor'],
  '/categories': ['Kategori', 'Pengelompokan jenis barang'],
  '/locations': ['Lokasi', 'Ruangan dan titik penyimpanan'],
  '/movement': ['Mutasi Barang', 'Catat barang masuk, keluar, dan opname'],
  '/history': ['Riwayat Mutasi', 'Jejak seluruh transaksi stok'],
  '/barcode': ['QR Generator', 'Cetak label QR untuk barang'],
  '/adjust/bulk': ['Opname Massal', 'Sesuaikan stok banyak barang via CSV'],
  '/users': ['Pengguna', 'Akun dan hak akses sistem'],
  '/audit': ['Audit Log', 'Jejak perubahan data oleh trigger basis data'],
  '/telegram': ['Telegram', 'Konfigurasi bot notifikasi Telegram'],
  '/password': ['Ganti Password', 'Perbarui kredensial akun Anda'],
}

const crumb = computed(() => {
  const found = Object.entries(TITLES).find(([p]) => route.path === p || route.path.startsWith(p + '/'))
  return found?.[1] ?? ['Inventaris Kantor', '']
})

function onKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    commandPaletteVisible.value = !commandPaletteVisible.value
  }
}

onMounted(() => {
  theme.apply()
  window.addEventListener('keydown', onKeydown)
  if (auth.isLoggedIn && !auth.user) auth.fetchMe()
})
onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
})

function logout() {
  auth.logout()
  router.push('/login')
}
function go(to: string) {
  router.push(to)
  sidebarVisible.value = false
}
const isActive = (to: string) => to === '/' ? route.path === '/' : route.path.startsWith(to)
</script>

<template>
  <div class="app-shell">
    <!-- ============ COMMAND PALETTE ============ -->
    <CommandPalette v-model:visible="commandPaletteVisible" />

    <!-- ============ SIDEBAR ============ -->
    <aside class="app-aside" :class="sidebarVisible ? 'is-open' : ''">
      <div class="app-brand cursor-pointer" @click="go('/')">
        <AppLogo />
      </div>

      <nav class="app-nav">
        <div v-for="sec in sections" :key="sec.section" class="nav-group">
          <div class="nav-section t-label">{{ sec.section }}</div>
          <button
            v-for="item in sec.items"
            :key="item.to"
            @click="go(item.to)"
            class="nav-item"
            :class="isActive(item.to) ? 'is-active' : ''"
          >
            <i :class="item.icon" class="nav-icon" />
            <span class="nav-label">{{ item.label }}</span>
          </button>
        </div>
      </nav>

      <div class="app-user">
        <div class="user-row">
          <div class="user-avatar">{{ auth.user?.name?.[0]?.toUpperCase() || '?' }}</div>
          <div class="user-meta">
            <div class="user-name">{{ auth.user?.name || '—' }}</div>
            <div class="t-label user-role">{{ auth.user?.role || '' }}</div>
          </div>
          <Button icon="pi pi-power-off" text rounded size="small" severity="danger"
                  v-tooltip.top="'Keluar'" @click="logout" />
        </div>
        <div class="user-actions">
          <Button icon="pi pi-key" text size="small" class="grow" v-tooltip.top="'Ganti password'"
                  @click="go('/password')" />
          <Button icon="pi pi-download" text size="small" class="grow"
                  v-tooltip.top="'Data Barang'" @click="go('/items')" />
        </div>
      </div>
    </aside>

    <div v-if="sidebarVisible" class="app-backdrop" @click="sidebarVisible = false" />

    <!-- ============ MAIN ============ -->
    <div class="app-main">
      <header class="app-topbar">
        <Button icon="pi pi-bars" text rounded size="small" class="topbar-burger"
                @click="sidebarVisible = !sidebarVisible" />
        <div class="topbar-title">
          <span class="topbar-page">{{ crumb[0] }}</span>
          <span class="topbar-sub hidden sm:inline">{{ crumb[1] }}</span>
        </div>

        <div class="topbar-right">
          <!-- Spotlight trigger button -->
          <button class="topbar-search" @click="commandPaletteVisible = true">
            <i class="pi pi-search text-[12px]" />
            <span class="text-[12px] hidden md:inline">Cari menu, barang…</span>
            <kbd class="t-mono text-[10px] px-1.5 py-0.5 rounded border ml-2 hidden sm:inline-block">⌘K</kbd>
          </button>

          <!-- Theme toggle button -->
          <Button
            :icon="theme.isDark ? 'pi pi-sun' : 'pi pi-moon'"
            text
            rounded
            size="small"
            severity="secondary"
            :v-tooltip.top="theme.isDark ? 'Ganti ke Mode Terang' : 'Ganti ke Mode Gelap'"
            @click="theme.toggle()"
          />

          <Tag severity="success" value="ONLINE" class="hidden sm:inline-flex" />
        </div>
      </header>

      <main class="app-body">
        <div class="app-container">
          <RouterView />
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.app-shell {
  display: flex;
  min-height: 100vh;
  background: var(--ink-950);
  color: var(--txt);
}

/* ---------- sidebar ---------- */
.app-aside {
  position: fixed;
  top: 0;
  left: 0;
  z-index: 40;
  display: flex;
  flex-direction: column;
  width: 250px;
  height: 100vh;
  flex-shrink: 0;
  background: var(--panel);
  border-right: 1px solid var(--line);
  transform: translateX(-100%);
  transition: transform 180ms cubic-bezier(0.4, 0, 0.2, 1);
}
.app-aside.is-open { transform: translateX(0); }

.app-brand {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 58px;
  padding: 0 18px;
  border-bottom: 1px solid var(--line);
}
.brand-mark {
  width: 32px; height: 32px;
  display: grid; place-items: center;
  border-radius: 8px;
  background: var(--acc-500);
  color: #ffffff;
  font-weight: 800;
  box-shadow: 0 2px 6px -1px rgba(79, 70, 229, 0.4);
}
.brand-text { min-width: 0; line-height: 1.15; }
.brand-name { font-size: 13px; font-weight: 800; letter-spacing: -0.01em; color: var(--txt); }
.brand-sub {
  font-size: 9.5px; font-weight: 700; letter-spacing: 0.08em;
  text-transform: uppercase; color: var(--txt-dim);
}

.app-nav { flex: 1; overflow-y: auto; padding: 14px 12px; }
.nav-group { margin-bottom: 18px; }
.nav-section { padding: 0 10px; margin-bottom: 6px; font-size: 10px; color: var(--txt-dim); font-weight: 700; text-transform: uppercase; letter-spacing: 0.08em; }

.nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 8px 12px;
  margin-bottom: 2px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: var(--txt-dim);
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
  transition: background 150ms ease, color 150ms ease;
}
.nav-item:hover {
  background: var(--panel-2);
  color: var(--txt);
}
.nav-item.is-active {
  background: var(--panel-2);
  color: var(--acc-500);
  font-weight: 600;
}
.nav-item.is-active::before {
  content: "";
  position: absolute;
  left: 0; top: 6px; bottom: 6px;
  width: 3px;
  border-radius: 0 4px 4px 0;
  background: var(--acc-500);
}
.nav-icon { width: 16px; text-align: center; font-size: 13px; flex-shrink: 0; }
.nav-label { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.app-user {
  padding: 12px;
  border-top: 1px solid var(--line);
  background: var(--panel-2);
}
.user-row { display: flex; align-items: center; gap: 10px; padding: 2px 4px; }
.user-avatar {
  width: 32px; height: 32px; flex-shrink: 0;
  display: grid; place-items: center;
  border-radius: 8px;
  border: 1px solid var(--line);
  background: var(--panel);
  color: var(--acc-500);
  font-size: 12px; font-weight: 700;
}
.user-meta { flex: 1; min-width: 0; line-height: 1.2; }
.user-name {
  font-size: 12.5px; font-weight: 600; color: var(--txt);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.user-role { font-size: 10px; text-transform: uppercase; color: var(--txt-dim); font-weight: 600; }
.user-actions { display: flex; gap: 6px; margin-top: 8px; }
.grow { flex: 1; }

.app-backdrop {
  position: fixed;
  inset: 0;
  z-index: 30;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(2px);
}

/* ---------- main column ---------- */
.app-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  height: 100vh;
  margin-left: 0;
}

.app-topbar {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 58px;
  flex-shrink: 0;
  padding: 0 24px;
  background: var(--panel);
  border-bottom: 1px solid var(--line);
}
.topbar-burger { display: inline-flex; }
.topbar-title { display: flex; align-items: baseline; gap: 10px; min-width: 0; }
.topbar-page {
  font-size: 14px; font-weight: 700; color: var(--txt);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.topbar-sub {
  font-size: 12px; color: var(--txt-dim);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.topbar-right { margin-left: auto; display: flex; align-items: center; gap: 10px; }

.topbar-search {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: 8px;
  border: 1px solid var(--line);
  background: var(--panel-2);
  color: var(--txt-dim);
  cursor: pointer;
  transition: all 150ms ease;
}
.topbar-search:hover {
  border-color: var(--acc-500);
  color: var(--txt);
}
.topbar-search kbd {
  border-color: var(--line);
  background: var(--panel);
}

.app-body { flex: 1; overflow-y: auto; background: var(--ink-950); }
.app-container { max-width: 1600px; padding: 24px 28px 40px; margin: 0 auto; }

@media (min-width: 1024px) {
  .app-aside { transform: translateX(0); }
  .app-main { margin-left: 250px; }
  .app-backdrop { display: none; }
  .topbar-burger { display: none; }
}
</style>