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
const refreshingBot = ref(false)

const settings = ref<Record<string, string>>({})
const botInfo = ref<any>(null)
const subscribers = ref<any[]>([])
const logs = ref<any[]>([])
const logsTotal = ref(0)
const logsPage = ref(0)
const logsPerPage = ref(25)

const newChatID = ref('')
const newChatTitle = ref('')

const testMsg = ref('')

// ---- computed ----
const tokenConfigured = computed(() => !!settings.value.telegram_bot_token)
const botConnected = computed(() => botInfo.value?.configured && !botInfo.value?.error)

// ---- fetch ----
async function fetchAll() {
  loading.value = true
  try {
    await Promise.all([fetchSettings(), fetchBotInfo(), fetchSubscribers(), fetchLogs(), fetchCommands()])
  } finally { loading.value = false }
}

async function fetchSettings() {
  const res = await api.get('/settings/telegram')
  settings.value = res.data
}

async function fetchBotInfo() {
  try {
    const res = await api.get('/settings/telegram/bot')
    botInfo.value = res.data
  } catch { botInfo.value = { configured: false } }
}

async function fetchSubscribers() {
  const res = await api.get('/settings/telegram/subscribers')
  subscribers.value = res.data
}

async function fetchLogs() {
  const res = await api.get('/settings/telegram/logs', { params: { page: logsPage.value + 1, per_page: logsPerPage.value } })
  logs.value = res.data.data
  logsTotal.value = res.data.total
}

onMounted(fetchAll)

// ---- actions ----
async function saveToken() {
  saving.value = true
  try {
    await api.put('/settings/telegram', {
      telegram_bot_token: settings.value.telegram_bot_token,
      telegram_alert_low_stock: settings.value.telegram_alert_low_stock || 'false',
      telegram_alert_daily_time: settings.value.telegram_alert_daily_time || '07:00',
    })
    toast.add({ severity: 'success', summary: 'Pengaturan disimpan', life: 2500 })
    await fetchBotInfo()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal menyimpan', detail: e.response?.data?.error, life: 4000 })
  } finally { saving.value = false }
}

async function refreshBot() {
  refreshingBot.value = true
  try { await fetchBotInfo() } finally { refreshingBot.value = false }
}

async function addSubscriber() {
  const chatID = parseInt(newChatID.value)
  if (!chatID || chatID <= 0) {
    toast.add({ severity: 'warn', summary: 'Chat ID tidak valid', life: 2500 })
    return
  }
  try {
    await api.post('/settings/telegram/subscribers', { chat_id: chatID, title: newChatTitle.value })
    toast.add({ severity: 'success', summary: 'Subscriber ditambahkan', life: 2500 })
    newChatID.value = ''
    newChatTitle.value = ''
    await fetchSubscribers()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal menambah', detail: e.response?.data?.error, life: 4000 })
  }
}

async function toggleSub(sub: any, field: string, val: boolean) {
  try {
    await api.put(`/settings/telegram/subscribers/${sub.chat_id}`, { [field]: val })
    sub[field] = val
  } catch {
    toast.add({ severity: 'error', summary: 'Gagal update', life: 2500 })
  }
}

function removeSub(sub: any) {
  confirm.require({
    message: `Hapus subscriber "${sub.title}" (${sub.chat_id})?`,
    header: 'Konfirmasi hapus',
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
      toast.add({ severity: 'error', summary: 'Gagal', detail: r?.error || 'Unknown', life: 5000 })
    }
    await fetchLogs()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Test gagal', detail: e.response?.data?.error, life: 5000 })
  } finally { testing.value = false }
}

async function testAll() {
  testingAll.value = true
  try {
    const res = await api.post('/settings/telegram/test', { all: true, message: testMsg.value })
    const results = res.data.results || []
    const ok = results.filter((r: any) => r.status === 'ok').length
    const fail = results.filter((r: any) => r.status === 'fail').length
    if (fail === 0 && ok > 0) {
      toast.add({ severity: 'success', summary: 'Semua terkirim', detail: `${ok} chat`, life: 3000 })
    } else if (ok > 0) {
      toast.add({ severity: 'warn', summary: 'Sebagian', detail: `${ok} ok, ${fail} gagal`, life: 4000 })
    } else {
      toast.add({ severity: 'error', summary: 'Semua gagal', life: 5000 })
    }
    await fetchLogs()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Test gagal', detail: e.response?.data?.error, life: 5000 })
  } finally { testingAll.value = false }
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
  is_menu: false, sort_order: 0, active: true,
})

async function fetchCommands() {
  cmdLoading.value = true
  try {
    const res = await api.get('/settings/telegram/commands')
    commands.value = res.data
  } catch { toast.add({ severity: 'error', summary: 'Gagal memuat commands', life: 2500 }) }
  finally { cmdLoading.value = false }
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
    toast.add({ severity: 'warn', summary: 'Command dan label wajib', life: 2500 })
    return
  }
  cmdSaving.value = true
  try {
    if (editingCmd.value.id) {
      await api.put(`/settings/telegram/commands/${editingCmd.value.id}`, editingCmd.value)
    } else {
      await api.post('/settings/telegram/commands', editingCmd.value)
    }
    toast.add({ severity: 'success', summary: 'Command disimpan', life: 2500 })
    showCmdForm.value = false
    await fetchCommands()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal menyimpan', detail: e.response?.data?.error, life: 4000 })
  } finally { cmdSaving.value = false }
}

function removeCmd(c: any) {
  confirm.require({
    message: `Hapus command "/${c.command}"?`,
    header: 'Konfirmasi hapus',
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
  <div>
    <PageHeader crumb="Administrasi" title="Konfigurasi Telegram"
      sub="Bot notifikasi, subscriber, dan pengaturan alert">
      <template #actions>
        <Button label="Refresh Bot" icon="pi pi-sync" size="small" text severity="secondary"
                :loading="refreshingBot" @click="refreshBot" />
        <Button label="Simpan" icon="pi pi-save" size="small"
                :loading="saving" :disabled="loading" @click="saveToken" />
      </template>
    </PageHeader>

    <div v-if="loading" class="grid place-items-center py-20">
      <i class="pi pi-spin pi-spinner text-xl" style="color: var(--txt-dim)" />
    </div>

    <div v-else class="flex flex-col gap-4">
      <!-- ============ BOT IDENTITY ============ -->
      <div class="grid lg:grid-cols-[1fr_320px] gap-4 items-start">
        <div class="flex flex-col gap-4">
          <!-- token + settings -->
          <Panel title="Bot Token & Alert" icon="pi pi-send">
            <div class="flex flex-col gap-4">
              <Field label="Bot Token" hint="Dari @BotFather. Simpan untuk fetch identitas bot.">
                <InputText v-model="settings.telegram_bot_token"
                           :placeholder="settings.telegram_bot_token ? '••••••••••••••••' : '123456:ABC-DEF...'"
                           class="w-full" fluid type="password" />
              </Field>
              <div class="grid sm:grid-cols-2 gap-4">
                <Field label="Waktu Daily Alert (WIB)" hint="Format 24 jam">
                  <InputText v-model="settings.telegram_alert_daily_time" placeholder="07:00" class="w-full" fluid />
                </Field>
                <div class="flex items-center justify-between gap-3 pt-6">
                  <div>
                    <div class="text-[13px] font-semibold">Daily Low-Stock Alert</div>
                    <div class="text-[11.5px]" style="color: var(--txt-dim)">Kirim ringkasan harian</div>
                  </div>
                  <ToggleButton v-model="settings.telegram_alert_low_stock"
                                :modelValue="settings.telegram_alert_low_stock === 'true'"
                                @update:modelValue="settings.telegram_alert_low_stock = $event ? 'true' : 'false'"
                                onLabel="Aktif" offLabel="Mati" onIcon="pi pi-check" offIcon="pi pi-times" />
                </div>
              </div>
            </div>
          </Panel>

          <!-- subscriber table -->
          <Panel title="Subscriber / Pengguna Bot" icon="pi pi-users" dense>
            <template #actions>
              <Tag severity="secondary" :value="subscribers.length + ' subscriber'" />
            </template>

            <!-- add form -->
            <div class="px-4 py-3 border-b" style="border-color: var(--line-soft)">
              <div class="flex flex-wrap gap-2.5 items-end">
                <Field label="Chat ID">
                  <InputText v-model="newChatID" placeholder="mis. -1001234567890" class="w-[220px]" fluid />
                </Field>
                <Field label="Nama (opsional)">
                  <InputText v-model="newChatTitle" placeholder="mis. Grup Puskesmas" class="w-[200px]" fluid />
                </Field>
                <Button label="Tambah" icon="pi pi-plus" size="small" @click="addSubscriber" />
              </div>
            </div>

            <EmptyState v-if="!subscribers.length" icon="pi pi-users" title="Belum ada subscriber"
                        sub="Tambahkan chat ID, atau minta user /start ke bot." />

            <div v-else class="overflow-x-auto">
              <table class="w-full text-[12.5px]">
                <thead>
                  <tr class="text-left" style="background: var(--paper-2)">
                    <th class="t-label px-4 py-2.5">Chat</th>
                    <th class="t-label px-4 py-2.5">Tipe</th>
                    <th class="t-label px-4 py-2.5 text-center">Masuk</th>
                    <th class="t-label px-4 py-2.5 text-center">Keluar</th>
                    <th class="t-label px-4 py-2.5 text-center">Opname</th>
                    <th class="t-label px-4 py-2.5 text-center">Low Stock</th>
                    <th class="t-label px-4 py-2.5 text-center">Aktif</th>
                    <th class="t-label px-4 py-2.5 w-[100px] text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody class="divide-y" style="border-color: var(--line-soft)">
                  <tr v-for="s in subscribers" :key="s.chat_id"
                      class="hover:bg-paper-2 transition-colors">
                    <td class="px-4 py-2.5">
                      <div class="font-semibold truncate">{{ s.title }}</div>
                      <div class="t-mono" style="color: var(--txt-dim)">{{ s.chat_id }}</div>
                      <div v-if="s.username" class="t-mono text-[10px]" style="color: var(--txt-dim)">@{{ s.username }}</div>
                    </td>
                    <td class="px-4 py-2.5">
                      <Tag :severity="s.type === 'private' ? 'info' : 'warn'" :value="s.type.toUpperCase()" />
                    </td>
                    <td class="px-4 py-2.5 text-center">
                      <ToggleButton :modelValue="s.notify_in" @update:modelValue="toggleSub(s, 'notify_in', $event)"
                                    onLabel="" offLabel="" onIcon="pi pi-check" offIcon="pi pi-times"
                                    class="!p-1" />
                    </td>
                    <td class="px-4 py-2.5 text-center">
                      <ToggleButton :modelValue="s.notify_out" @update:modelValue="toggleSub(s, 'notify_out', $event)"
                                    onLabel="" offLabel="" onIcon="pi pi-check" offIcon="pi pi-times"
                                    class="!p-1" />
                    </td>
                    <td class="px-4 py-2.5 text-center">
                      <ToggleButton :modelValue="s.notify_adjust" @update:modelValue="toggleSub(s, 'notify_adjust', $event)"
                                    onLabel="" offLabel="" onIcon="pi pi-check" offIcon="pi pi-times"
                                    class="!p-1" />
                    </td>
                    <td class="px-4 py-2.5 text-center">
                      <ToggleButton :modelValue="s.notify_low_stock" @update:modelValue="toggleSub(s, 'notify_low_stock', $event)"
                                    onLabel="" offLabel="" onIcon="pi pi-check" offIcon="pi pi-times"
                                    class="!p-1" />
                    </td>
                    <td class="px-4 py-2.5 text-center">
                      <ToggleButton :modelValue="s.active" @update:modelValue="toggleSub(s, 'active', $event)"
                                    onLabel="" offLabel="" onIcon="pi pi-check" offIcon="pi pi-times"
                                    class="!p-1" />
                    </td>
                    <td class="px-4 py-2.5 text-right">
                      <Button icon="pi pi-trash" text rounded size="small" severity="danger"
                              v-tooltip.top="'Hapus subscriber'" @click="removeSub(s)" />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </Panel>

          <!-- test -->
          <Panel title="Test Notifikasi" icon="pi pi-send">
            <div class="flex flex-col gap-3">
              <Field label="Pesan Test (opsional)">
                <InputText v-model="testMsg" placeholder="Default: pesan test bawaan" class="w-full" fluid />
              </Field>
              <div class="flex gap-2.5">
                <Button label="Test Semua Aktif" icon="pi pi-send" size="small"
                        :loading="testingAll" :disabled="!subscribers.length" @click="testAll" />
                <Button label="Refresh Log" icon="pi pi-sync" size="small" text severity="secondary"
                        @click="fetchLogs" />
              </div>
            </div>
          </Panel>

          <!-- logs -->
          <Panel title="Riwayat Pesan" icon="pi pi-history" dense>
            <template #actions>
              <Tag severity="secondary" :value="logsTotal + ' log'" />
            </template>

            <EmptyState v-if="!logs.length" icon="pi pi-history" title="Belum ada log"
                        sub="Pesan test akan muncul di sini." />

            <div v-else class="overflow-x-auto">
              <table class="w-full text-[12.5px]">
                <thead>
                  <tr class="text-left" style="background: var(--paper-2)">
                    <th class="t-label px-4 py-2.5 w-[160px]">Waktu</th>
                    <th class="t-label px-4 py-2.5">Tujuan</th>
                    <th class="t-label px-4 py-2.5">Pesan</th>
                    <th class="t-label px-4 py-2.5 w-[90px]">Status</th>
                  </tr>
                </thead>
                <tbody class="divide-y" style="border-color: var(--line-soft)">
                  <tr v-for="l in logs" :key="l.id" class="hover:bg-paper-2 transition-colors">
                    <td class="px-4 py-2.5 whitespace-nowrap t-mono" style="color: var(--txt-dim)">{{ l.sent_at }}</td>
                    <td class="px-4 py-2.5">
                      <div class="font-semibold truncate">{{ l.chat_title }}</div>
                      <div class="t-mono text-[10px]" style="color: var(--txt-dim)">{{ l.chat_id }}</div>
                    </td>
                    <td class="px-4 py-2.5 truncate max-w-[300px]">{{ l.message }}</td>
                    <td class="px-4 py-2.5">
                      <Tag :severity="l.status === 'ok' ? 'success' : 'danger'"
                           :value="l.status === 'ok' ? 'TERKIRIM' : 'GAGAL'" />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <template v-if="logs.length" #footer>
              <div class="flex items-center justify-between">
                <span class="text-[11.5px]" style="color: var(--txt-dim)">
                  {{ logsPage * logsPerPage + 1 }}–{{ Math.min(logsTotal, (logsPage + 1) * logsPerPage) }} dari {{ logsTotal }}
                </span>
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

        <!-- right sidebar -->
        <div class="flex flex-col gap-4 lg:sticky lg:top-4">
          <!-- bot identity -->
          <Panel title="Identitas Bot" icon="pi pi-id-card">
            <div v-if="!tokenConfigured" class="text-center py-4">
              <i class="pi pi-exclamation-triangle text-2xl" style="color: var(--sig-warn)" />
              <div class="text-[13px] font-semibold mt-2">Token belum diisi</div>
              <div class="text-[11.5px] mt-1" style="color: var(--txt-dim)">Isi bot token lalu Simpan.</div>
            </div>
            <div v-else-if="botInfo?.error" class="text-center py-4">
              <i class="pi pi-times-circle text-2xl text-sig-bad" />
              <div class="text-[13px] font-semibold mt-2">Bot tidak terhubung</div>
              <div class="text-[11.5px] mt-1" style="color: var(--txt-dim)">{{ botInfo.error }}</div>
            </div>
            <div v-else-if="botConnected" class="flex flex-col gap-3">
              <div class="flex items-center gap-3">
                <div class="w-12 h-12 rounded-full grid place-items-center text-[20px] font-bold"
                     style="background: var(--p-primary-500); color: var(--ink-950)">
                  {{ botInfo.first_name?.[0] || 'B' }}
                </div>
                <div class="min-w-0">
                  <div class="text-[14px] font-bold truncate">{{ botInfo.first_name }} {{ botInfo.last_name }}</div>
                  <div class="t-mono text-[11px]" style="color: var(--txt-dim)">@{{ botInfo.username }}</div>
                  <div class="t-mono text-[10px]" style="color: var(--txt-dim)">ID: {{ botInfo.id }}</div>
                </div>
              </div>
              <div class="pt-3 border-t flex flex-col gap-2" style="border-color: var(--line)">
                <div class="flex items-center justify-between">
                  <span class="t-label">Join Groups</span>
                  <Tag :severity="botInfo.can_join_groups ? 'success' : 'danger'"
                       :value="botInfo.can_join_groups ? 'YA' : 'TIDAK'" />
                </div>
                <div class="flex items-center justify-between">
                  <span class="t-label">Read All Msgs</span>
                  <Tag :severity="botInfo.can_read_all ? 'success' : 'danger'"
                       :value="botInfo.can_read_all ? 'YA' : 'TIDAK'" />
                </div>
                <div class="flex items-center justify-between">
                  <span class="t-label">Inline Query</span>
                  <Tag :severity="botInfo.supports_inline ? 'success' : 'secondary'"
                       :value="botInfo.supports_inline ? 'YA' : 'TIDAK'" />
                </div>
              </div>
            </div>
            <div v-else class="text-center py-4">
              <i class="pi pi-spin pi-spinner text-xl" style="color: var(--txt-dim)" />
            </div>
          </Panel>

          <!-- bot commands editor -->
          <Panel title="Daftar Command" icon="pi pi-list" dense>
            <template #actions>
              <Tag severity="secondary" :value="commands.length + ' command'" />
              <Button label="Tambah" icon="pi pi-plus" size="small" text @click="addCmd" />
            </template>

            <div v-if="cmdLoading" class="grid place-items-center py-8">
              <i class="pi pi-spin pi-spinner text-lg" style="color: var(--txt-dim)" />
            </div>

            <EmptyState v-else-if="!commands.length" icon="pi pi-list" title="Belum ada command"
                        sub="Tambah command untuk bot Telegram." />

            <div v-else class="overflow-x-auto">
              <table class="w-full text-[12.5px]">
                <thead>
                  <tr class="text-left" style="background: var(--paper-2)">
                    <th class="t-label px-4 py-2.5">Command</th>
                    <th class="t-label px-4 py-2.5">Label</th>
                    <th class="t-label px-4 py-2.5 w-[80px] text-center">Menu</th>
                    <th class="t-label px-4 py-2.5 w-[70px] text-center">Urut</th>
                    <th class="t-label px-4 py-2.5 w-[70px] text-center">Aktif</th>
                    <th class="t-label px-4 py-2.5 w-[80px] text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody class="divide-y" style="border-color: var(--line-soft)">
                  <tr v-for="c in commands" :key="c.id" class="hover:bg-paper-2 transition-colors">
                    <td class="px-4 py-2.5">
                      <span class="t-mono font-bold">/{{ c.command }}</span>
                      <div class="text-[10.5px]" style="color: var(--txt-dim)">{{ c.description }}</div>
                    </td>
                    <td class="px-4 py-2.5">{{ c.label }}</td>
                    <td class="px-4 py-2.5 text-center">
                      <Tag :severity="c.is_menu ? 'success' : 'secondary'" :value="c.is_menu ? 'YA' : '-'" />
                    </td>
                    <td class="px-4 py-2.5 text-center t-num">{{ c.sort_order }}</td>
                    <td class="px-4 py-2.5 text-center">
                      <Tag :severity="c.active ? 'success' : 'danger'" :value="c.active ? 'AKTIF' : 'MATI'" />
                    </td>
                    <td class="px-4 py-2.5 text-right">
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

          <!-- setup guide -->
          <Panel title="Panduan Setup" icon="pi pi-book" dense>
            <div class="flex flex-col gap-3 px-4 py-1 text-[12px] leading-relaxed">

              <div class="flex flex-col gap-1">
                <div class="font-bold text-[12.5px] flex items-center gap-1.5">
                  <span class="grid place-items-center w-5 h-5 rounded-full text-[10px] font-bold"
                        style="background: var(--acc-500); color: var(--ink-950)">1</span>
                  Buat Bot
                </div>
                <div style="color: var(--txt-dim)">
                  Buka Telegram, cari <span class="t-mono">@BotFather</span>. Kirim perintah
                  <span class="t-mono">/newbot</span>. Beri nama dan username (harus berakhiran <span class="t-mono">bot</span>).
                </div>
              </div>

              <div class="flex flex-col gap-1">
                <div class="font-bold text-[12.5px] flex items-center gap-1.5">
                  <span class="grid place-items-center w-5 h-5 rounded-full text-[10px] font-bold"
                        style="background: var(--acc-500); color: var(--ink-950)">2</span>
                  Salin Token
                </div>
                <div style="color: var(--txt-dim)">
                  BotFather mengirim token format <span class="t-mono">123456789:ABCdef...</span>.
                  Salin, isi di kolom Bot Token, klik Simpan.
                </div>
              </div>

              <div class="flex flex-col gap-1">
                <div class="font-bold text-[12.5px] flex items-center gap-1.5">
                  <span class="grid place-items-center w-5 h-5 rounded-full text-[10px] font-bold"
                        style="background: var(--acc-500); color: var(--ink-950)">3</span>
                  Verifikasi Bot
                </div>
                <div style="color: var(--txt-dim)">
                  Setelah Simpan, panel Identitas Bot menampilkan nama, username, dan ID bot.
                  Jika muncul error, periksa token.
                </div>
              </div>

              <div class="flex flex-col gap-1">
                <div class="font-bold text-[12.5px] flex items-center gap-1.5">
                  <span class="grid place-items-center w-5 h-5 rounded-full text-[10px] font-bold"
                        style="background: var(--acc-500); color: var(--ink-950)">4</span>
                  Dapatkan Chat ID
                </div>
                <div style="color: var(--txt-dim)">
                  <div class="mb-1">Untuk chat private:</div>
                  <div class="flex flex-col gap-1 pl-3">
                    <div>• Buka <span class="t-mono">@userinfobot</span> di Telegram</div>
                    <div>• Kirim pesan apa saja</div>
                    <div>• Bot membalas dengan Chat ID (angka)</div>
                  </div>
                  <div class="mt-1.5 mb-1">Untuk grup:</div>
                  <div class="flex flex-col gap-1 pl-3">
                    <div>• Tambahkan bot ke grup Telegram</div>
                    <div>• Kirim pesan di grup</div>
                    <div>• Buka <span class="t-mono">https://api.telegram.org/bot&lt;TOKEN&gt;/getUpdates</span></div>
                    <div>• Cari <span class="t-mono">"chat":{"id":-100...}</span> — itu Chat ID grup</div>
                  </div>
                </div>
              </div>

              <div class="flex flex-col gap-1">
                <div class="font-bold text-[12.5px] flex items-center gap-1.5">
                  <span class="grid place-items-center w-5 h-5 rounded-full text-[10px] font-bold"
                        style="background: var(--acc-500); color: var(--ink-950)">5</span>
                  Tambah Subscriber
                </div>
                <div style="color: var(--txt-dim)">
                  Isi Chat ID di form "Chat ID Tujuan", klik Tambah.
                  Bot otomatis mengambil info nama dan tipe chat.
                </div>
              </div>

              <div class="flex flex-col gap-1">
                <div class="font-bold text-[12.5px] flex items-center gap-1.5">
                  <span class="grid place-items-center w-5 h-5 rounded-full text-[10px] font-bold"
                        style="background: var(--acc-500); color: var(--ink-950)">6</span>
                  Atur Notifikasi
                </div>
                <div style="color: var(--txt-dim)">
                  Setiap subscriber punya 5 toggle: Masuk, Keluar, Opname, Low Stock, dan Active.
                  Klik toggle untuk on/off per subscriber.
                </div>
              </div>

              <div class="flex flex-col gap-1">
                <div class="font-bold text-[12.5px] flex items-center gap-1.5">
                  <span class="grid place-items-center w-5 h-5 rounded-full text-[10px] font-bold"
                        style="background: var(--acc-500); color: var(--ink-950)">7</span>
                  Test Koneksi
                </div>
                <div style="color: var(--txt-dim)">
                  Tulis pesan test (opsional), klik "Test Semua Aktif".
                  Cek panel Riwayat Pesan untuk status terkirim/gagal.
                </div>
              </div>

              <div class="pt-2 border-t" style="border-color: var(--line)">
                <div class="t-label mb-1.5">Catatan</div>
                <div style="color: var(--txt-dim)" class="text-[11px] leading-relaxed">
                  Bot harus ditambahkan sebagai admin di grup agar bisa kirim pesan.
                  Untuk channel, bot harus jadi admin dengan hak posting.
                  Chat ID grup selalu negatif (mis. <span class="t-mono">-1001234567890</span>).
                </div>
              </div>

            </div>
          </Panel>
        </div>
      </div>
    </div>

    <!-- command editor dialog -->
    <div v-if="showCmdForm" class="fixed inset-0 z-50 grid place-items-center p-4"
         style="background: rgb(0 0 0 / 0.6)" @click.self="showCmdForm = false">
      <div class="panel w-full max-w-[520px] overflow-hidden">
        <div class="flex items-center justify-between px-4 border-b" style="height: 48px; border-color: var(--line)">
          <div class="flex items-center gap-2">
            <i class="pi pi-list text-[13px]" style="color: var(--acc-500)" />
            <span class="text-[13px] font-bold">{{ editingCmd.id ? 'Edit Command' : 'Tambah Command' }}</span>
          </div>
          <Button icon="pi pi-times" text rounded size="small" severity="secondary" @click="showCmdForm = false" />
        </div>
        <div class="p-4 flex flex-col gap-4">
          <div class="grid sm:grid-cols-2 gap-3">
            <Field label="Command" hint="Tanpa /">
              <InputText v-model="editingCmd.command" placeholder="masuk" class="w-full" fluid />
            </Field>
            <Field label="Label" hint="Teks tombol menu">
              <InputText v-model="editingCmd.label" placeholder="Barang Masuk" class="w-full" fluid />
            </Field>
          </div>
          <Field label="Deskripsi" hint="Penjelasan singkat">
            <InputText v-model="editingCmd.description" placeholder="Catat barang masuk" class="w-full" fluid />
          </Field>
          <Field label="Response / Balasan Bot" hint="Pesan yang dikirim bot saat command dipanggil">
            <Textarea v-model="editingCmd.response" rows="4" autoResize
                      placeholder="Format: /masuk SKU jumlah" class="w-full" />
          </Field>
          <div class="grid grid-cols-3 gap-3">
            <div class="flex flex-col gap-1.5">
              <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">Tampil di Menu</span>
              <ToggleButton v-model="editingCmd.is_menu" onLabel="Ya" offLabel="Tidak"
                            onIcon="pi pi-check" offIcon="pi pi-times" />
            </div>
            <Field label="Urutan">
              <InputText v-model.number="editingCmd.sort_order" type="number" class="w-full" fluid />
            </Field>
            <div class="flex flex-col gap-1.5">
              <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">Aktif</span>
              <ToggleButton v-model="editingCmd.active" onLabel="Aktif" offLabel="Mati"
                            onIcon="pi pi-check" offIcon="pi pi-times" />
            </div>
          </div>
        </div>
        <div class="flex justify-end gap-2.5 px-4 py-3 border-t" style="border-color: var(--line)">
          <Button label="Batal" text severity="secondary" @click="showCmdForm = false" />
          <Button label="Simpan" icon="pi pi-save" :loading="cmdSaving" @click="saveCmd" />
        </div>
      </div>
    </div>
  </div>
</template>