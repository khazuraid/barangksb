<script setup lang="ts">
import { ref, watch } from 'vue'
import api from '@/api'
import { useRouter } from 'vue-router'

const router = useRouter()
const tab = ref('in')
const items = ref([])
const form = ref({ item_id: '', quantity: 0, received_by: '', notes: '' })

watch(tab, async () => {
  if (items.value.length === 0) {
    const res = await api.get('/items', { params: { per_page: 100 } })
    items.value = res.data.data
  }
}, { immediate: true })

async function submit() {
  const endpoint = tab.value === 'out' ? '/movement/out' : tab.value === 'adjust' ? '/movement/adjust' : '/movement/in'
  await api.post(endpoint, form.value)
  form.value = { item_id: '', quantity: 0, received_by: '', notes: '' }
  router.push('/history')
}
</script>

<template>
  <div class="space-y-4">
    <div role="tablist" class="tabs tabs-boxed">
      <a role="tab" class="tab" :class="{ 'tab-active': tab === 'in' }" @click="tab = 'in'">Barang Masuk</a>
      <a role="tab" class="tab" :class="{ 'tab-active': tab === 'out' }" @click="tab = 'out'">Barang Keluar</a>
      <a role="tab" class="tab" :class="{ 'tab-active': tab === 'adjust' }" @click="tab = 'adjust'">Opname</a>
    </div>
    <div class="card bg-base-100 shadow-sm border border-base-200 max-w-xl">
      <div class="card-body">
        <form @submit.prevent="submit" class="space-y-3">
          <label class="form-control w-full">
            <span class="label-text">Barang</span>
            <select v-model="form.item_id" required class="select select-bordered select-sm">
              <option value="">— pilih —</option>
              <option v-for="i in items" :key="i.id" :value="i.id">{{ i.name }} ({{ i.sku }}) — stok {{ i.current_stock }}</option>
            </select>
          </label>
          <label class="form-control w-full">
            <span class="label-text">{{ tab === 'adjust' ? 'Stok Fisik' : 'Jumlah' }}</span>
            <input v-model.number="form.quantity" type="number" min="0" required class="input input-bordered input-sm" />
          </label>
          <label class="form-control w-full" v-if="tab !== 'adjust'">
            <span class="label-text">{{ tab === 'out' ? 'Diberikan Ke' : 'Petugas' }}</span>
            <input v-model="form.received_by" class="input input-bordered input-sm" />
          </label>
          <label class="form-control w-full">
            <span class="label-text">Keterangan</span>
            <textarea v-model="form.notes" rows="2" class="textarea textarea-bordered textarea-sm"></textarea>
          </label>
          <button type="submit" class="btn btn-primary btn-sm w-full">{{ tab === 'out' ? 'Catat Keluar' : tab === 'adjust' ? 'Simpan Opname' : 'Catat Masuk' }}</button>
        </form>
      </div>
    </div>
  </div>
</template>
