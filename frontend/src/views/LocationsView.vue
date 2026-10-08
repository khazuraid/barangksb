<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import Card from 'primevue/card'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
const confirm = useConfirm(); const toast = useToast()
const locs = ref([]); const newName = ref('')
onMounted(async () => { const res = await api.get('/locations'); locs.value = res.data })
async function add() { if (!newName.value) return; await api.post('/locations', { name: newName.value }); newName.value=''; const res = await api.get('/locations'); locs.value = res.data; toast.add({severity:'success',summary:'Ditambah',life:2000}) }
function del(id:string, name:string) { confirm.require({ message:`Hapus "${name}"?`, header:'Konfirmasi', accept: async () => { await api.delete(`/locations/${id}`); const res = await api.get('/locations'); locs.value = res.data } }) }
</script>
<template>
  <div class="flex flex-col gap-4">
    <h2 class="text-xl font-bold">Lokasi ({{ locs.length }})</h2>
    <Card class="shadow-sm"><template #content>
      <div class="flex gap-2"><InputText v-model="newName" placeholder="Nama ruangan" class="flex-1" @keyup.enter="add" /><Button label="Tambah" icon="pi pi-plus" @click="add" /></div>
    </template></Card>
    <Card class="shadow-sm"><template #content>
      <DataTable :value="locs" class="p-datatable-sm" stripedRows>
        <Column field="name" header="Nama"><template #body="{data}"><strong>{{data.name}}</strong></template></Column>
        <Column header="" style="width:80px"><template #body="{data}"><Button icon="pi pi-trash" text rounded size="small" severity="danger" @click="del(data.id, data.name)" /></template></Column>
      </DataTable>
    </template></Card>
  </div>
</template>
