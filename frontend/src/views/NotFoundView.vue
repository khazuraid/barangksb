<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import AppLogo from '@/components/AppLogo.vue'

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
  <div class="min-h-dvh shell grid place-items-center p-5">
    <div class="w-full max-w-[440px] flex flex-col items-center">
      <div class="mb-7">
        <AppLogo size="md" />
      </div>

      <div class="panel w-full p-7 text-center flex flex-col items-center">
        <div class="t-mono !text-[56px] leading-none font-semibold tracking-[-0.04em] text-acc-500 select-none">
          404
        </div>

        <h1 class="text-[19px] font-bold mt-3 tracking-[-0.02em] text-txt">Halaman tidak ditemukan</h1>
        <p class="text-[13px] mt-1.5 text-txt-dim leading-relaxed max-w-xs">
          Tautan yang Anda tuju mungkin sudah dipindahkan, dihapus, atau alamat URL tidak sesuai.
        </p>

        <div class="mt-4 px-3 py-1.5 rounded-[7px] border border-line bg-panel-2 t-mono !text-[11px] text-txt-dim max-w-full truncate">
          {{ route.fullPath }}
        </div>

        <div class="flex flex-col sm:flex-row gap-2 w-full mt-6 pt-5 border-t border-line">
          <Button
            label="Dashboard"
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
          type="button"
          class="text-[12px] text-txt-dim hover:text-txt mt-4 flex items-center gap-1.5 transition-colors cursor-pointer"
          @click="goBack"
        >
          <i class="pi pi-arrow-left text-[10.5px]" aria-hidden="true" />
          <span>Kembali ke halaman sebelumnya</span>
        </button>
      </div>
    </div>
  </div>
</template>
