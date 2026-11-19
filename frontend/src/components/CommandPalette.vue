<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
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

const inputRef = ref<HTMLInputElement | null>(null)
const query = ref('')
const searchResults = ref<any[]>([])
const loadingSearch = ref(false)
const selectedIndex = ref(0)
let debounceTimer: ReturnType<typeof setTimeout> | null = null

type PaletteEntry =
  | { kind: 'item'; id: string; item: any }
  | { kind: 'action'; id: string; label: string; icon: string; category: string; to?: string; action?: () => void }

const staticActions = computed(() => {
  const list: Array<{ label: string; icon: string; category: string; to?: string; action?: () => void }> = [
    { label: 'Dashboard', icon: 'pi pi-th-large', to: '/', category: 'Navigasi' },
    { label: 'Daftar Barang', icon: 'pi pi-box', to: '/items', category: 'Navigasi' },
    { label: 'Tambah Barang Baru', icon: 'pi pi-plus-circle', to: '/items/new', category: 'Aksi Cepat' },
    { label: 'Mutasi Barang (Masuk/Keluar)', icon: 'pi pi-arrow-right-arrow-left', to: '/movement', category: 'Aksi Cepat' },
    { label: 'Pengaduan & Laporan', icon: 'pi pi-megaphone', to: '/reports', category: 'Operasional' },
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
      { label: 'Kelola Foto', icon: 'pi pi-images', to: '/photos', category: 'Administrasi' },
      { label: 'Audit Log Sistem', icon: 'pi pi-shield', to: '/audit', category: 'Administrasi' },
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

const flatEntries = computed<PaletteEntry[]>(() => {
  const entries: PaletteEntry[] = []
  for (const it of searchResults.value) {
    entries.push({ kind: 'item', id: `item:${it.id}`, item: it })
  }
  for (const act of filteredActions.value) {
    entries.push({ kind: 'action', id: `act:${act.label}`, ...act })
  }
  return entries
})

watch(() => props.visible, async (open) => {
  if (open) {
    selectedIndex.value = 0
    await nextTick()
    inputRef.value?.focus()
  }
})

watch(flatEntries, (list) => {
  if (selectedIndex.value >= list.length) selectedIndex.value = Math.max(0, list.length - 1)
})

watch(query, (val) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  selectedIndex.value = 0
  const q = val.trim()
  if (q.length < 2) {
    searchResults.value = []
    loadingSearch.value = false
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
  }, 200)
})

function close() {
  emit('update:visible', false)
  query.value = ''
  searchResults.value = []
  selectedIndex.value = 0
}

function runEntry(entry: PaletteEntry) {
  close()
  if (entry.kind === 'item') {
    router.push(`/items/${entry.item.id}`)
    return
  }
  if (entry.action) entry.action()
  else if (entry.to) router.push(entry.to)
}

function onInputKeydown(e: KeyboardEvent) {
  const len = flatEntries.value.length
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    if (len) selectedIndex.value = (selectedIndex.value + 1) % len
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    if (len) selectedIndex.value = (selectedIndex.value - 1 + len) % len
  } else if (e.key === 'Enter') {
    e.preventDefault()
    const entry = flatEntries.value[selectedIndex.value]
    if (entry) runEntry(entry)
  } else if (e.key === 'Escape') {
    e.preventDefault()
    close()
  }
}

function entryIndexOfItem(idx: number) {
  return idx
}

function entryIndexOfAction(idx: number) {
  return searchResults.value.length + idx
}
</script>

<template>
  <Dialog
    :visible="visible"
    @update:visible="(v) => !v && close()"
    modal
    :closable="false"
    :showHeader="false"
    :style="{ width: '560px', maxWidth: '94vw' }"
    :pt="{
      root: { class: '!rounded-[12px] !overflow-hidden !border !border-line !bg-panel shadow-[var(--shadow-lg)] !p-0' },
      content: { class: '!p-0' }
    }"
  >
    <div class="flex flex-col">
      <!-- Search Bar -->
      <div class="flex items-center gap-2.5 px-4 h-[48px] border-b border-line bg-panel-2">
        <i
          :class="loadingSearch ? 'pi pi-spin pi-spinner text-acc-500' : 'pi pi-search text-txt-dim'"
          class="text-[13px] shrink-0"
          aria-hidden="true"
        />
        <input
          ref="inputRef"
          v-model="query"
          type="text"
          autocomplete="off"
          placeholder="Cari barang, SKU, lokasi, atau lompat ke menu…"
          class="flex-1 bg-transparent border-0 outline-none text-[13.5px] text-txt placeholder:text-txt-dim"
          @keydown="onInputKeydown"
        />
        <kbd class="t-mono !text-[10px] px-1.5 py-0.5 rounded border border-line bg-panel text-txt-dim">ESC</kbd>
      </div>

      <!-- Results -->
      <div class="max-h-[380px] overflow-y-auto p-2 space-y-2">
        <!-- Matching Items -->
        <div v-if="searchResults.length">
          <div class="px-2.5 py-1 text-[10.5px] font-semibold uppercase tracking-[0.06em] text-txt-dim">
            Barang ditemukan
          </div>
          <button
            v-for="(it, idx) in searchResults"
            :key="it.id"
            type="button"
            class="w-full flex items-center justify-between gap-3 px-2.5 py-2 rounded-[7px] text-left transition-colors duration-75 cursor-pointer"
            :class="selectedIndex === entryIndexOfItem(idx)
              ? 'bg-[var(--acc-wash)] text-txt shadow-[inset_2px_0_0_var(--acc-500)]'
              : 'hover:bg-[var(--row-hover)] text-txt'"
            @mouseenter="selectedIndex = entryIndexOfItem(idx)"
            @click="runEntry({ kind: 'item', id: `item:${it.id}`, item: it })"
          >
            <div class="flex items-center gap-2.5 min-w-0">
              <div class="w-7 h-7 rounded-[6px] border border-line bg-panel-2 grid place-items-center shrink-0">
                <i class="pi pi-box text-[12px] text-acc-500" aria-hidden="true" />
              </div>
              <div class="min-w-0">
                <div class="text-[13px] font-semibold truncate">{{ it.name }}</div>
                <div class="flex items-center gap-2 text-[11px] text-txt-dim">
                  <span class="t-mono">{{ it.sku }}</span>
                  <span v-if="it.location">· {{ it.location }}</span>
                </div>
              </div>
            </div>
            <span class="t-num text-[11.5px] font-semibold text-txt-dim shrink-0">
              {{ it.current_stock }} {{ it.unit }}
            </span>
          </button>
        </div>

        <!-- Navigation & Quick Actions -->
        <div v-if="filteredActions.length">
          <div class="px-2.5 py-1 text-[10.5px] font-semibold uppercase tracking-[0.06em] text-txt-dim">
            {{ query ? 'Menu & Aksi' : 'Pintasan Cepat' }}
          </div>
          <button
            v-for="(act, idx) in filteredActions"
            :key="act.label"
            type="button"
            class="w-full flex items-center justify-between gap-3 px-2.5 py-2 rounded-[7px] text-left transition-colors duration-75 cursor-pointer"
            :class="selectedIndex === entryIndexOfAction(idx)
              ? 'bg-[var(--acc-wash)] text-txt shadow-[inset_2px_0_0_var(--acc-500)]'
              : 'hover:bg-[var(--row-hover)] text-txt'"
            @mouseenter="selectedIndex = entryIndexOfAction(idx)"
            @click="runEntry({ kind: 'action', id: `act:${act.label}`, ...act })"
          >
            <div class="flex items-center gap-2.5 min-w-0">
              <i
                :class="[act.icon, selectedIndex === entryIndexOfAction(idx) ? 'text-acc-500' : 'text-txt-dim']"
                class="text-[13px] w-4 text-center shrink-0"
                aria-hidden="true"
              />
              <span class="text-[13px] font-medium truncate">{{ act.label }}</span>
            </div>
            <span class="text-[11px] text-txt-dim shrink-0">{{ act.category }}</span>
          </button>
        </div>

        <!-- No Results -->
        <div v-if="!loadingSearch && !flatEntries.length" class="py-10 text-center">
          <div class="text-[13px] font-medium text-txt">Tidak ada hasil untuk "{{ query }}"</div>
          <div class="text-[11.5px] text-txt-dim mt-1">Coba cari nama barang, kode SKU, atau menu lain.</div>
        </div>
      </div>

      <!-- Keyboard Legend Footer -->
      <div class="flex items-center justify-between px-4 py-2 border-t border-line bg-panel-2 text-[11px] text-txt-dim">
        <div class="flex items-center gap-3">
          <span class="flex items-center gap-1"><kbd class="t-mono !text-[10px] px-1 py-0.5 rounded border border-line bg-panel">↑↓</kbd> pilih</span>
          <span class="flex items-center gap-1"><kbd class="t-mono !text-[10px] px-1 py-0.5 rounded border border-line bg-panel">↵</kbd> buka</span>
        </div>
        <span class="t-mono !text-[10.5px]">⌘K</span>
      </div>
    </div>
  </Dialog>
</template>
