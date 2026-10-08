<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'

const route = useRoute()
const router = useRouter()

function goBack() {
  if (window.history.length > 1) {
    window.history.back()
  } else {
    router.push('/')
  }
}
</script>

<template>
  <div class="min-h-screen shell grid place-items-center p-5 selection:bg-acc-500 selection:text-ink-950">
    <div class="w-full max-w-[480px] text-center flex flex-col items-center">
      <!-- App Header Badge -->
      <div class="flex items-center gap-2.5 mb-8">
        <div class="w-9 h-9 grid place-items-center rounded-md bg-acc-500 text-ink-950 font-extrabold text-[13px] shadow-lg shadow-acc-500/20">
          IK
        </div>
        <div class="leading-tight text-left">
          <div class="text-[13px] font-bold tracking-wide">INVENTARIS KANTOR</div>
          <div class="t-label !text-[9.5px]">Sistem Manajemen Aset &amp; Geotag</div>
        </div>
      </div>

      <!-- Main 404 Card -->
      <div class="panel shell-panel-2 p-8 w-full rounded-xl border flex flex-col items-center shadow-2xl relative overflow-hidden"
           style="border-color: var(--line); background: var(--paper-1)">
        <!-- Subtle amber accent line top -->
        <div class="absolute top-0 inset-x-0 h-1 bg-gradient-to-r from-acc-500 via-amber-400 to-acc-600" />

        <!-- 404 Hero Number -->
        <div class="relative my-2 select-none">
          <div class="text-7xl sm:text-8xl font-black tracking-tight text-acc-500/90 font-mono">
            404
          </div>
          <div class="absolute inset-0 blur-2xl bg-acc-500/15 -z-10 rounded-full" />
        </div>

        <h1 class="text-[20px] font-bold mt-2 tracking-tight">Halaman Tidak Ditemukan</h1>
        <p class="text-[12.5px] mt-2 text-ink-300 leading-relaxed max-w-sm">
          Tautan yang Anda tuju mungkin sudah dipindahkan, dihapus, atau alamat URL salah ketik.
        </p>

        <!-- Monospace Path Pill -->
        <div class="mt-4 px-3 py-1.5 rounded-lg border text-[11px] t-mono text-ink-400 max-w-full truncate"
             style="background: var(--paper-2); border-color: var(--line)">
          <i class="pi pi-link text-[10px] mr-1 text-acc-500" />
          <span>{{ route.fullPath }}</span>
        </div>

        <!-- Navigation Buttons -->
        <div class="flex flex-col sm:flex-row gap-2.5 w-full mt-6 pt-4 border-t" style="border-color: var(--line)">
          <Button
            label="Dashboard Utama"
            icon="pi pi-th-large"
            size="small"
            class="flex-1"
            @click="router.push('/')"
          />
          <Button
            label="Daftar Barang"
            icon="pi pi-box"
            size="small"
            severity="secondary"
            outlined
            class="flex-1"
            @click="router.push('/items')"
          />
        </div>

        <button
          class="text-[11.5px] text-ink-400 hover:text-ink-200 mt-4 flex items-center gap-1.5 transition-colors cursor-pointer"
          @click="goBack"
        >
          <i class="pi pi-arrow-left text-[10px]" /> Kembali ke halaman sebelumnya
        </button>
      </div>

      <div class="mt-8 text-[11px] text-ink-500">
        © 2026 Inventaris Kantor · Hak Cipta Dilindungi
      </div>
    </div>
  </div>
</template>
