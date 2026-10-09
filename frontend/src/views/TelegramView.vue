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
const syncing = ref(false)
const testing = ref(false)
const testingAll = ref(false)
const sendingAlertNow = ref(false)
const refreshingBot = ref(false)

const settings = ref<Record<string, string>>({})
const botInfo = ref<any>(null)
const subscribers = ref<any[]>([])
const logs = ref<any[]>([])
const logsTotal = ref(0)
const logsPage = ref(0)
const logsPerPage = ref(15)

const newChatID = ref('')
const newChatTitle = ref('')
const testCustomChatID = ref('')
const testMsg = ref('')
const clearingLogs = ref(false)

// ---- computed ----
const tokenConfigured = computed(() => !!settings.value.telegram_bot_token)
const botConnected = computed(() => botInfo.value?.configured && !botInfo.value?.error)
const activeSubscribersCount = computed(() => subscribers.value.filter(s => s.active).length)
const lastLogPage = computed(() => Math.max(0, Math.ceil(logsTotal.value / logsPerPage.value) - 1))

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
    toast.add({ severity: 'success', summary: 'Pengaturan Disimpan', detail: 'Konfigurasi bot dan alert berhasil diperbarui', life: 2500 })
    await fetchBotInfo()
    await syncTelegram(true)
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal Menyimpan', detail: e.response?.data?.error, life: 4000 })
  } finally {
    saving.value = false
  }
}

async function syncTelegram(silent = false) {
  syncing.value = true
  try {
    const res = await api.post('/settings/telegram/sync')
    if (!silent) {
      toast.add({
        severity: 'success',
        summary: 'Sinkronisasi Berhasil',
        detail: res.data?.message || 'Menu perintah dan tombol Mini App bot telah disinkronkan ke Telegram',
        life: 3000
      })
    }
  } catch (e: any) {
    if (!silent) {
      toast.add({
        severity: 'error',
        summary: 'Sinkronisasi Gagal',
        detail: e.response?.data?.error || 'Pastikan bot token sudah benar dan bot aktif',
        life: 4000
      })
    }
  } finally {
    syncing.value = false
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
    toast.add({ severity: 'success', summary: 'Subscriber Ditambahkan', life: 2500 })
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
      message: '🚨 [Broadcast Alert] Peringatan stok kritis inventaris dikirim dari Web Panel.'
    })
    toast.add({ severity: 'success', summary: 'Broadcast Alert Terkirim', life: 3000 })
    await fetchLogs()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal kirim alert', detail: e.response?.data?.error, life: 4000 })
  } finally {
    sendingAlertNow.value = false
  }
}

function openUrl(url: string) {
  if (typeof window !== 'undefined' && url) {
    window.open(url, '_blank')
  }
}

async function testCustom() {
  const cid = parseInt(testCustomChatID.value)
  if (!cid) {
    toast.add({ severity: 'warn', summary: 'Chat ID tidak valid', detail: 'Masukkan angka Chat ID tujuan uji coba', life: 2500 })
    return
  }
  await testOne(cid)
}

async function clearLogs() {
  confirm.require({
    message: 'Apakah Anda yakin ingin menghapus semua riwayat log pesan Telegram?',
    header: 'Konfirmasi Bersihkan Log',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: async () => {
      clearingLogs.value = true
      try {
        await api.delete('/settings/telegram/logs')
        toast.add({ severity: 'success', summary: 'Log Dibersihkan', detail: 'Semua log pesan berhasil dihapus', life: 2500 })
        logsPage.value = 0
        await fetchLogs()
      } catch (e: any) {
        toast.add({ severity: 'error', summary: 'Gagal Bersihkan Log', detail: e.response?.data?.error || 'Terjadi kesalahan', life: 4000 })
      } finally {
        clearingLogs.value = false
      }
    }
  })
}

function copyText(txt: string) {
  navigator.clipboard.writeText(txt)
  toast.add({ severity: 'info', summary: 'Tersalin ke Clipboard', detail: txt, life: 2000 })
}

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
    toast.add({ severity: 'error', summary: 'Gagal memuat commands', life: 2500 })
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
    await syncTelegram(true)
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
      await syncTelegram(true)
    },
  })
}
</script>

<template>
  <div>
    <PageHeader crumb="Administrasi" title="Konfigurasi Telegram"
      sub="Bot pintar, Telegram Mini App, scanner barcode, subscriber, dan notifikasi real-time">
      <template #actions>
        <Button label="Refresh Bot" icon="pi pi-sync" size="small" text severity="secondary"
                :loading="refreshingBot" @click="refreshBot" />
        <Button label="Sinkron ke Telegram" icon="pi pi-bolt" size="small" severity="warn"
                :loading="syncing" :disabled="!tokenConfigured" @click="syncTelegram(false)" />
        <Button label="Simpan Pengaturan" icon="pi pi-save" size="small"
                :loading="saving" :disabled="loading" @click="saveSettings" />
      </template>
    </PageHeader>

    <div v-if="loading" class="grid place-items-center py-20">
      <i class="pi pi-spin pi-spinner text-2xl" style="color: var(--txt-dim)" />
    </div>

    <div v-else class="flex flex-col gap-4">
      <!-- ============ MAIN 2-COLUMN LAYOUT ============ -->
      <div class="grid lg:grid-cols-[1fr_340px] gap-4 items-start">

        <!-- LEFT COLUMN: Main Controls & Tables -->
        <div class="flex flex-col gap-4">

          <!-- Token & System Settings -->
          <Panel title="Koneksi Bot & Telegram Mini App" icon="pi pi-send">
            <div class="flex flex-col gap-4">
              <Field label="Bot Token" hint="Dari @BotFather. Rahasiakan token ini. Simpan untuk verifikasi identitas bot.">
                <InputText v-model="settings.telegram_bot_token"
                           :placeholder="settings.telegram_bot_token ? '••••••••••••••••' : '123456:ABC-DEF...'"
                           class="w-full t-mono text-[12px]" fluid type="password" />
              </Field>

              <Field label="URL Telegram Mini App (Web App)" hint="URL aplikasi web yang dibuka saat tombol 'Buka Inventaris' ditekan di Telegram.">
                <div class="flex gap-2">
                  <InputText v-model="settings.telegram_webapp_url"
                             placeholder="https://barang.kesling.biz.id"
                             class="w-full t-mono text-[12px]" fluid />
                  <Button icon="pi pi-external-link" text size="small" severity="secondary"
                          v-tooltip.top="'Buka URL'"
                          @click="openUrl(settings.telegram_webapp_url || 'https://barang.kesling.biz.id')" />
                </div>
              </Field>

              <Field label="Telegram API Proxy (Opsional)" hint="Kosongkan untuk default (https://api.telegram.org). Diisi hanya jika jaringan hosting memblokir Telegram.">
                <InputText v-model="settings.telegram_api_url"
                           placeholder="https://api.telegram.org"
                           class="w-full t-mono text-[12px]" fluid />
              </Field>

              <div class="grid sm:grid-cols-2 gap-4 pt-1">
                <Field label="Waktu Daily Alert (WIB)" hint="Format 24 jam (misal 07:00)">
                  <InputText v-model="settings.telegram_alert_daily_time" placeholder="07:00" class="w-full" fluid />
                </Field>
                <div class="flex items-center justify-between gap-3 pt-6">
                  <div>
                    <div class="text-[13px] font-semibold" style="color: var(--txt)">Daily Low-Stock Alert</div>
                    <div class="text-[11.5px]" style="color: var(--txt-dim)">Kirim rekap barang menipis harian</div>
                  </div>
                  <ToggleButton v-model="settings.telegram_alert_low_stock"
                                :modelValue="settings.telegram_alert_low_stock === 'true'"
                                @update:modelValue="settings.telegram_alert_low_stock = $event ? 'true' : 'false'"
                                onLabel="Aktif" offLabel="Mati" onIcon="pi pi-check" offIcon="pi pi-times" />
                </div>
              </div>
            </div>
          </Panel>

          <!-- Subscribers Table -->
          <Panel title="Subscriber & Pengguna Bot" icon="pi pi-users" dense>
            <template #actions>
              <Tag severity="info" :value="subscribers.length + ' subscriber'" class="!text-[11px]" />
            </template>

            <!-- add subscriber form -->
            <div class="px-4 py-3 border-b" style="border-color: var(--line); background: var(--panel-2)">
              <div class="flex flex-wrap gap-2.5 items-end">
                <Field label="Chat ID (Angka)">
                  <InputText v-model="newChatID" placeholder="mis. 123456789 atau -100..." class="w-[220px] t-mono text-[12px]" fluid />
                </Field>
                <Field label="Nama / Label Penerima">
                  <InputText v-model="newChatTitle" placeholder="mis. Petugas Farmasi" class="w-[200px]" fluid />
                </Field>
                <Button label="Tambah" icon="pi pi-plus" size="small" @click="addSubscriber" />
              </div>
            </div>

            <EmptyState v-if="!subscribers.length" icon="pi pi-users" title="Belum ada subscriber"
                        sub="Tambahkan chat ID secara manual, atau minta petugas mengirimkan perintah /start ke bot." />

            <div v-else class="overflow-x-auto">
              <table class="w-full text-[12.5px]">
                <thead>
                  <tr class="text-left border-b" style="border-color: var(--line); background: var(--panel-2)">
                    <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider" style="color: var(--txt-dim)">Chat Tujuan</th>
                    <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider" style="color: var(--txt-dim)">Tipe</th>
                    <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider text-center" style="color: var(--txt-dim)">Masuk</th>
                    <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider text-center" style="color: var(--txt-dim)">Keluar</th>
                    <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider text-center" style="color: var(--txt-dim)">Opname</th>
                    <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider text-center" style="color: var(--txt-dim)">Low Stock</th>
                    <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider text-center" style="color: var(--txt-dim)">Aktif</th>
                    <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider text-right w-[110px]" style="color: var(--txt-dim)">Aksi</th>
                  </tr>
                </thead>
                <tbody class="divide-y" style="border-color: var(--line)">
                  <tr v-for="s in subscribers" :key="s.chat_id"
                      class="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors">
                    <td class="px-4 py-3">
                      <div class="font-semibold truncate" style="color: var(--txt)">{{ s.title || 'Tanpa Nama' }}</div>
                      <div class="t-mono text-[11px]" style="color: var(--txt-dim)">ID: {{ s.chat_id }}</div>
                      <div v-if="s.username" class="t-mono text-[10.5px]" style="color: var(--sig-ok)">@{{ s.username }}</div>
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

          <!-- Test & Broadcast Panel -->
          <Panel title="Uji Notifikasi & Broadcast Alert" icon="pi pi-bell">
            <div class="flex flex-col gap-3.5">
              <div class="grid sm:grid-cols-[1fr_auto] gap-2 items-end">
                <Field label="Uji Chat ID Tertentu (Langsung)" hint="Masukkan Chat ID langsung tanpa perlu terdaftar sebagai subscriber terlebih dahulu.">
                  <InputText v-model="testCustomChatID" placeholder="mis. 123456789 atau -100..." class="w-full t-mono text-[12px]" fluid />
                </Field>
                <Button label="Uji Chat ID Ini" icon="pi pi-send" size="small" severity="info"
                        :loading="testing" :disabled="!testCustomChatID || !tokenConfigured" @click="testCustom" />
              </div>

              <Field label="Pesan Uji Coba (Opsional)">
                <InputText v-model="testMsg" placeholder="Default: pesan test bawaan sistem inventaris" class="w-full" fluid />
              </Field>

              <div class="flex flex-wrap gap-2.5 pt-1">
                <Button label="Test Semua Subscriber" icon="pi pi-send" size="small"
                        :loading="testingAll" :disabled="!tokenConfigured" @click="testAll" />
                <Button label="Broadcast Alert Stok Kritis" icon="pi pi-exclamation-triangle" size="small" severity="warn"
                        :loading="sendingAlertNow" :disabled="!tokenConfigured" @click="sendAlertNow" />
                <Button label="Segarkan Log" icon="pi pi-sync" size="small" text severity="secondary"
                        @click="fetchLogs" />
              </div>
            </div>
          </Panel>

          <!-- Logs Table -->
          <Panel title="Riwayat Log Notifikasi Terkirim" icon="pi pi-history" dense>
            <template #actions>
              <div class="flex items-center gap-2">
                <Tag severity="secondary" :value="logsTotal + ' log pesan'" class="!text-[11px]" />
                <Button v-if="logs.length" label="Bersihkan Log" icon="pi pi-trash" size="small" severity="danger" text
                        :loading="clearingLogs" @click="clearLogs" />
              </div>
            </template>

            <EmptyState v-if="!logs.length" icon="pi pi-history" title="Belum ada riwayat pesan"
                        sub="Pesan notifikasi dan alert yang dikirimkan bot akan tercatat di sini." />

            <div v-else class="overflow-x-auto">
              <table class="w-full text-[12.5px]">
                <thead>
                  <tr class="text-left border-b" style="border-color: var(--line); background: var(--panel-2)">
                    <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[160px]" style="color: var(--txt-dim)">Waktu</th>
                    <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider" style="color: var(--txt-dim)">Tujuan</th>
                    <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider" style="color: var(--txt-dim)">Pesan</th>
                    <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[100px] text-right" style="color: var(--txt-dim)">Status</th>
                  </tr>
                </thead>
                <tbody class="divide-y" style="border-color: var(--line)">
                  <tr v-for="l in logs" :key="l.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors">
                    <td class="px-4 py-3 whitespace-nowrap t-mono text-[11px]" style="color: var(--txt-dim)">{{ l.sent_at }}</td>
                    <td class="px-4 py-3">
                      <div class="font-semibold truncate" style="color: var(--txt)">{{ l.chat_title || 'Chat ID: ' + l.chat_id }}</div>
                      <div class="t-mono text-[10.5px]" style="color: var(--txt-dim)">{{ l.chat_id }}</div>
                    </td>
                    <td class="px-4 py-3 truncate max-w-[320px]" style="color: var(--txt)">{{ l.message }}</td>
                    <td class="px-4 py-3 text-right">
                      <Tag :severity="l.status === 'ok' ? 'success' : 'danger'"
                           :value="l.status === 'ok' ? 'TERKIRIM' : 'GAGAL'" class="!text-[10px]" />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <template v-if="logs.length" #footer>
              <div class="flex items-center justify-between text-[11.5px]" style="color: var(--txt-dim)">
                <span>{{ logsPage * logsPerPage + 1 }}–{{ Math.min(logsTotal, (logsPage + 1) * logsPerPage) }} dari {{ logsTotal }} log</span>
                <div class="flex items-center gap-2">
                  <Button icon="pi pi-angle-left" size="small" text severity="secondary"
                          :disabled="logsPage === 0" @click="logsPage--; fetchLogs()" />
                  <span class="t-num text-[12px]">Hal. {{ logsPage + 1 }} / {{ lastLogPage + 1 }}</span>
                  <Button icon="pi pi-angle-right" size="small" text severity="secondary"
                          :disabled="logsPage >= lastLogPage" @click="logsPage++; fetchLogs()" />
                </div>
              </div>
            </template>
          </Panel>
        </div>

        <!-- RIGHT COLUMN: Bot Identity, Commands & Setup Guide -->
        <div class="flex flex-col gap-4 lg:sticky lg:top-4">

          <!-- Bot Identity Card -->
          <Panel title="Identitas Bot Telegram" icon="pi pi-id-card">
            <div v-if="!tokenConfigured" class="text-center py-4">
              <i class="pi pi-exclamation-triangle text-2xl" style="color: var(--sig-warn)" />
              <div class="text-[13px] font-semibold mt-2" style="color: var(--txt)">Token belum diisi</div>
              <div class="text-[11.5px] mt-1" style="color: var(--txt-dim)">Isi bot token lalu klik Simpan Pengaturan.</div>
            </div>
            <div v-else-if="botInfo?.error" class="text-center py-4">
              <i class="pi pi-times-circle text-2xl" style="color: var(--sig-bad)" />
              <div class="text-[13px] font-semibold mt-2" style="color: var(--txt)">Bot tidak terhubung</div>
              <div class="text-[11.5px] mt-1 break-words" style="color: var(--txt-dim)">{{ botInfo.error }}</div>
            </div>
            <div v-else-if="botConnected" class="flex flex-col gap-3">
              <div v-if="botInfo.warning" class="p-2.5 rounded-lg text-[11px] border"
                   style="border-color: var(--sig-warn); background: rgba(245, 158, 11, 0.1); color: var(--sig-warn)">
                <i class="pi pi-exclamation-triangle mr-1" />{{ botInfo.warning }}
              </div>

              <div class="flex items-center gap-3">
                <div class="w-12 h-12 rounded-xl grid place-items-center text-[18px] font-bold shadow-sm"
                     style="background: var(--acc-500); color: #ffffff">
                  {{ botInfo.first_name?.[0] || 'B' }}
                </div>
                <div class="min-w-0">
                  <div class="text-[14px] font-bold truncate" style="color: var(--txt)">
                    {{ botInfo.first_name }} {{ botInfo.last_name || '' }}
                  </div>
                  <div class="t-mono text-[11px]" style="color: var(--sig-ok)">@{{ botInfo.username }}</div>
                  <div class="t-mono text-[10px]" style="color: var(--txt-dim)">ID: {{ botInfo.id }}</div>
                </div>
              </div>

              <div class="pt-3 border-t flex flex-col gap-2 text-[12px]" style="border-color: var(--line)">
                <div class="flex items-center justify-between">
                  <span class="t-label">Status Bot</span>
                  <Tag severity="success" value="ONLINE" class="!text-[10px]" />
                </div>
                <div class="flex items-center justify-between">
                  <span class="t-label">Mini App Menu</span>
                  <Tag severity="info" value="AKTIF" class="!text-[10px]" />
                </div>
                <div class="flex items-center justify-between">
                  <span class="t-label">Join Groups</span>
                  <Tag :severity="botInfo.can_join_groups ? 'success' : 'secondary'"
                       :value="botInfo.can_join_groups ? 'YA' : 'TIDAK'" class="!text-[10px]" />
                </div>
                <div class="flex items-center justify-between">
                  <span class="t-label">Inline Query</span>
                  <Tag :severity="botInfo.supports_inline ? 'success' : 'secondary'"
                       :value="botInfo.supports_inline ? 'YA' : 'TIDAK'" class="!text-[10px]" />
                </div>
              </div>

              <div class="pt-2">
                <Button label="Buka Bot di Telegram" icon="pi pi-external-link" size="small" class="w-full" outlined
                        @click="openUrl('https://t.me/' + botInfo.username)" />
              </div>
            </div>
            <div v-else class="text-center py-4">
              <i class="pi pi-spin pi-spinner text-xl" style="color: var(--txt-dim)" />
            </div>
          </Panel>

          <!-- Bot Commands Manager -->
          <Panel title="Daftar Command & Menu" icon="pi pi-list" dense>
            <template #actions>
              <Tag severity="secondary" :value="commands.length + ' command'" class="!text-[10px]" />
              <Button icon="pi pi-sync" size="small" text rounded severity="warn"
                      v-tooltip.top="'Sinkronkan ke menu Telegram Bot API'"
                      :loading="syncing" @click="syncTelegram(false)" />
              <Button label="Tambah" icon="pi pi-plus" size="small" text @click="addCmd" />
            </template>

            <div v-if="cmdLoading" class="grid place-items-center py-8">
              <i class="pi pi-spin pi-spinner text-lg" style="color: var(--txt-dim)" />
            </div>

            <EmptyState v-else-if="!commands.length" icon="pi pi-list" title="Belum ada command"
                        sub="Tambah command untuk menu bot Telegram." />

            <div v-else class="overflow-x-auto">
              <table class="w-full text-[12px]">
                <thead>
                  <tr class="text-left border-b" style="border-color: var(--line); background: var(--panel-2)">
                    <th class="px-3 py-2.5 font-semibold text-[10.5px] uppercase tracking-wider" style="color: var(--txt-dim)">Perintah</th>
                    <th class="px-3 py-2.5 font-semibold text-[10.5px] uppercase tracking-wider" style="color: var(--txt-dim)">Label</th>
                    <th class="px-3 py-2.5 font-semibold text-[10.5px] uppercase tracking-wider text-center" style="color: var(--txt-dim)">Menu</th>
                    <th class="px-3 py-2.5 font-semibold text-[10.5px] uppercase tracking-wider text-right" style="color: var(--txt-dim)">Aksi</th>
                  </tr>
                </thead>
                <tbody class="divide-y" style="border-color: var(--line)">
                  <tr v-for="c in commands" :key="c.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors">
                    <td class="px-3 py-2.5">
                      <span class="t-mono font-bold" style="color: var(--txt)">/{{ c.command }}</span>
                      <div class="text-[10px] truncate max-w-[130px]" style="color: var(--txt-dim)">{{ c.description }}</div>
                    </td>
                    <td class="px-3 py-2.5 truncate max-w-[100px]" style="color: var(--txt)">{{ c.label }}</td>
                    <td class="px-3 py-2.5 text-center">
                      <Tag :severity="c.is_menu ? 'success' : 'secondary'" :value="c.is_menu ? 'YA' : '-'" class="!text-[9px]" />
                    </td>
                    <td class="px-3 py-2.5 text-right whitespace-nowrap">
                      <Button icon="pi pi-pencil" text rounded size="small" severity="secondary"
                              v-tooltip.top="'Edit'" @click="editCmd(c)" />
                      <Button icon="pi pi-trash" text rounded size="small" severity="danger"
                              v-tooltip.top="'Hapus'" @click="removeCmd(c)" />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </Panel>

          <!-- Interactive Setup & Mini App Guide -->
          <Panel title="Panduan Setup & Mini App" icon="pi pi-book" dense>
            <div class="flex flex-col gap-3 px-4 py-3 text-[12px] leading-relaxed">

              <div class="flex flex-col gap-1">
                <div class="font-bold flex items-center gap-1.5" style="color: var(--txt)">
                  <span class="grid place-items-center w-5 h-5 rounded-full text-[10px] font-bold"
                        style="background: var(--acc-500); color: #ffffff">1</span>
                  Buat Bot Telegram
                </div>
                <div style="color: var(--txt-dim)">
                  Buka chat dengan <span class="t-mono font-semibold" style="color: var(--txt)">@BotFather</span>. Kirim
                  <span class="t-mono">/newbot</span>, tentukan nama dan username bot.
                </div>
              </div>

              <div class="flex flex-col gap-1">
                <div class="font-bold flex items-center gap-1.5" style="color: var(--txt)">
                  <span class="grid place-items-center w-5 h-5 rounded-full text-[10px] font-bold"
                        style="background: var(--acc-500); color: #ffffff">2</span>
                  Atur Tombol Menu Mini App
                </div>
                <div style="color: var(--txt-dim)">
                  Kirim perintah ke @BotFather untuk tombol Web App pojok kiri bawah:
                </div>
                <div class="flex items-center justify-between p-2 rounded border text-[11px] t-mono"
                     style="border-color: var(--line); background: var(--panel-2)">
                  <span>/setmenubutton</span>
                  <Button icon="pi pi-copy" text size="small" @click="copyText('/setmenubutton')" />
                </div>
                <div style="color: var(--txt-dim)" class="text-[11px]">
                  Masukkan URL: <span class="t-mono font-semibold" style="color: var(--txt)">https://barang.kesling.biz.id</span>
                  dan judul: <strong style="color: var(--txt)">Buka Inventaris</strong>.
                </div>
              </div>

              <div class="flex flex-col gap-1">
                <div class="font-bold flex items-center gap-1.5" style="color: var(--txt)">
                  <span class="grid place-items-center w-5 h-5 rounded-full text-[10px] font-bold"
                        style="background: var(--acc-500); color: #ffffff">3</span>
                  Dapatkan Chat ID
                </div>
                <div style="color: var(--txt-dim)">
                  Buka chat bot Anda, kirim <span class="t-mono font-bold" style="color: var(--txt)">/start</span>. Bot akan menampilkan Chat ID. Salin lalu tambahkan pada tabel subscriber.
                </div>
              </div>

              <div class="pt-2 border-t text-[11px]" style="border-color: var(--line); color: var(--txt-dim)">
                💡 Bot sudah dilengkapi scanner barcode foto, pencarian cerdas, dan pencatatan mutasi tanpa perlu membuka web view.
              </div>
            </div>
          </Panel>
        </div>
      </div>
    </div>

    <!-- Dialog Edit/Add Command -->
    <Dialog v-model:visible="showCmdForm" modal :header="editingCmd?.id ? 'Edit Command' : 'Tambah Command'"
            :style="{ width: '480px' }" class="p-fluid">
      <div v-if="editingCmd" class="flex flex-col gap-3 pt-2">
        <Field label="Command (tanpa tanda /)" hint="Contoh: stok, cari, kontak, jadwal">
          <InputText v-model="editingCmd.command" placeholder="stok" class="t-mono" />
        </Field>
        <Field label="Label Tombol Menu" hint="Teks singkat yang muncul pada tombol">
          <InputText v-model="editingCmd.label" placeholder="Cek Stok" />
        </Field>
        <Field label="Deskripsi Perintah">
          <InputText v-model="editingCmd.description" placeholder="Menampilkan daftar stok barang menipis" />
        </Field>
        <Field label="Respon Pesan Otomatis (Opsional)">
          <Textarea v-model="editingCmd.response" rows="3" placeholder="Pesan balasan bot..." />
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
