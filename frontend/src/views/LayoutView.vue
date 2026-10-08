<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Button from 'primevue/button'
import Tag from 'primevue/tag'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const sidebarVisible = ref(false)

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

// Whole app is a dark shell; PrimeVue's darkModeSelector points at body.dark.
onMounted(() => {
  document.body.classList.add('dark')
  if (auth.isLoggedIn && !auth.user) auth.fetchMe()
})
onUnmounted(() => document.body.classList.remove('dark'))

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
    <!-- ============ SIDEBAR ============ -->
    <aside class="app-aside" :class="sidebarVisible ? 'is-open' : ''">
      <div class="app-brand">
        <div class="brand-mark">IK</div>
        <div class="brand-text">
          <div class="brand-name">INVENTARIS</div>
          <div class="brand-sub">Kantor · v2</div>
        </div>
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
                  v-tooltip.top="'Unduh CSV barang'" @click="go('/items')" />
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
          <span class="topbar-sub">{{ crumb[1] }}</span>
        </div>
        <div class="topbar-right">
          <Tag severity="success" value="SISTEM AKTIF" icon="pi pi-circle-fill" />
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
/* Fixed px rhythm: 8px steps, 24px gutters, 56px chrome rows. */
.app-shell {
  display: flex;
  min-height: 100vh;
  background: var(--ink-950);
}

/* ---------- sidebar ---------- */
.app-aside {
  position: fixed;
  top: 0;
  left: 0;
  z-index: 40;
  display: flex;
  flex-direction: column;
  width: 248px;
  height: 100vh;
  flex-shrink: 0;
  background: var(--ink-900);
  border-right: 1px solid var(--ink-700);
  transform: translateX(-100%);
  transition: transform 160ms ease-out;
}
.app-aside.is-open { transform: translateX(0); }

.app-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 56px;
  padding: 0 16px;
  border-bottom: 1px solid var(--ink-700);
}
.brand-mark {
  width: 32px; height: 32px;
  display: grid; place-items: center;
  border-radius: 6px;
  background: var(--acc-500);
  color: var(--ink-950);
  font-size: 12px; font-weight: 800;
}
.brand-text { min-width: 0; line-height: 1.15; }
.brand-name { font-size: 12.5px; font-weight: 700; color: var(--ink-100); }
.brand-sub {
  font-size: 9.5px; font-weight: 700; letter-spacing: 0.09em;
  text-transform: uppercase; color: var(--ink-500);
}

.app-nav { flex: 1; overflow-y: auto; padding: 12px 10px; }
.nav-group { margin-bottom: 16px; }
.nav-section { padding: 0 10px; margin-bottom: 6px; font-size: 9.5px; }

.nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 8px 8px 8px 10px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--ink-300);
  font-family: inherit;
  font-size: 12.5px;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
  transition: background 160ms ease-out, color 160ms ease-out;
}
.nav-item:hover { background: var(--ink-850); color: var(--ink-100); }
.nav-item.is-active {
  background: var(--ink-800);
  color: var(--acc-300);
  font-weight: 600;
}
.nav-item.is-active::before {
  content: "";
  position: absolute;
  left: 0; top: 6px; bottom: 6px;
  width: 2.5px;
  border-radius: 0 2px 2px 0;
  background: var(--acc-500);
}
.nav-icon { width: 16px; text-align: center; font-size: 13px; flex-shrink: 0; }
.nav-label { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.app-user { padding: 10px; border-top: 1px solid var(--ink-700); }
.user-row { display: flex; align-items: center; gap: 10px; padding: 4px 6px; }
.user-avatar {
  width: 32px; height: 32px; flex-shrink: 0;
  display: grid; place-items: center;
  border-radius: 6px;
  border: 1px solid var(--ink-700);
  background: var(--ink-800);
  color: var(--acc-300);
  font-size: 12px; font-weight: 700;
}
.user-meta { flex: 1; min-width: 0; line-height: 1.2; }
.user-name {
  font-size: 12px; font-weight: 600; color: var(--ink-100);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.user-role { font-size: 9.5px; }
.user-actions { display: flex; gap: 6px; margin-top: 8px; }
.grow { flex: 1; }

.app-backdrop {
  position: fixed;
  inset: 0;
  z-index: 30;
  background: rgb(0 0 0 / 0.5);
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
  height: 56px;
  flex-shrink: 0;
  padding: 0 24px;
  background: var(--ink-850);
  border-bottom: 1px solid var(--ink-800);
}
.topbar-burger { display: inline-flex; }
.topbar-title { display: flex; align-items: baseline; gap: 10px; min-width: 0; }
.topbar-page {
  font-size: 13.5px; font-weight: 700; color: var(--ink-100);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.topbar-sub {
  font-size: 11.5px; color: var(--ink-400);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.topbar-right { margin-left: auto; display: flex; align-items: center; gap: 8px; }
.topbar-status i { font-size: 6px; }

.app-body { flex: 1; overflow-y: auto; }
.app-container { max-width: 1600px; padding: 20px 24px 32px; }

@media (min-width: 1024px) {
  .app-aside { transform: translateX(0); }
  .app-main { margin-left: 248px; }
  .app-backdrop { display: none; }
  .topbar-burger { display: none; }
}
</style>