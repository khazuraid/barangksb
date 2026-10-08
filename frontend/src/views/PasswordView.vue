<script setup lang="ts">
import { ref } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Password from 'primevue/password'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import Field from '@/components/Field.vue'

const toast = useToast()
const router = useRouter()
const oldPw = ref(''), newPw = ref(''), confirmPw = ref('')
const loading = ref(false)

const rules = [
  { t: 'Minimal 8 karakter', ok: () => newPw.value.length >= 8 },
  { t: 'Mengandung huruf', ok: () => /[a-zA-Z]/.test(newPw.value) },
  { t: 'Mengandung angka', ok: () => /\d/.test(newPw.value) },
  { t: 'Konfirmasi cocok', ok: () => newPw.value.length > 0 && newPw.value === confirmPw.value },
]

async function submit() {
  if (newPw.value !== confirmPw.value) {
    toast.add({ severity: 'warn', summary: 'Konfirmasi tidak cocok', life: 2500 })
    return
  }
  if (newPw.value.length < 8) {
    toast.add({ severity: 'warn', summary: 'Password terlalu pendek', detail: 'Minimal 8 karakter', life: 2500 })
    return
  }
  loading.value = true
  try {
    await api.post('/auth/password', { old: oldPw.value, new: newPw.value })
    toast.add({ severity: 'success', summary: 'Password diganti', detail: 'Gunakan password baru saat login berikutnya', life: 3500 })
    oldPw.value = newPw.value = confirmPw.value = ''
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal mengganti password', detail: e.response?.data?.error, life: 4000 })
  } finally { loading.value = false }
}
</script>

<template>
  <div>
    <PageHeader crumb="Akun" title="Ganti Password" sub="Perbarui kredensial akun Anda" />

    <div class="grid lg:grid-cols-[1fr_300px] gap-4 items-start max-w-4xl">
      <Panel title="Kredensial" icon="pi pi-key">
        <form @submit.prevent="submit" class="flex flex-col gap-4">
          <Field label="Password Lama" required>
            <Password v-model="oldPw" :feedback="false" toggleMask class="w-full" inputClass="w-full" />
          </Field>
          <Field label="Password Baru" required>
            <Password v-model="newPw" :feedback="false" toggleMask class="w-full" inputClass="w-full" />
          </Field>
          <Field label="Ulangi Password Baru" required>
            <Password v-model="confirmPw" :feedback="false" toggleMask class="w-full" inputClass="w-full" />
          </Field>

          <div class="flex gap-2.5 pt-1">
            <Button type="submit" label="Simpan Password" icon="pi pi-save" :loading="loading"
                    :disabled="!oldPw || !newPw || !confirmPw" />
            <Button label="Kembali" icon="pi pi-arrow-left" text severity="secondary" @click="router.push('/')" />
          </div>
        </form>
      </Panel>

      <Panel title="Syarat Kekuatan" icon="pi pi-shield">
        <ul class="flex flex-col gap-2.5 text-[12px]">
          <li v-for="r in rules" :key="r.t" class="flex items-center gap-2.5">
            <i class="pi text-[11px]" :class="r.ok() ? 'pi-check-circle text-sig-ok' : 'pi-circle text-ink-300'" />
            <span :style="r.ok() ? '' : 'color: var(--txt-dim)'">{{ r.t }}</span>
          </li>
        </ul>
        <div class="mt-4 pt-3.5 border-t text-[11.5px] leading-relaxed"
             style="border-color: var(--line); color: var(--txt-dim)">
          Setelah password diganti, aktivitas login Anda tetap tercatat pada audit log sistem.
        </div>
      </Panel>
    </div>
  </div>
</template>