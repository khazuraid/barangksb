<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useThemeStore } from '@/stores/theme'
import { useAuthStore } from '@/stores/auth'
import api from '@/api'
import Dialog from 'primevue/dialog'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits(['update:visible'])

const router = useRouter()
const theme = useThemeStore()
const auth = useAuthStore()

const query = ref('')
const searchResults = ref<any[]>([])
const loadingSearch = ref(false)
const selectedIndex = ref(0)
let debounceTimer: any = null

const staticActions = computed(() => {
  const list = [
    { label: 'Dashboard', icon: 'pi pi-th-large', to: '/', category: 'Navigasi' },
    { label: 'Daftar Barang', icon: 'pi pi-box', to: '/items', category: 'Navigasi' },
    { label: 'Tambah Barang Baru', icon: 'pi pi-plus-circle', to: '/items/new', category: 'Aksi Cepat' },
    { label: 'Mutasi Barang (Masuk/Keluar)', icon: 'pi pi-arrow-right-arrow-left', to: '/movement', category: 'Aksi Cepat' },
    { label: 'Riwayat Transaksi', icon: 'pi pi-history', to: '/history', category: 'Navigasi' },
    { label: 'QR & Barcode Generator', icon: 'pi pi-qrcode', to: '/barcode', category: 'Perangkat' },
    { label: 'Opname Massal (CSV)', icon: 'pi pi-clipboard', to: '/adjust/bulk', category: 'Perangkat' },
    { label: 'Master Kategori', icon: 'pi pi-tags', to: '/categories', category: 'Master Data' },
    { label: 'Master Lokasi', icon: 'pi pi-map-marker', to: '/locations', category: 'Master Data' },
    { label: 'Ganti Tema (Terang / Gelap)', icon: theme.isDark ? 'pi pi-sun' : 'pi pi-moon', action: () => theme.toggle(), category: 'Pengaturan' },
    { label: 'Ganti Password', icon: 'pi pi-key', to: '/password', category: 'Akun' },
  ]
  if (auth.isAdmin) {
    list.push(
      { label: 'Kelola Pengguna', icon: 'pi pi-users', to: '/users', category: 'Administrasi' },
      { label: 'Audit Log', icon: 'pi pi-shield', to: '/audit', category: 'Administrasi' },
      { label: 'Pengaturan Telegram', icon: 'pi pi-send', to: '/telegram', category: 'Administrasi' }
    )
  }
  return list
})

const filteredActions = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return staticActions.value.slice(0, 8)
  return staticActions.value.filter(a => a.label.toLowerCase().includes(q) || a.category.toLowerCase().includes(q))
})

watch(query, (val) => {
  clearTimeout(debounceTimer)
  selectedIndex.value = 0
  const q = val.trim()
  if (q.length < 2) {
    searchResults.value = []
    return
  }
  debounceTimer = setTimeout(async () => {
    loadingSearch.value = true
    try {
      const res = await api.get('/items', { params: { q, per_page: 8 } })
      searchResults.value = res.data.data || []
    } catch {
      searchResults.value = []
    } finally {
      loadingSearch.value = false
    }
  }, 250)
})

function close() {
  emit('update:visible', false)
  query.value = ''
  searchResults.value = []
}

function execute(item: any) {
  close()
  if (item.action) {
    item.action()
  } else if (item.to) {
    router.push(item.to)
  }
}

function openItem(item: any) {
  close()
  router.push(`/items/${item.id}`)
}
</script>

<template>
  <Dialog
    :visible="visible"
    @update:visible="$emit('update:visible', $event)"
    modal
    :closable="false"
    :showHeader="false"
    :style="{ width: '560px', maxWidth: '95vw' }"
    :pt="{
      root: { class: '!rounded-2xl !overflow-hidden !border !border-[var(--line)] !bg-[var(--panel)] shadow-2xl p-0' },
      content: { class: '!p-0' }
    }"
  >
    <div class="p-3 border-b flex items-center gap-2.5" style="border-color: var(--line); background: var(--panel)">
      <i class="pi pi-search text-[14px]" style="color: var(--txt-dim)" />
      <input
        v-model="query"
        type="text"
        placeholder="Cari menu, barang, atau aksi cepat… (Esc untuk tutup)"
        class="w-full bg-transparent border-0 outline-none text-[14px] font-medium"
        style="color: var(--txt)"
        autofocus
        @keydown.esc="close"
      />
      <span class="text-[10px] font-semibold px-1.5 py-0.5 rounded border t-mono shrink-0"
            style="border-color: var(--line); color: var(--txt-dim); background: var(--panel-2)">ESC</span>
    </div>

    <div class="max-h-[380px] overflow-y-auto p-2 flex flex-col gap-1">
      <!-- Item Results from search -->
      <div v-if="searchResults.length" class="mb-2">
        <div class="text-[11px] font-semibold uppercase px-2.5 py-1 tracking-wider" style="color: var(--txt-dim)">
          Barang Terkait
        </div>
        <div
          v-for="item in searchResults"
          :key="item.id"
          @click="openItem(item)"
          class="flex items-center justify-between gap-3 px-3 py-2 rounded-lg cursor-pointer transition-colors"
          style="background: var(--panel-2)"
        >
          <div class="flex items-center gap-2.5 min-w-0">
            <span class="w-7 h-7 rounded-md grid place-items-center border text-[11px] shrink-0 font-bold"
                  style="border-color: var(--line); background: var(--panel)">
              <i class="pi pi-box text-acc-500" />
            </span>
            <div class="truncate">
              <div class="text-[13px] font-semibold truncate">{{ item.name }}</div>
              <div class="t-mono text-[11px]" style="color: var(--txt-dim)">{{ item.sku }} · {{ item.location }}</div>
            </div>
          </div>
          <span class="t-num text-[12px] font-bold px-2 py-0.5 rounded"
                style="background: var(--panel); border: 1px solid var(--line)">
            {{ item.current_stock }} {{ item.unit }}
          </span>
        </div>
      </div>

      <!-- Action items -->
      <div>
        <div class="text-[11px] font-semibold uppercase px-2.5 py-1 tracking-wider" style="color: var(--txt-dim)">
          Menu &amp; Aksi
        </div>
        <div
          v-for="action in filteredActions"
          :key="action.label"
          @click="execute(action)"
          class="flex items-center justify-between px-3 py-2 rounded-lg cursor-pointer transition-colors hover:bg-slate-100 dark:hover:bg-slate-800/60"
        >
          <div class="flex items-center gap-2.5 min-w-0">
            <i :class="action.icon" class="text-[14px] text-acc-500 shrink-0" />
            <span class="text-[13px] font-medium truncate" style="color: var(--txt)">{{ action.label }}</span>
          </div>
          <span class="text-[10.5px] uppercase font-semibold tracking-wider px-1.5 py-0.5 rounded border"
                style="border-color: var(--line); color: var(--txt-dim)">
            {{ action.category }}
          </span>
        </div>
      </div>
    </div>

    <div class="p-2.5 border-t flex items-center justify-between text-[11px]"
         style="border-color: var(--line); background: var(--panel-2); color: var(--txt-dim)">
      <span>Tip: Gunakan <kbd class="px-1 py-0.5 rounded border t-mono text-[10px]">Ctrl+K</kbd> untuk membuka kapan saja</span>
      <span>Pilih aksi untuk berpindah</span>
    </div>
  </Dialog>
</template>
