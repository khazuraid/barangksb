<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import { useToast } from 'primevue/usetoast'
import AppLogo from '@/components/AppLogo.vue'

const auth = useAuthStore()
const router = useRouter()
const toast = useToast()
const email = ref('')
const password = ref('')
const loading = ref(false)
const showHint = ref(false)

async function login() {
  loading.value = true
  try {
    await auth.login(email.value, password.value)
    router.push('/')
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Login gagal', detail: e.response?.data?.error || 'Periksa email dan password', life: 3500 })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen shell grid lg:grid-cols-[1.1fr_1fr]">
    <!-- ============ LEFT: identity panel ============ -->
    <section class="relative hidden lg:flex flex-col justify-between p-12 overflow-hidden">
      <div class="grid-bg absolute inset-0 opacity-[0.35]" style="--line-soft: #1b2130" />

      <div class="relative flex items-center">
        <AppLogo size="lg" />
      </div>

      <div class="relative max-w-lg">
        <div class="t-label mb-3">Kendali Inventaris</div>
        <h1 class="text-[42px] leading-[1.06] font-extrabold tracking-[-0.03em]">
          Setiap unit<br />
          <span class="text-acc-500">tercatat</span>, setiap<br />
          gerakan terlacak.
        </h1>
        <p class="mt-5 text-[13.5px] leading-relaxed text-ink-400 max-w-md">
          Barcode, mutasi stok, opname massal, dan jejak audit dalam satu panel.
          Data tersimpan di basis data kantor Anda sendiri.
        </p>

        <div class="grid grid-cols-3 gap-px mt-9 rounded-lg overflow-hidden border" style="border-color: var(--line); background: var(--line)">
          <div v-for="f in [
            { i: 'pi pi-qrcode', t: 'QR Label' },
            { i: 'pi pi-chart-bar', t: 'Stok Live' },
            { i: 'pi pi-shield', t: 'Audit Trail' },
          ]" :key="f.t" class="px-4 py-4" style="background: var(--panel)">
            <i :class="f.i" class="text-acc-500 text-sm" />
            <div class="text-[11.5px] font-semibold mt-2">{{ f.t }}</div>
          </div>
        </div>
      </div>

      <div class="relative t-label">
        © 2026 · Inventaris Kantor · Internal
      </div>
    </section>

    <!-- ============ RIGHT: form ============ -->
    <section class="flex items-center justify-center p-6 lg:p-12" style="background: var(--panel)">
      <div class="w-full max-w-[360px]">
        <div class="lg:hidden flex items-center justify-center mb-8">
          <AppLogo variant="stacked" size="lg" />
        </div>

        <div class="t-label mb-1.5">Autentikasi</div>
        <h2 class="text-[24px] font-bold tracking-tight">Masuk ke panel</h2>
        <p class="text-[12.5px] mt-1.5 text-ink-400">Gunakan akun kantor yang terdaftar.</p>

        <form @submit.prevent="login" class="flex flex-col gap-4 mt-8">
          <label class="flex flex-col gap-1.5">
            <span class="text-[11.5px] font-semibold text-ink-300">Email</span>
            <InputText v-model="email" type="email" required placeholder="nama@kantor.id" class="w-full" />
          </label>

          <label class="flex flex-col gap-1.5">
            <span class="text-[11.5px] font-semibold text-ink-300">Password</span>
            <Password v-model="password" required :feedback="false" toggleMask
                      placeholder="••••••••" class="w-full" inputClass="w-full" />
          </label>

          <Button type="submit" label="Masuk" icon="pi pi-arrow-right" iconPos="right"
                  severity="warn" :loading="loading" class="w-full mt-1" />

          <Button type="button" label="Akun default sistem" icon="pi pi-info-circle"
                  text severity="secondary" class="text-left"
                  @click="showHint = !showHint" />
          <div v-if="showHint" class="panel shell-panel-2 px-3 py-2.5 t-mono text-ink-300">
            email: admin@kantor.id<br />
            password: lihat ADMIN_PASSWORD di .env
          </div>
        </form>

        <p class="text-[11px] mt-8 text-ink-600 leading-relaxed">
          Dengan masuk, aktivitas Anda tercatat pada audit log sistem.
        </p>
      </div>
    </section>
  </div>
</template>
