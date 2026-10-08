<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import api from '@/api'
import Card from 'primevue/card'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Paginator from 'primevue/paginator'

const txs = ref([]); const total = ref(0); const page = ref(0); const perPage = ref(25)
const sku = ref(''); const type = ref('')
const typeOptions = [{label:'IN',value:'IN'},{label:'OUT',value:'OUT'},{label:'ADJUST+',value:'ADJUST+'},{label:'ADJUST-',value:'ADJUST-'}]

async function fetch() {
  const res = await api.get('/transactions', { params: { sku: sku.value, type: type.value, page: page.value+1, per_page: perPage.value } })
  txs.value = res.data.data; total.value = res.data.total
}
onMounted(fetch); watch([page, perPage], fetch)
</script>
<template>
  <div class="flex flex-col gap-4">
    <div class="flex justify-between items-center"><h2 class="text-xl font-bold">Riwayat Mutasi ({{ total }})</h2></div>
    <Card class="shadow-sm"><template #content>
      <div class="flex gap-3 flex-wrap items-end">
        <div><label class="block text-xs font-medium mb-1">SKU</label><InputText v-model="sku" @keyup.enter="fetch" class="w-full" /></div>
        <div><label class="block text-xs font-medium mb-1">Jenis</label><Select v-model="type" :options="typeOptions" optionLabel="label" optionValue="value" showClear placeholder="Semua" class="w-full" @change="fetch" /></div>
        <Button label="Filter" icon="pi pi-filter" @click="fetch" />
      </div>
    </template></Card>
    <Card class="shadow-sm"><template #content>
      <DataTable :value="txs" responsiveLayout="scroll" class="p-datatable-sm" stripedRows>
        <Column field="timestamp" header="Waktu" style="width:180px"><template #body="{data}"><span class="text-sm text-gray-500">{{ data.timestamp }}</span></template></Column>
        <Column field="type" header="Jenis" style="width:100px"><template #body="{data}"><Tag :severity="data.type==='IN'?'success':data.type==='OUT'?'danger':'warn'" :value="data.type" /></template></Column>
        <Column field="item_name" header="Barang"><template #body="{data}"><strong>{{data.item_name}}</strong> <code class="text-xs">{{data.item_sku}}</code></template></Column>
        <Column field="quantity" header="Qty" style="width:100px"><template #body="{data}">{{data.quantity}} {{data.unit}}</template></Column>
        <Column field="received_by" header="Person" style="width:120px"><template #body="{data}"><span class="text-gray-500 text-sm">{{data.received_by}}</span></template></Column>
      </DataTable>
      <Paginator :rows="perPage" :totalRecords="total" v-model:first="page" :rowsPerPageOptions="[20,25,50,100]" @page="(e:any)=>{perPage=e.rows;page=e.page}" />
    </template></Card>
  </div>
</template>
