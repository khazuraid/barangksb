<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import { useToast } from 'primevue/usetoast'
import AppLogo from '@/components/AppLogo.vue'

const auth = useAuthStore()
const theme = useThemeStore()
const router = useRouter()
const route = useRoute()
const toast = useToast()

const email = ref('')
const password = ref('')
const loading = ref(false)
const errorMsg = ref('')
const showHint = ref(false)

function safeRedirectPath(raw: unknown): string {
  if (typeof raw !== 'string') return '/'
  const s = raw.trim()
  if (!s.startsWith('/') || s.startsWith('//') || s.startsWith('/\\') || s.includes('\r') || s.includes('\n')) {
    return '/'
  }
  return s
}

async function login() {
  errorMsg.value = ''
  loading.value = true
  try {
    await auth.login(email.value.trim(), password.value)
    router.push(safeRedirectPath(route.query.redirect))
  } catch (e: any) {
    const detail = e.response?.data?.error || 'Periksa kembali email dan password Anda.'
    errorMsg.value = detail
    toast.add({ severity: 'error', summary: 'Login gagal', detail, life: 3500 })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-dvh shell grid lg:grid-cols-[1.15fr_1fr]">
    <!-- ============ LEFT: Architectural Identity Panel ============ -->
    <section class="relative hidden lg:flex flex-col justify-between p-12 overflow-hidden border-r border-line bg-[var(--sidebar-bg)]">
      <div class="grid-bg absolute inset-0 opacity-60 pointer-events-none" aria-hidden="true" />
      <div
        class="absolute -bottom-32 -left-24 w-[460px] h-[460px] rounded-full pointer-events-none"
        style="background: radial-gradient(circle, var(--acc-wash) 0%, transparent 70%);"
        aria-hidden="true"
      />

      <div class="relative flex items-center justify-between">
        <AppLogo size="lg" />
        <span class="t-mono !text-[10.5px] px-2.5 py-1 rounded-[6px] border border-line bg-panel text-txt-dim">
          BMD &amp; Aset Operasional
        </span>
      </div>

      <div class="relative max-w-lg my-auto">
        <div class="t-mono !text-[11px] uppercase tracking-[0.08em] text-acc-500 mb-3 flex items-center gap-2">
          <span class="w-2 h-2 rounded-full bg-acc-500" aria-hidden="true" />
          <span>Kendali Inventaris</span>
        </div>
        <h1 class="text-[40px] leading-[1.08] font-extrabold tracking-[-0.035em] text-txt">
          Setiap unit<br />
          <span class="text-acc-500">tercatat</span>, setiap<br />
          mutasi terlacak.
        </h1>
        <p class="mt-4 text-[14px] leading-relaxed text-txt-dim max-w-[46ch]">
          Label QR, pencatatan keluar-masuk stok, opname massal, dan jejak audit dalam satu ruang kerja terpusat.
        </p>

        <div class="grid grid-cols-3 gap-2.5 mt-9">
          <div
            v-for="f in [
              { i: 'pi pi-qrcode', t: 'Label QR', d: 'Cetak & pindai cepat' },
              { i: 'pi pi-chart-bar', t: 'Stok Live', d: 'Pantau batas minimum' },
              { i: 'pi pi-shield', t: 'Audit Trail', d: 'Log perubahan otomatis' },
            ]"
            :key="f.t"
            class="panel p-3.5"
          >
            <i :class="f.i" class="text-acc-500 text-[14px]" aria-hidden="true" />
            <div class="text-[12.5px] font-semibold text-txt mt-2">{{ f.t }}</div>
            <div class="text-[11px] text-txt-dim mt-0.5">{{ f.d }}</div>
          </div>
        </div>
      </div>

      <div class="relative flex items-center justify-between text-[11.5px] text-txt-dim">
        <span class="t-mono !text-[11px]">© 2026 · Inventaris Puskesmas</span>
        <button
          type="button"
          class="text-acc-500 hover:underline font-medium cursor-pointer flex items-center gap-1.5"
          @click="router.push('/lapor')"
        >
          <i class="pi pi-megaphone text-[11px]" aria-hidden="true" />
          <span>Portal Pengaduan Publik</span>
        </button>
      </div>
    </section>

    <!-- ============ RIGHT: Authentication Form ============ -->
    <section class="relative flex flex-col justify-between p-6 lg:p-12 bg-panel">
      <div class="flex items-center justify-between">
        <div class="lg:hidden">
          <AppLogo size="md" />
        </div>
        <div class="ml-auto flex items-center gap-2">
          <Button
            :icon="theme.isDark ? 'pi pi-sun' : 'pi pi-moon'"
            text
            rounded
            size="small"
            severity="secondary"
            :aria-label="theme.isDark ? 'Ganti ke mode terang' : 'Ganti ke mode gelap'"
            @click="theme.toggle()"
          />
        </div>
      </div>

      <div class="w-full max-w-[360px] mx-auto my-auto py-8">
        <div class="t-mono !text-[10.5px] uppercase tracking-[0.08em] text-acc-500 mb-1.5">Autentikasi</div>
        <h2 class="text-[24px] font-bold tracking-[-0.025em] text-txt">Masuk ke panel</h2>
        <p class="text-[13px] mt-1 text-txt-dim">Gunakan akun petugas atau administrator yang terdaftar.</p>

        <div
          v-if="errorMsg"
          role="alert"
          class="mt-5 px-3.5 py-2.5 rounded-[8px] border border-[var(--sig-bad-line)] bg-[var(--sig-bad-bg)] text-sig-bad text-[12.5px] flex items-start gap-2"
        >
          <i class="pi pi-exclamation-circle mt-0.5 shrink-0 text-[13px]" aria-hidden="true" />
          <span>{{ errorMsg }}</span>
        </div>

        <form @submit.prevent="login" class="flex flex-col gap-4 mt-6">
          <label class="flex flex-col gap-1.5">
            <span class="text-[12px] font-medium text-txt-dim">Email</span>
            <InputText
              v-model="email"
              type="email"
              autocomplete="username"
              required
              placeholder="nama@puskesmas.id"
              class="w-full"
            />
          </label>

          <label class="flex flex-col gap-1.5">
            <span class="text-[12px] font-medium text-txt-dim">Password</span>
            <Password
              v-model="password"
              required
              :feedback="false"
              toggleMask
              autocomplete="current-password"
              placeholder="••••••••"
              class="w-full"
              inputClass="w-full"
            />
          </label>

          <Button
            type="submit"
            label="Masuk"
            icon="pi pi-arrow-right"
            iconPos="right"
            :loading="loading"
            class="w-full mt-1"
          />

          <button
            type="button"
            class="text-[12px] text-txt-dim hover:text-txt flex items-center gap-1.5 self-start cursor-pointer transition-colors"
            @click="showHint = !showHint"
          >
            <i class="pi pi-info-circle text-[11px]" aria-hidden="true" />
            <span>Informasi akun default sistem</span>
          </button>

          <div
            v-if="showHint"
            class="rounded-[8px] border border-line bg-panel-2 px-3 py-2.5 t-mono !text-[11.5px] text-txt-dim"
          >
            email: <span class="text-txt">admin@kantor.id</span><br />
            password: <span class="text-txt">lihat ADMIN_PASSWORD di .env</span>
          </div>
        </form>

        <div class="mt-8 pt-5 border-t border-line flex items-center justify-between text-[11.5px] text-txt-dim">
          <span>Aktivitas masuk tercatat di audit log.</span>
          <button
            type="button"
            class="lg:hidden text-acc-500 font-medium cursor-pointer"
            @click="router.push('/lapor')"
          >
            Lapor Aset
          </button>
        </div>
      </div>

      <div class="h-4" aria-hidden="true" />
    </section>
  </div>
</template>
