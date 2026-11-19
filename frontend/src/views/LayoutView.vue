<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import Button from 'primevue/button'
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
    { label: 'Mutasi Stok', icon: 'pi pi-arrow-right-arrow-left', to: '/movement' },
    { label: 'Pengaduan & Laporan', icon: 'pi pi-megaphone', to: '/reports' },
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
    { label: 'Kelola Foto', icon: 'pi pi-images', to: '/photos', admin: true },
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
  '/items': ['Daftar Barang', 'Master inventaris puskesmas'],
  '/categories': ['Kategori', 'Pengelompokan jenis barang'],
  '/locations': ['Lokasi', 'Ruangan dan titik penyimpanan'],
  '/movement': ['Mutasi Barang', 'Catat barang masuk, keluar, dan opname'],
  '/history': ['Riwayat Mutasi', 'Jejak seluruh transaksi stok'],
  '/reports': ['Pengaduan & Pelaporan Aset', 'Kelola laporan kerusakan dan pengaduan dari publik'],
  '/barcode': ['QR Generator', 'Cetak label QR untuk barang'],
  '/adjust/bulk': ['Opname Massal', 'Sesuaikan stok banyak barang via CSV'],
  '/users': ['Pengguna', 'Akun dan hak akses sistem'],
  '/photos': ['Kelola Foto', 'Deteksi foto yatim dan audit ruang penyimpanan foto'],
  '/audit': ['Audit Log', 'Jejak perubahan data oleh trigger basis data'],
  '/telegram': ['Telegram', 'Konfigurasi bot notifikasi Telegram'],
  '/password': ['Ganti Password', 'Perbarui kredensial akun Anda'],
}

const crumb = computed(() => {
  const found = Object.entries(TITLES).find(([p]) => route.path === p || route.path.startsWith(p + '/'))
  return found?.[1] ?? ['Inventaris Puskesmas', '']
})

function onKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    commandPaletteVisible.value = !commandPaletteVisible.value
  } else if (e.key === 'Escape' && sidebarVisible.value) {
    sidebarVisible.value = false
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
const isActive = (to: string) => to === '/' ? route.path === '/' : route.path === to || route.path.startsWith(to + '/')
</script>

<template>
  <a href="#main-content" class="skip-link">Lewati ke konten utama</a>

  <div class="app-shell">
    <CommandPalette v-model:visible="commandPaletteVisible" />

    <!-- ============ SIDEBAR ============ -->
    <aside class="app-aside" :class="{ 'is-open': sidebarVisible }" aria-label="Sidebar navigasi">
      <div class="app-brand">
        <AppLogo size="md" />
      </div>

      <nav class="app-nav" aria-label="Navigasi utama">
        <div v-for="sec in sections" :key="sec.section" class="nav-group">
          <div class="nav-section">{{ sec.section }}</div>
          <button
            v-for="item in sec.items"
            :key="item.to"
            type="button"
            class="nav-item"
            :class="{ 'is-active': isActive(item.to) }"
            :aria-current="isActive(item.to) ? 'page' : undefined"
            @click="go(item.to)"
          >
            <i :class="item.icon" class="nav-icon" aria-hidden="true" />
            <span class="nav-label">{{ item.label }}</span>
          </button>
        </div>
      </nav>

      <div class="app-user">
        <div class="user-row">
          <div class="user-avatar" aria-hidden="true">
            {{ auth.user?.name?.[0]?.toUpperCase() || '?' }}
          </div>
          <div class="user-meta">
            <div class="user-name">{{ auth.user?.name || '—' }}</div>
            <div class="user-role">{{ auth.user?.role || '' }}</div>
          </div>
          <Button
            icon="pi pi-sign-out"
            text
            rounded
            size="small"
            severity="danger"
            aria-label="Keluar dari akun"
            v-tooltip.top="'Keluar'"
            @click="logout"
          />
        </div>
        <div class="user-actions">
          <Button
            icon="pi pi-key"
            label="Password"
            text
            size="small"
            severity="secondary"
            class="grow !text-[11.5px] !py-1"
            @click="go('/password')"
          />
          <Button
            icon="pi pi-plus"
            label="Barang"
            text
            size="small"
            severity="secondary"
            class="grow !text-[11.5px] !py-1"
            @click="go('/items/new')"
          />
        </div>
      </div>
    </aside>

    <div
      v-if="sidebarVisible"
      class="app-backdrop"
      aria-hidden="true"
      @click="sidebarVisible = false"
    />

    <!-- ============ MAIN ============ -->
    <div class="app-main">
      <header class="app-topbar">
        <Button
          icon="pi pi-bars"
          text
          rounded
          size="small"
          severity="secondary"
          class="topbar-burger"
          aria-label="Buka menu navigasi"
          @click="sidebarVisible = !sidebarVisible"
        />

        <div class="topbar-title">
          <span class="topbar-page">{{ crumb[0] }}</span>
          <span v-if="crumb[1]" class="topbar-sep hidden md:inline" aria-hidden="true">/</span>
          <span class="topbar-sub hidden md:inline">{{ crumb[1] }}</span>
        </div>

        <div class="topbar-right">
          <button
            type="button"
            class="topbar-search"
            aria-label="Buka pencarian cepat (Command + K)"
            @click="commandPaletteVisible = true"
          >
            <i class="pi pi-search text-[11.5px]" aria-hidden="true" />
            <span class="text-[12px] hidden sm:inline">Cari barang atau menu…</span>
            <kbd class="t-mono !text-[10px] px-1.5 py-0.5 rounded border ml-1.5 hidden sm:inline-block">⌘K</kbd>
          </button>

          <Button
            :icon="theme.isDark ? 'pi pi-sun' : 'pi pi-moon'"
            text
            rounded
            size="small"
            severity="secondary"
            :aria-label="theme.isDark ? 'Ganti ke mode terang' : 'Ganti ke mode gelap'"
            v-tooltip.bottom="theme.isDark ? 'Mode Terang' : 'Mode Gelap'"
            @click="theme.toggle()"
          />

          <div class="topbar-status hidden sm:inline-flex" title="Koneksi server aktif">
            <span class="status-dot" aria-hidden="true" />
            <span>Aktif</span>
          </div>
        </div>
      </header>

      <main id="main-content" class="app-body" tabindex="-1">
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
  min-height: 100dvh;
  background: var(--ink-950);
  color: var(--txt);
}

/* ---------- Sidebar (Subtle Tinted Spine, Never Pitch-Black) ---------- */
.app-aside {
  position: fixed;
  top: 0;
  left: 0;
  z-index: var(--z-sidebar);
  display: flex;
  flex-direction: column;
  width: 236px;
  height: 100dvh;
  flex-shrink: 0;
  background: var(--sidebar-bg);
  border-right: 1px solid var(--line);
  transform: translateX(-100%);
  transition: transform 160ms var(--ease-out);
}
.app-aside.is-open { transform: translateX(0); }

.app-brand {
  display: flex;
  align-items: center;
  height: 52px;
  padding: 0 16px;
  border-bottom: 1px solid var(--line);
  flex-shrink: 0;
}

.app-nav {
  flex: 1;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 12px 10px;
}
.nav-group { margin-bottom: 16px; }
.nav-group:last-child { margin-bottom: 0; }
.nav-section {
  padding: 0 10px;
  margin-bottom: 5px;
  font-size: 10.5px;
  color: var(--txt-muted);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 9px;
  width: 100%;
  padding: 7px 10px;
  margin-bottom: 2px;
  border: 0;
  border-radius: 7px;
  background: transparent;
  color: var(--txt-dim);
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
  transition: background-color 110ms ease, color 110ms ease;
}
.nav-item:hover {
  background: var(--row-hover);
  color: var(--txt);
}
.nav-item.is-active {
  background: var(--acc-wash);
  color: var(--acc-500);
  font-weight: 600;
}
.nav-item.is-active::before {
  content: "";
  position: absolute;
  left: 0;
  top: 6px;
  bottom: 6px;
  width: 2.5px;
  border-radius: 0 3px 3px 0;
  background: var(--acc-500);
}
.nav-icon {
  width: 16px;
  text-align: center;
  font-size: 12.5px;
  flex-shrink: 0;
  opacity: 0.9;
}
.nav-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-user {
  padding: 10px;
  border-top: 1px solid var(--line);
  background: var(--panel);
  flex-shrink: 0;
}
.user-row {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 2px 4px;
}
.user-avatar {
  width: 30px;
  height: 30px;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  border-radius: 7px;
  border: 1px solid var(--acc-border);
  background: var(--acc-50);
  color: var(--acc-500);
  font-size: 12px;
  font-weight: 700;
}
.user-meta { flex: 1; min-width: 0; line-height: 1.25; }
.user-name {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--txt);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.user-role {
  font-family: var(--font-mono);
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--txt-dim);
}
.user-actions {
  display: flex;
  gap: 4px;
  margin-top: 6px;
  padding-top: 6px;
  border-top: 1px solid var(--line-soft);
}
.grow { flex: 1; }

.app-backdrop {
  position: fixed;
  inset: 0;
  z-index: var(--z-backdrop);
  background: rgba(2, 6, 23, 0.5);
  backdrop-filter: blur(2px);
}

/* ---------- Main Column ---------- */
.app-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  height: 100dvh;
  margin-left: 0;
}

.app-topbar {
  position: sticky;
  top: 0;
  z-index: var(--z-sticky);
  display: flex;
  align-items: center;
  gap: 10px;
  height: 52px;
  flex-shrink: 0;
  padding: 0 20px;
  background: var(--topbar-bg);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--line);
}
.topbar-burger { display: inline-flex; }
.topbar-title {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.topbar-page {
  font-size: 13.5px;
  font-weight: 700;
  letter-spacing: -0.015em;
  color: var(--txt);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.topbar-sep {
  color: var(--line);
  font-size: 12px;
}
.topbar-sub {
  font-size: 12px;
  color: var(--txt-dim);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.topbar-right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 8px;
}

.topbar-search {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  height: 32px;
  padding: 0 10px;
  border-radius: 7px;
  border: 1px solid var(--line);
  background: var(--panel-2);
  color: var(--txt-dim);
  cursor: pointer;
  transition: border-color 120ms ease, color 120ms ease, background-color 120ms ease;
}
.topbar-search:hover {
  border-color: var(--acc-border);
  color: var(--txt);
}
.topbar-search kbd {
  border-color: var(--line);
  background: var(--panel);
  color: var(--txt-dim);
}

.topbar-status {
  align-items: center;
  gap: 6px;
  height: 26px;
  padding: 0 9px;
  border-radius: 6px;
  border: 1px solid var(--line);
  background: var(--panel-2);
  font-size: 11px;
  font-weight: 600;
  color: var(--txt-dim);
}
.status-dot {
  width: 6.5px;
  height: 6.5px;
  border-radius: 999px;
  background: var(--sig-ok);
  box-shadow: 0 0 0 2px var(--sig-ok-bg);
}

.app-body {
  flex: 1;
  overflow-y: auto;
  background: var(--ink-950);
}
.app-container {
  width: 100%;
  max-width: 1440px;
  padding: 20px 24px 44px;
  margin: 0 auto;
  box-sizing: border-box;
}

@media (max-width: 640px) {
  .app-topbar { padding: 0 12px; }
  .app-container { padding: 14px 12px 32px; }
}

@media (min-width: 1024px) {
  .app-aside { transform: translateX(0); }
  .app-main { margin-left: 236px; }
  .topbar-burger { display: none; }
}
</style>
