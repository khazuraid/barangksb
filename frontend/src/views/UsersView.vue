<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import { useConfirm } from 'primevue/useconfirm'
import { useAuthStore } from '@/stores/auth'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Select from 'primevue/select'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import StatusChip from '@/components/StatusChip.vue'
import EmptyState from '@/components/EmptyState.vue'
import Field from '@/components/Field.vue'
import Tag from 'primevue/tag'

const toast = useToast()
const confirm = useConfirm()
const auth = useAuthStore()

const users = ref<any[]>([])
const loading = ref(true)
const saving = ref(false)
const showForm = ref(false)
const q = ref('')

const form = ref({ name: '', email: '', password: '', role: 'petugas' })
const roleOptions = [
  { label: 'Petugas', value: 'petugas' },
  { label: 'Administrator', value: 'admin' },
]

async function fetchUsers() {
  loading.value = true
  try { users.value = (await api.get('/users')).data } finally { loading.value = false }
}
onMounted(fetchUsers)

const filtered = computed(() =>
  users.value.filter(u => !q.value || [u.name, u.email, u.role].join(' ').toLowerCase().includes(q.value.toLowerCase()))
)
const stats = computed(() => ({
  total: users.value.length,
  admin: users.value.filter(u => u.role === 'admin').length,
  petugas: users.value.filter(u => u.role === 'petugas').length,
}))

function openCreate() {
  form.value = { name: '', email: '', password: '', role: 'petugas' }
  showForm.value = true
}

async function save() {
  saving.value = true
  try {
    await api.post('/users', form.value)
    toast.add({ severity: 'success', summary: 'Tersimpan', detail: `${form.value.email} dibuat/diperbarui`, life: 2500 })
    showForm.value = false
    fetchUsers()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal', detail: e.response?.data?.error, life: 3500 })
  } finally { saving.value = false }
}

async function changeRole(u: any, role: string) {
  if (role === u.role) return
  try {
    await api.put(`/users/${encodeURIComponent(u.email)}/role`, { role })
    toast.add({ severity: 'success', summary: 'Role diperbarui', detail: `${u.email} → ${role}`, life: 2500 })
    fetchUsers()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal', detail: e.response?.data?.error, life: 3500 })
    fetchUsers()
  }
}

function remove(u: any) {
  confirm.require({
    message: `Hapus akun "${u.name}" (${u.email})? Tindakan ini tidak dapat dibatalkan.`,
    header: 'Konfirmasi hapus pengguna',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    acceptLabel: 'Hapus',
    rejectLabel: 'Batal',
    accept: async () => {
      try {
        await api.delete(`/users/${encodeURIComponent(u.email)}`)
        toast.add({ severity: 'success', summary: 'Dihapus', detail: u.email, life: 2500 })
        fetchUsers()
      } catch (e: any) {
        toast.add({ severity: 'error', summary: 'Gagal', detail: e.response?.data?.error, life: 3500 })
      }
    },
  })
}
</script>

<template>
  <div>
    <PageHeader crumb="Administrasi" title="Pengguna"
      sub="Kelola akun petugas dan administrator beserta hak aksesnya">
      <template #actions>
        <Button label="Tambah Pengguna" icon="pi pi-user-plus" size="small" @click="openCreate" />
      </template>
    </PageHeader>

    <!-- summary -->
    <div class="grid grid-cols-3 gap-px mb-4 rounded-lg overflow-hidden border max-w-lg"
         style="border-color: var(--line); background: var(--line)">
      <div v-for="c in [
        { k: 'Total Akun', v: stats.total },
        { k: 'Administrator', v: stats.admin },
        { k: 'Petugas', v: stats.petugas },
      ]" :key="c.k" class="px-4 py-3" style="background: var(--panel)">
        <div class="t-label">{{ c.k }}</div>
        <div class="t-num text-[20px] font-bold mt-1">{{ c.v }}</div>
      </div>
    </div>

    <Panel title="Daftar Akun" icon="pi pi-users" dense>
      <template #actions>
        <InputText v-model="q" placeholder="Cari nama atau email…" class="!text-[12px] !py-1.5 w-[220px]" />
      </template>

      <EmptyState v-if="loading" icon="pi pi-spin pi-spinner" title="Memuat pengguna…" />
      <EmptyState v-else-if="!filtered.length" icon="pi pi-users" title="Tidak ada pengguna"
                    sub="Tambahkan akun baru untuk memberi akses ke sistem." />

      <div v-else class="overflow-x-auto">
        <table class="w-full text-[12.5px]">
          <thead>
            <tr class="text-left" style="background: var(--paper-2)">
              <th class="t-label px-4 py-2.5">Nama</th>
              <th class="t-label px-4 py-2.5">Email</th>
              <th class="t-label px-4 py-2.5 w-[180px]">Role</th>
              <th class="t-label px-4 py-2.5 w-[110px] text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y" style="border-color: var(--line-soft)">
            <tr v-for="u in filtered" :key="u.email" class="hover:bg-paper-2 transition-colors">
              <td class="px-4 py-2.5">
                <div class="flex items-center gap-2.5">
                  <div class="w-7 h-7 shrink-0 grid place-items-center rounded-md border t-num font-bold text-[11px]"
                       style="border-color: var(--line); background: var(--paper-2)">
                    {{ u.name?.[0]?.toUpperCase() || '?' }}
                  </div>
                  <span class="font-semibold">{{ u.name }}</span>
                  <Tag v-if="u.email === auth.user?.email" severity="info" value="ANDA" />
                </div>
              </td>
              <td class="px-4 py-2.5" style="color: var(--txt-dim)">{{ u.email }}</td>
              <td class="px-4 py-2.5">
                <Select
                  :modelValue="u.role"
                  :options="roleOptions"
                  optionLabel="label"
                  optionValue="value"
                  class="!text-[12px] !py-1 w-[160px]"
                  :disabled="u.email === auth.user?.email"
                  @update:modelValue="(v: any) => changeRole(u, v)"
                />
              </td>
              <td class="px-4 py-2.5 text-right">
                <Button icon="pi pi-trash" text rounded size="small" severity="danger"
                        v-tooltip.top="u.email === auth.user?.email ? 'Tidak dapat menghapus akun sendiri' : 'Hapus akun'"
                        :disabled="u.email === auth.user?.email"
                        @click="remove(u)" />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </Panel>

    <!-- create dialog -->
    <Dialog v-model:visible="showForm" modal header="Tambah / Perbarui Pengguna" :style="{ width: '440px' }">
      <form @submit.prevent="save" class="grid sm:grid-cols-2 gap-3.5">
        <Field label="Nama Lengkap" required span>
          <InputText v-model="form.name" required placeholder="mis. Budi Santoso" class="w-full" />
        </Field>
        <Field label="Email" required span>
          <InputText v-model="form.email" type="email" required placeholder="budi@kantor.id" class="w-full" />
        </Field>
        <Field label="Password" required hint="Minimal 8 karakter. Akun dengan email sama akan diperbarui.">
          <Password v-model="form.password" :feedback="false" toggleMask class="w-full" inputClass="w-full" />
        </Field>
        <Field label="Role">
          <Select v-model="form.role" :options="roleOptions" optionLabel="label" optionValue="value" class="w-full" />
        </Field>
      </form>
      <template #footer>
        <Button label="Batal" text severity="secondary" @click="showForm = false" />
        <Button label="Simpan" icon="pi pi-save" :loading="saving" @click="save" />
      </template>
    </Dialog>
  </div>
</template>
