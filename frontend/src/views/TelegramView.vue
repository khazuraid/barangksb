<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import { useConfirm } from 'primevue/useconfirm'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import ToggleButton from 'primevue/togglebutton'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import Field from '@/components/Field.vue'
import EmptyState from '@/components/EmptyState.vue'

const toast = useToast()
const confirm = useConfirm()

// ---- state ----
const loading = ref(true)
const saving = ref(false)
const testing = ref(false)
const testingAll = ref(false)
const sendingAlertNow = ref(false)
const refreshingBot = ref(false)
const activeTab = ref<'overview' | 'settings' | 'subscribers' | 'commands' | 'logs' | 'guide'>('overview')

const settings = ref<Record<string, string>>({})
const botInfo = ref<any>(null)
const subscribers = ref<any[]>([])
const logs = ref<any[]>([])
const logsTotal = ref(0)
const logsPage = ref(0)
const logsPerPage = ref(15)

const newChatID = ref('')
const newChatTitle = ref('')
const testMsg = ref('')

// ---- computed ----
const tokenConfigured = computed(() => !!settings.value.telegram_bot_token)
const botConnected = computed(() => botInfo.value?.configured && !botInfo.value?.error)
const activeSubscribersCount = computed(() => subscribers.value.filter(s => s.active).length)

// ---- fetch ----
async function fetchAll() {
  loading.value = true
  try {
    await Promise.all([fetchSettings(), fetchBotInfo(), fetchSubscribers(), fetchLogs(), fetchCommands()])
  } finally {
    loading.value = false
  }
}

async function fetchSettings() {
  const res = await api.get('/settings/telegram')
  settings.value = res.data
}

async function fetchBotInfo() {
  try {
    const res = await api.get('/settings/telegram/bot')
    botInfo.value = res.data
  } catch {
    botInfo.value = { configured: false }
  }
}

async function fetchSubscribers() {
  const res = await api.get('/settings/telegram/subscribers')
  subscribers.value = res.data
}

async function fetchLogs() {
  const res = await api.get('/settings/telegram/logs', {
    params: { page: logsPage.value + 1, per_page: logsPerPage.value }
  })
  logs.value = res.data.data
  logsTotal.value = res.data.total
}

onMounted(fetchAll)

// ---- actions ----
async function saveSettings() {
  saving.value = true
  try {
    await api.put('/settings/telegram', {
      telegram_bot_token: settings.value.telegram_bot_token,
      telegram_api_url: settings.value.telegram_api_url || '',
      telegram_webapp_url: settings.value.telegram_webapp_url || '',
      telegram_alert_low_stock: settings.value.telegram_alert_low_stock || 'false',
      telegram_alert_daily_time: settings.value.telegram_alert_daily_time || '07:00',
    })
    toast.add({ severity: 'success', summary: 'Pengaturan Disimpan', detail: 'Konfigurasi bot berhasil diperbarui', life: 2500 })
    await fetchBotInfo()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal Menyimpan', detail: e.response?.data?.error, life: 4000 })
  } finally {
    saving.value = false
  }
}

async function refreshBot() {
  refreshingBot.value = true
  try {
    await fetchBotInfo()
    toast.add({ severity: 'info', summary: 'Status Diperbarui', life: 2000 })
  } finally {
    refreshingBot.value = false
  }
}

async function addSubscriber() {
  const chatID = parseInt(newChatID.value)
  if (!chatID) {
    toast.add({ severity: 'warn', summary: 'Chat ID tidak valid', life: 2500 })
    return
  }
  try {
    await api.post('/settings/telegram/subscribers', { chat_id: chatID, title: newChatTitle.value })
    toast.add({ severity: 'success', summary: 'Subscriber Berhasil Ditambahkan', life: 2500 })
    newChatID.value = ''
    newChatTitle.value = ''
    await fetchSubscribers()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal Menambah Subscriber', detail: e.response?.data?.error, life: 4000 })
  }
}

async function toggleSub(sub: any, field: string, val: boolean) {
  try {
    await api.put(`/settings/telegram/subscribers/${sub.chat_id}`, { [field]: val })
    sub[field] = val
    toast.add({ severity: 'success', summary: 'Pengaturan Diperbarui', life: 1500 })
  } catch {
    toast.add({ severity: 'error', summary: 'Gagal update subscriber', life: 2500 })
  }
}

function removeSub(sub: any) {
  confirm.require({
    message: `Hapus subscriber "${sub.title || sub.chat_id}" (${sub.chat_id})?`,
    header: 'Konfirmasi Hapus',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: async () => {
      await api.delete(`/settings/telegram/subscribers/${sub.chat_id}`)
      toast.add({ severity: 'success', summary: 'Subscriber dihapus', life: 2500 })
      await fetchSubscribers()
    },
  })
}

async function testOne(chatID: number) {
  testing.value = true
  try {
    const res = await api.post('/settings/telegram/test', { chat_id: chatID, message: testMsg.value })
    const r = res.data.results?.[0]
    if (r?.status === 'ok') {
      toast.add({ severity: 'success', summary: 'Terkirim', detail: r.title, life: 3000 })
    } else {
      toast.add({ severity: 'error', summary: 'Gagal', detail: r?.error || 'Unknown error', life: 5000 })
    }
    await fetchLogs()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Test gagal', detail: e.response?.data?.error, life: 5000 })
  } finally {
    testing.value = false
  }
}

async function testAll() {
  testingAll.value = true
  try {
    const res = await api.post('/settings/telegram/test', { all: true, message: testMsg.value })
    const results = res.data.results || []
    const ok = results.filter((r: any) => r.status === 'ok').length
    const fail = results.filter((r: any) => r.status === 'fail').length
    if (fail === 0 && ok > 0) {
      toast.add({ severity: 'success', summary: 'Semua Pesan Terkirim', detail: `${ok} chat tujuan sukses`, life: 3000 })
    } else if (ok > 0) {
      toast.add({ severity: 'warn', summary: 'Terkirim Sebagian', detail: `${ok} sukses, ${fail} gagal`, life: 4000 })
    } else {
      toast.add({ severity: 'error', summary: 'Semua Gagal Terkirim', life: 5000 })
    }
    await fetchLogs()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Test gagal', detail: e.response?.data?.error, life: 5000 })
  } finally {
    testingAll.value = false
  }
}

async function sendAlertNow() {
  sendingAlertNow.value = true
  try {
    await api.post('/settings/telegram/test', {
      all: true,
      message: '🚨 [Broadcast Manual] Ringkasan stok inventaris dikirim dari Web Panel.'
    })
    toast.add({ severity: 'success', summary: 'Alert Stok Terkirim', life: 3000 })
    await fetchLogs()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal kirim alert', detail: e.response?.data?.error, life: 4000 })
  } finally {
    sendingAlertNow.value = false
  }
}

function copyText(txt: string) {
  navigator.clipboard.writeText(txt)
  toast.add({ severity: 'info', summary: 'Tersalin ke Clipboard', detail: txt, life: 2000 })
}

const lastLogPage = computed(() => Math.max(0, Math.ceil(logsTotal.value / logsPerPage.value) - 1))

// ---- commands CRUD ----
const commands = ref<any[]>([])
const cmdLoading = ref(false)
const cmdSaving = ref(false)
const editingCmd = ref<any>(null)
const showCmdForm = ref(false)

const blankCmd = () => ({
  id: 0, command: '', label: '', description: '', response: '',
  is_menu: true, sort_order: 10, active: true,
})

async function fetchCommands() {
  cmdLoading.value = true
  try {
    const res = await api.get('/settings/telegram/commands')
    commands.value = res.data
  } catch {
    toast.add({ severity: 'error', summary: 'Gagal memuat daftar command', life: 2500 })
  } finally {
    cmdLoading.value = false
  }
}

function addCmd() {
  editingCmd.value = blankCmd()
  showCmdForm.value = true
}

function editCmd(c: any) {
  editingCmd.value = { ...c }
  showCmdForm.value = true
}

async function saveCmd() {
  if (!editingCmd.value.command || !editingCmd.value.label) {
    toast.add({ severity: 'warn', summary: 'Command dan label wajib diisi', life: 2500 })
    return
  }
  cmdSaving.value = true
  try {
    if (editingCmd.value.id) {
      await api.put(`/settings/telegram/commands/${editingCmd.value.id}`, editingCmd.value)
    } else {
      await api.post('/settings/telegram/commands', editingCmd.value)
    }
    toast.add({ severity: 'success', summary: 'Command Berhasil Disimpan', life: 2500 })
    showCmdForm.value = false
    await fetchCommands()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal Menyimpan Command', detail: e.response?.data?.error, life: 4000 })
  } finally {
    cmdSaving.value = false
  }
}

function removeCmd(c: any) {
  confirm.require({
    message: `Hapus command "/${c.command}"?`,
    header: 'Konfirmasi Hapus Command',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: async () => {
      await api.delete(`/settings/telegram/commands/${c.id}`)
      toast.add({ severity: 'success', summary: 'Command dihapus', life: 2500 })
      await fetchCommands()
    },
  })
}
</script>

<template>
  <div class="space-y-5">
    <PageHeader crumb="Administrasi" title="Integrasi Telegram & Bot Pintar"
      sub="Pengaturan bot cerdas, Telegram Mini App, scanner barcode, subscriber, dan notifikasi real-time">
      <template #actions>
        <Button label="Refresh Status" icon="pi pi-sync" size="small" text severity="secondary"
                :loading="refreshingBot" @click="refreshBot" />
        <Button label="Simpan Pengaturan" icon="pi pi-save" size="small"
                :loading="saving" :disabled="loading" @click="saveSettings" />
      </template>
    </PageHeader>

    <div v-if="loading" class="grid place-items-center py-24">
      <i class="pi pi-spin pi-spinner text-3xl text-emerald-500" />
      <span class="text-xs text-slate-400 mt-2">Memuat konfigurasi Telegram...</span>
    </div>

    <div v-else class="space-y-5">
      <!-- ============ HERO BOT STATUS BANNER ============ -->
      <div class="p-5 rounded-2xl border transition-all"
           :class="botConnected
             ? 'bg-gradient-to-r from-emerald-950/30 via-slate-900 to-slate-900 border-emerald-500/30 shadow-lg shadow-emerald-950/20'
             : 'bg-gradient-to-r from-rose-950/25 via-slate-900 to-slate-900 border-rose-500/30'">
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div class="flex items-center gap-4">
            <div class="relative w-14 h-14 rounded-2xl grid place-items-center font-bold text-xl shadow-inner border border-white/10"
                 :class="botConnected ? 'bg-emerald-500 text-slate-950' : 'bg-rose-500/20 text-rose-400 border-rose-500/30'">
              <i v-if="!botConnected" class="pi pi-exclamation-circle text-2xl" />
              <span v-else>{{ botInfo?.first_name?.[0] || 'B' }}</span>
              <span class="absolute -bottom-1 -right-1 w-4 h-4 rounded-full border-2 border-slate-900"
                    :class="botConnected ? 'bg-emerald-400' : 'bg-rose-500'"></span>
            </div>
            <div>
              <div class="flex items-center gap-2.5">
                <h2 class="text-lg font-bold tracking-tight text-white">
                  {{ botConnected ? (botInfo.first_name + (botInfo.last_name ? ' ' + botInfo.last_name : '')) : 'Bot Belum Terhubung' }}
                </h2>
                <Tag :severity="botConnected ? 'success' : 'danger'"
                     :value="botConnected ? 'ONLINE & SIAP' : 'TERPUTUS'" class="!text-[11px] font-semibold" />
              </div>
              <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-400 mt-1">
                <span v-if="botConnected" class="font-mono text-emerald-400">@{{ botInfo.username }}</span>
                <span v-if="botConnected" class="font-mono text-slate-500">ID: {{ botInfo.id }}</span>
                <span class="flex items-center gap-1.5">
                  <i class="pi pi-users text-[11px]"></i>
                  <span class="text-slate-300 font-semibold">{{ activeSubscribersCount }}</span> subscriber aktif
                </span>
                <span class="flex items-center gap-1.5">
                  <i class="pi pi-bell text-[11px]"></i>
                  Daily Alert: <span class="text-slate-300 font-medium">{{ settings.telegram_alert_daily_time || '07:00' }} WIB</span>
                </span>
              </div>
            </div>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <Button label="Kirim Alert Stok Kritis" icon="pi pi-bell" size="small" severity="warn"
                    :loading="sendingAlertNow" :disabled="!activeSubscribersCount || !botConnected"
                    @click="sendAlertNow" />
            <Button v-if="botConnected" label="Buka Bot" icon="pi pi-external-link" size="small" outlined
                    @click="window.open(`https://t.me/${botInfo.username}`, '_blank')" />
          </div>
        </div>
      </div>

      <!-- ============ TAB NAVIGATION ============ -->
      <div class="flex items-center gap-1.5 p-1 bg-slate-900/80 border border-slate-800 rounded-xl overflow-x-auto">
        <button v-for="tab in [
          { id: 'overview', label: 'Ringkasan & Status', icon: 'pi pi-home' },
          { id: 'settings', label: 'Koneksi & Mini App', icon: 'pi pi-cog' },
          { id: 'subscribers', label: `Subscriber (${subscribers.length})`, icon: 'pi pi-users' },
          { id: 'commands', label: `Command Menu (${commands.length})`, icon: 'pi pi-list' },
          { id: 'logs', label: `Riwayat Pesan (${logsTotal})`, icon: 'pi pi-history' },
          { id: 'guide', label: 'Panduan Setup', icon: 'pi pi-book' }
        ]" :key="tab.id"
          class="flex items-center gap-2 px-3.5 py-2 rounded-lg text-xs font-semibold whitespace-nowrap transition-all"
          :class="activeTab === tab.id
            ? 'bg-emerald-500 text-slate-950 shadow-md shadow-emerald-500/10'
            : 'text-slate-400 hover:text-white hover:bg-slate-800/60'"
          @click="activeTab = tab.id as any">
          <i :class="tab.icon"></i>
          <span>{{ tab.label }}</span>
        </button>
      </div>

      <!-- ============ TAB 1: OVERVIEW ============ -->
      <div v-if="activeTab === 'overview'" class="grid md:grid-cols-3 gap-4">
        <!-- Feature 1: Bot Capabilities Card -->
        <div class="bg-slate-900/90 border border-slate-800/80 rounded-2xl p-5 space-y-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <div class="w-8 h-8 rounded-lg bg-emerald-500/10 border border-emerald-500/20 grid place-items-center text-emerald-400">
                <i class="pi pi-bolt text-sm"></i>
              </div>
              <h3 class="font-bold text-sm text-white">Fitur Pintar Aktif</h3>
            </div>
            <Tag severity="success" value="v2.0 PRO" class="!text-[10px]" />
          </div>

          <div class="space-y-2.5 text-xs">
            <div class="flex items-center justify-between p-2.5 rounded-xl bg-slate-950/60 border border-slate-800/60">
              <span class="flex items-center gap-2 text-slate-300">
                <i class="pi pi-camera text-emerald-400"></i> Scan Barcode & QR Foto
              </span>
              <span class="text-emerald-400 font-semibold">Aktif</span>
            </div>
            <div class="flex items-center justify-between p-2.5 rounded-xl bg-slate-950/60 border border-slate-800/60">
              <span class="flex items-center gap-2 text-slate-300">
                <i class="pi pi-search text-emerald-400"></i> Cari Barang Otomatis
              </span>
              <span class="text-emerald-400 font-semibold">Aktif</span>
            </div>
            <div class="flex items-center justify-between p-2.5 rounded-xl bg-slate-950/60 border border-slate-800/60">
              <span class="flex items-center gap-2 text-slate-300">
                <i class="pi pi-plus-circle text-emerald-400"></i> Quick Command /tambah
              </span>
              <span class="text-emerald-400 font-semibold">Aktif</span>
            </div>
            <div class="flex items-center justify-between p-2.5 rounded-xl bg-slate-950/60 border border-slate-800/60">
              <span class="flex items-center gap-2 text-slate-300">
                <i class="pi pi-table text-emerald-400"></i> Unduh Rekap CSV
              </span>
              <span class="text-emerald-400 font-semibold">Aktif</span>
            </div>
            <div class="flex items-center justify-between p-2.5 rounded-xl bg-slate-950/60 border border-slate-800/60">
              <span class="flex items-center gap-2 text-slate-300">
                <i class="pi pi-mobile text-emerald-400"></i> Telegram Mini App Button
              </span>
              <span class="text-emerald-400 font-semibold">Terintegrasi</span>
            </div>
          </div>
        </div>

        <!-- Feature 2: Quick Test & Broadcast -->
        <div class="bg-slate-900/90 border border-slate-800/80 rounded-2xl p-5 space-y-4">
          <div class="flex items-center gap-2.5">
            <div class="w-8 h-8 rounded-lg bg-sky-500/10 border border-sky-500/20 grid place-items-center text-sky-400">
              <i class="pi pi-send text-sm"></i>
            </div>
            <h3 class="font-bold text-sm text-white">Uji Kirim Pesan Cepat</h3>
          </div>

          <div class="space-y-3">
            <div>
              <label class="block text-xs font-semibold text-slate-300 mb-1">Pesan Uji Coba</label>
              <InputText v-model="testMsg" placeholder="Tes koneksi bot inventaris..." class="w-full text-xs" />
            </div>
            <div class="flex gap-2">
              <Button label="Kirim ke Semua" icon="pi pi-send" size="small" class="w-full"
                      :loading="testingAll" :disabled="!subscribers.length" @click="testAll" />
            </div>
            <p class="text-[11px] text-slate-400 leading-relaxed">
              Pesan uji coba akan dikirim ke seluruh subscriber aktif dan dicatat ke log sistem.
            </p>
          </div>
        </div>

        <!-- Feature 3: Quick Registration Info -->
        <div class="bg-slate-900/90 border border-slate-800/80 rounded-2xl p-5 space-y-4">
          <div class="flex items-center gap-2.5">
            <div class="w-8 h-8 rounded-lg bg-indigo-500/10 border border-indigo-500/20 grid place-items-center text-indigo-400">
              <i class="pi pi-id-card text-sm"></i>
            </div>
            <h3 class="font-bold text-sm text-white">Akses Petugas & Grup</h3>
          </div>

          <div class="space-y-2 text-xs text-slate-300 leading-relaxed">
            <p>Untuk menghubungkan akun petugas baru:</p>
            <ol class="list-decimal list-inside space-y-1 text-slate-400 text-[11.5px]">
              <li>Buka bot di Telegram: <strong class="text-white">@{{ botInfo?.username || 'bot' }}</strong></li>
              <li>Kirim perintah <code class="px-1.5 py-0.5 rounded bg-slate-950 text-emerald-400 font-mono">/start</code></li>
              <li>Salin <strong class="text-white">Chat ID</strong> yang ditampilkan bot</li>
              <li>Masukkan pada tab <button class="text-emerald-400 underline font-semibold" @click="activeTab = 'subscribers'">Subscriber</button></li>
            </ol>
          </div>
        </div>
      </div>

      <!-- ============ TAB 2: SETTINGS & MINI APP ============ -->
      <div v-if="activeTab === 'settings'" class="grid lg:grid-cols-2 gap-5">
        <Panel title="Token & Koneksi Jaringan" icon="pi pi-key">
          <div class="space-y-4">
            <Field label="Bot Token" hint="Diberikan oleh @BotFather. Rahasiakan token ini.">
              <InputText v-model="settings.telegram_bot_token"
                         :placeholder="settings.telegram_bot_token ? '••••••••••••••••' : '123456:ABC-DEF...'"
                         class="w-full font-mono text-xs" type="password" />
            </Field>

            <Field label="URL Telegram Mini App (Web App)" hint="URL aplikasi web yang dibuka saat tombol 'Buka Inventaris' ditekan di bot.">
              <div class="flex gap-2">
                <InputText v-model="settings.telegram_webapp_url"
                           placeholder="https://barang.kesling.biz.id"
                           class="w-full text-xs font-mono" />
                <Button icon="pi pi-external-link" text size="small"
                        @click="window.open(settings.telegram_webapp_url || 'https://barang.kesling.biz.id', '_blank')" />
              </div>
            </Field>

            <Field label="Telegram API Proxy (Opsional)" hint="Kosongkan jika default (https://api.telegram.org). Diisi bila hosting memblokir Telegram.">
              <InputText v-model="settings.telegram_api_url"
                         placeholder="https://api.telegram.org"
                         class="w-full text-xs font-mono" />
            </Field>

            <div class="pt-2">
              <Button label="Simpan Koneksi" icon="pi pi-check" size="small" :loading="saving" @click="saveSettings" />
            </div>
          </div>
        </Panel>

        <Panel title="Jadwal Daily Low-Stock Alert" icon="pi pi-clock">
          <div class="space-y-4">
            <div class="flex items-center justify-between p-3.5 rounded-xl bg-slate-950/70 border border-slate-800">
              <div>
                <div class="text-xs font-bold text-white">Daily Low-Stock Alert Otomatis</div>
                <div class="text-[11px] text-slate-400">Mengirim rekap barang menipis setiap hari kerja</div>
              </div>
              <ToggleButton v-model="settings.telegram_alert_low_stock"
                            :modelValue="settings.telegram_alert_low_stock === 'true'"
                            @update:modelValue="settings.telegram_alert_low_stock = $event ? 'true' : 'false'"
                            onLabel="Aktif" offLabel="Mati" onIcon="pi pi-check" offIcon="pi pi-times" />
            </div>

            <Field label="Waktu Pengiriman Harian (WIB)" hint="Format 24 jam (misal 07:00)">
              <InputText v-model="settings.telegram_alert_daily_time" placeholder="07:00" class="w-full text-xs" />
            </Field>

            <div class="p-3 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-300 text-xs flex gap-2.5 items-start">
              <i class="pi pi-info-circle text-sm mt-0.5"></i>
              <div class="leading-relaxed text-[11.5px]">
                Notifikasi harian dikirimkan hanya kepada subscriber yang mengaktifkan opsi <strong>Low Stock</strong> pada tabel subscriber.
              </div>
            </div>

            <div class="pt-2">
              <Button label="Simpan Jadwal" icon="pi pi-check" size="small" :loading="saving" @click="saveSettings" />
            </div>
          </div>
        </Panel>
      </div>

      <!-- ============ TAB 3: SUBSCRIBERS ============ -->
      <div v-if="activeTab === 'subscribers'" class="space-y-4">
        <Panel title="Daftar Subscriber & Hak Notifikasi" icon="pi pi-users" dense>
          <template #actions>
            <Tag severity="info" :value="subscribers.length + ' Penerima Terdaftar'" />
          </template>

          <!-- add subscriber form -->
          <div class="p-4 bg-slate-950/50 border-b border-slate-800/80">
            <div class="flex flex-wrap gap-3 items-end">
              <Field label="Chat ID (Angka)">
                <InputText v-model="newChatID" placeholder="mis. 123456789 atau -100..." class="w-56 text-xs font-mono" />
              </Field>
              <Field label="Label / Nama Penerima">
                <InputText v-model="newChatTitle" placeholder="mis. dr. Fikri (Poli Umum)" class="w-64 text-xs" />
              </Field>
              <Button label="Tambah Penerima" icon="pi pi-plus" size="small" @click="addSubscriber" />
            </div>
          </div>

          <EmptyState v-if="!subscribers.length" icon="pi pi-users" title="Belum Ada Subscriber"
                      sub="Tambahkan Chat ID secara manual atau minta staf mengirimkan perintah /start ke bot." />

          <div v-else class="overflow-x-auto">
            <table class="w-full text-xs">
              <thead>
                <tr class="text-left border-b border-slate-800 bg-slate-950/60 text-slate-400 text-[11px] uppercase tracking-wider">
                  <th class="px-4 py-3">Penerima / Chat</th>
                  <th class="px-4 py-3">Tipe</th>
                  <th class="px-4 py-3 text-center">Masuk</th>
                  <th class="px-4 py-3 text-center">Keluar</th>
                  <th class="px-4 py-3 text-center">Opname</th>
                  <th class="px-4 py-3 text-center">Low Stock</th>
                  <th class="px-4 py-3 text-center">Aktif</th>
                  <th class="px-4 py-3 text-right">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-800/60">
                <tr v-for="s in subscribers" :key="s.chat_id" class="hover:bg-slate-800/30 transition-colors">
                  <td class="px-4 py-3">
                    <div class="font-bold text-white text-xs">{{ s.title || 'Tanpa Nama' }}</div>
                    <div class="font-mono text-[11px] text-slate-400">ID: {{ s.chat_id }}</div>
                    <div v-if="s.username" class="font-mono text-[10.5px] text-emerald-400">@{{ s.username }}</div>
                  </td>
                  <td class="px-4 py-3">
                    <Tag :severity="s.type === 'private' ? 'info' : 'warn'" :value="s.type.toUpperCase()" class="!text-[10px]" />
                  </td>
                  <td class="px-4 py-3 text-center">
                    <ToggleButton :modelValue="s.notify_in" @update:modelValue="toggleSub(s, 'notify_in', $event)"
                                  onLabel="" offLabel="" onIcon="pi pi-check" offIcon="pi pi-times" class="!p-1" />
                  </td>
                  <td class="px-4 py-3 text-center">
                    <ToggleButton :modelValue="s.notify_out" @update:modelValue="toggleSub(s, 'notify_out', $event)"
                                  onLabel="" offLabel="" onIcon="pi pi-check" offIcon="pi pi-times" class="!p-1" />
                  </td>
                  <td class="px-4 py-3 text-center">
                    <ToggleButton :modelValue="s.notify_adjust" @update:modelValue="toggleSub(s, 'notify_adjust', $event)"
                                  onLabel="" offLabel="" onIcon="pi pi-check" offIcon="pi pi-times" class="!p-1" />
                  </td>
                  <td class="px-4 py-3 text-center">
                    <ToggleButton :modelValue="s.notify_low_stock" @update:modelValue="toggleSub(s, 'notify_low_stock', $event)"
                                  onLabel="" offLabel="" onIcon="pi pi-check" offIcon="pi pi-times" class="!p-1" />
                  </td>
                  <td class="px-4 py-3 text-center">
                    <ToggleButton :modelValue="s.active" @update:modelValue="toggleSub(s, 'active', $event)"
                                  onLabel="" offLabel="" onIcon="pi pi-check" offIcon="pi pi-times" class="!p-1" />
                  </td>
                  <td class="px-4 py-3 text-right">
                    <div class="flex items-center justify-end gap-1">
                      <Button icon="pi pi-send" text rounded size="small" severity="info"
                              v-tooltip.top="'Uji kirim ke chat ini'" @click="testOne(s.chat_id)" />
                      <Button icon="pi pi-trash" text rounded size="small" severity="danger"
                              v-tooltip.top="'Hapus subscriber'" @click="removeSub(s)" />
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </Panel>
      </div>

      <!-- ============ TAB 4: COMMANDS ============ -->
      <div v-if="activeTab === 'commands'" class="space-y-4">
        <Panel title="Daftar Command & Respon Bot" icon="pi pi-list" dense>
          <template #actions>
            <Button label="Tambah Command" icon="pi pi-plus" size="small" @click="addCmd" />
          </template>

          <EmptyState v-if="!commands.length" icon="pi pi-list" title="Belum Ada Command"
                      sub="Tambah perintah khusus untuk bot." />

          <div v-else class="overflow-x-auto">
            <table class="w-full text-xs">
              <thead>
                <tr class="text-left border-b border-slate-800 bg-slate-950/60 text-slate-400 text-[11px] uppercase tracking-wider">
                  <th class="px-4 py-3">Command</th>
                  <th class="px-4 py-3">Label Menu</th>
                  <th class="px-4 py-3 text-center">Tampil Menu</th>
                  <th class="px-4 py-3 text-center">Urutan</th>
                  <th class="px-4 py-3 text-center">Status</th>
                  <th class="px-4 py-3 text-right">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-800/60">
                <tr v-for="c in commands" :key="c.id" class="hover:bg-slate-800/30 transition-colors">
                  <td class="px-4 py-3">
                    <span class="font-mono font-bold text-emerald-400">/{{ c.command }}</span>
                    <div class="text-[11px] text-slate-400">{{ c.description || '-' }}</div>
                  </td>
                  <td class="px-4 py-3 font-medium text-white">{{ c.label }}</td>
                  <td class="px-4 py-3 text-center">
                    <Tag :severity="c.is_menu ? 'success' : 'secondary'" :value="c.is_menu ? 'YA' : 'TIDAK'" />
                  </td>
                  <td class="px-4 py-3 text-center font-mono">{{ c.sort_order }}</td>
                  <td class="px-4 py-3 text-center">
                    <Tag :severity="c.active ? 'success' : 'danger'" :value="c.active ? 'AKTIF' : 'NONAKTIF'" />
                  </td>
                  <td class="px-4 py-3 text-right">
                    <Button icon="pi pi-pencil" text rounded size="small" severity="secondary" @click="editCmd(c)" />
                    <Button icon="pi pi-trash" text rounded size="small" severity="danger" @click="removeCmd(c)" />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </Panel>
      </div>

      <!-- ============ TAB 5: LOGS ============ -->
      <div v-if="activeTab === 'logs'" class="space-y-4">
        <Panel title="Riwayat Log Notifikasi Terkirim" icon="pi pi-history" dense>
          <template #actions>
            <Button label="Segarkan" icon="pi pi-sync" size="small" text severity="secondary" @click="fetchLogs" />
          </template>

          <EmptyState v-if="!logs.length" icon="pi pi-history" title="Belum Ada Log Pesan"
                      sub="Pesan notifikasi dan alert yang dikirim bot akan tercatat di sini." />

          <div v-else class="overflow-x-auto">
            <table class="w-full text-xs">
              <thead>
                <tr class="text-left border-b border-slate-800 bg-slate-950/60 text-slate-400 text-[11px] uppercase tracking-wider">
                  <th class="px-4 py-3 w-40">Waktu</th>
                  <th class="px-4 py-3">Tujuan</th>
                  <th class="px-4 py-3">Pesan</th>
                  <th class="px-4 py-3 w-28 text-right">Status</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-800/60">
                <tr v-for="l in logs" :key="l.id" class="hover:bg-slate-800/30 transition-colors">
                  <td class="px-4 py-3 font-mono text-[11px] text-slate-400 whitespace-nowrap">{{ l.sent_at }}</td>
                  <td class="px-4 py-3">
                    <div class="font-bold text-white">{{ l.chat_title || 'Chat ID: ' + l.chat_id }}</div>
                    <div class="font-mono text-[10.5px] text-slate-500">{{ l.chat_id }}</div>
                  </td>
                  <td class="px-4 py-3 text-slate-300 max-w-md truncate">{{ l.message }}</td>
                  <td class="px-4 py-3 text-right">
                    <Tag :severity="l.status === 'ok' ? 'success' : 'danger'"
                         :value="l.status === 'ok' ? 'TERKIRIM' : 'GAGAL'" class="!text-[10px]" />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <template v-if="logs.length" #footer>
            <div class="flex items-center justify-between p-3 border-t border-slate-800 text-xs text-slate-400">
              <span>{{ logsPage * logsPerPage + 1 }}–{{ Math.min(logsTotal, (logsPage + 1) * logsPerPage) }} dari {{ logsTotal }} log</span>
              <div class="flex items-center gap-2">
                <Button icon="pi pi-angle-left" size="small" text severity="secondary"
                        :disabled="logsPage === 0" @click="logsPage--; fetchLogs()" />
                <span class="font-mono">Hal. {{ logsPage + 1 }} / {{ lastLogPage + 1 }}</span>
                <Button icon="pi pi-angle-right" size="small" text severity="secondary"
                        :disabled="logsPage >= lastLogPage" @click="logsPage++; fetchLogs()" />
              </div>
            </div>
          </template>
        </Panel>
      </div>

      <!-- ============ TAB 6: SETUP GUIDE ============ -->
      <div v-if="activeTab === 'guide'" class="grid md:grid-cols-2 gap-5">
        <Panel title="Langkah 1: Setup Bot & Token di @BotFather" icon="pi pi-shield">
          <div class="space-y-3.5 text-xs text-slate-300 leading-relaxed">
            <div class="p-3 rounded-xl bg-slate-950/70 border border-slate-800 space-y-1.5">
              <div class="font-bold text-emerald-400">1. Buat Bot Baru</div>
              <p class="text-slate-400">Buka chat dengan <strong class="text-white">@BotFather</strong> di Telegram, kirim perintah:</p>
              <div class="flex items-center justify-between p-2 rounded bg-slate-900 font-mono text-emerald-400 text-xs">
                <span>/newbot</span>
                <Button icon="pi pi-copy" text size="small" @click="copyText('/newbot')" />
              </div>
            </div>

            <div class="p-3 rounded-xl bg-slate-950/70 border border-slate-800 space-y-1.5">
              <div class="font-bold text-emerald-400">2. Simpan Bot Token</div>
              <p class="text-slate-400">Salin token dari BotFather dan masukkan ke menu <strong>Koneksi & Mini App</strong>, lalu klik Simpan.</p>
            </div>
          </div>
        </Panel>

        <Panel title="Langkah 2: Konfigurasi Mini App Menu Button" icon="pi pi-mobile">
          <div class="space-y-3.5 text-xs text-slate-300 leading-relaxed">
            <div class="p-3 rounded-xl bg-slate-950/70 border border-slate-800 space-y-2">
              <div class="font-bold text-sky-400">Atur Menu Button di BotFather</div>
              <p class="text-slate-400">Kirim perintah ini ke @BotFather untuk memunculkan tombol menu Web App di pojok kiri bawah chat bot:</p>
              <div class="flex items-center justify-between p-2 rounded bg-slate-900 font-mono text-sky-400 text-xs">
                <span>/setmenubutton</span>
                <Button icon="pi pi-copy" text size="small" @click="copyText('/setmenubutton')" />
              </div>
              <p class="text-slate-400 text-[11.5px]">Lalu pilih bot Anda, masukkan URL Web App:</p>
              <div class="flex items-center justify-between p-2 rounded bg-slate-900 font-mono text-slate-300 text-xs">
                <span>https://barang.kesling.biz.id</span>
                <Button icon="pi pi-copy" text size="small" @click="copyText('https://barang.kesling.biz.id')" />
              </div>
              <p class="text-slate-400 text-[11.5px]">Dan beri judul tombol, misalnya: <strong class="text-white">Buka Inventaris</strong></p>
            </div>
          </div>
        </Panel>
      </div>
    </div>

    <!-- Dialog Edit/Add Command -->
    <Dialog v-model:visible="showCmdForm" modal :header="editingCmd?.id ? 'Edit Perintah Bot' : 'Tambah Perintah Baru'"
            :style="{ width: '480px' }" class="p-fluid">
      <div v-if="editingCmd" class="space-y-3.5 pt-2 text-xs">
        <Field label="Command (tanpa slash)" hint="Contoh: stok, cari, kontak">
          <InputText v-model="editingCmd.command" placeholder="stok" class="font-mono text-xs" />
        </Field>
        <Field label="Label Menu" hint="Teks yang muncul pada tombol menu">
          <InputText v-model="editingCmd.label" placeholder="Cek Stok" class="text-xs" />
        </Field>
        <Field label="Deskripsi">
          <InputText v-model="editingCmd.description" placeholder="Menampilkan daftar stok barang menipis" class="text-xs" />
        </Field>
        <Field label="Respon Teks (Opsional)">
          <Textarea v-model="editingCmd.response" rows="3" placeholder="Pesan otomatis bot..." class="text-xs" />
        </Field>
        <div class="grid grid-cols-2 gap-3 pt-1">
          <Field label="Tampilkan di Menu">
            <ToggleButton v-model="editingCmd.is_menu" onLabel="Ya" offLabel="Tidak" />
          </Field>
          <Field label="Status Aktif">
            <ToggleButton v-model="editingCmd.active" onLabel="Aktif" offLabel="Mati" />
          </Field>
        </div>
      </div>
      <template #footer>
        <Button label="Batal" icon="pi pi-times" text size="small" severity="secondary" @click="showCmdForm = false" />
        <Button label="Simpan" icon="pi pi-check" size="small" :loading="cmdSaving" @click="saveCmd" />
      </template>
    </Dialog>
  </div>
</template>
