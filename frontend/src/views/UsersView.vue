<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import Card from 'primevue/card'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Select from 'primevue/select'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
const toast = useToast()
const users = ref([]); const form = ref({ name:'', email:'', password:'', role:'petugas' })
const roleOptions = [{label:'Petugas',value:'petugas'},{label:'Admin',value:'admin'}]
async function fetch() { users.value = (await api.get('/users')).data }
async function create() { try { await api.post('/users', form.value); form.value={name:'',email:'',password:'',role:'petugas'}; fetch(); toast.add({severity:'success',summary:'Dibuat',life:2000}) } catch(e:any){toast.add({severity:'error',summary:'Gagal',detail:e.response?.data?.error,life:3000})} }
onMounted(fetch)
</script>
<template>
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
    <Card class="shadow-sm"><template #title><i class="pi pi-user-plus text-emerald-500 mr-2"></i>Tambah Pengguna</template><template #content>
      <form @submit.prevent="create" class="flex flex-col gap-3">
        <div><label class="block text-sm font-medium mb-1">Nama</label><InputText v-model="form.name" class="w-full" /></div>
        <div><label class="block text-sm font-medium mb-1">Email</label><InputText v-model="form.email" class="w-full" /></div>
        <div><label class="block text-sm font-medium mb-1">Password</label><Password v-model="form.password" :feedback="false" toggleMask class="w-full" inputClass="w-full" /></div>
        <div><label class="block text-sm font-medium mb-1">Role</label><Select v-model="form.role" :options="roleOptions" optionLabel="label" optionValue="value" class="w-full" /></div>
        <Button type="submit" label="Simpan" icon="pi pi-save" />
      </form>
    </template></Card>
    <Card class="shadow-sm"><template #title><i class="pi pi-list text-emerald-500 mr-2"></i>Daftar</template><template #content>
      <DataTable :value="users" class="p-datatable-sm" stripedRows>
        <Column field="name" header="Nama"><template #body="{data}"><strong>{{data.name}}</strong></template></Column>
        <Column field="email" header="Email"><template #body="{data}"><span class="text-gray-500 text-sm">{{data.email}}</span></template></Column>
        <Column field="role" header="Role" style="width:100px"><template #body="{data}"><Tag :severity="data.role==='admin'?'info':'success'" :value="data.role" /></template></Column>
      </DataTable>
    </template></Card>
  </div>
</template>
