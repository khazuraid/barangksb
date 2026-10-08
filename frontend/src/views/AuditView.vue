<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import Card from 'primevue/card'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
const logs = ref([])
onMounted(async () => { logs.value = (await api.get('/audit')).data })
</script>
<template>
  <Card class="shadow-sm"><template #title><i class="pi pi-shield text-emerald-500 mr-2"></i>Audit Log</template><template #content>
    <DataTable :value="logs" class="p-datatable-sm" stripedRows responsiveLayout="scroll">
      <Column field="at" header="Waktu" style="width:180px"><template #body="{data}"><span class="text-sm text-gray-500">{{data.at}}</span></template></Column>
      <Column field="table_name" header="Tabel" style="width:120px"><template #body="{data}"><strong>{{data.table_name}}</strong></template></Column>
      <Column field="op" header="Aksi" style="width:100px"><template #body="{data}"><Tag :severity="data.op==='INSERT'?'success':data.op==='DELETE'?'danger':'warn'" :value="data.op" /></template></Column>
      <Column field="row_id" header="Row ID"><template #body="{data}"><code class="text-xs">{{data.row_id}}</code></template></Column>
    </DataTable>
  </template></Card>
</template>
