<script setup lang="ts">
import { ref } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Password from 'primevue/password'
const toast = useToast()
const oldPw = ref(''); const newPw = ref(''); const confirm = ref('')
async function submit() {
  if (newPw.value !== confirm.value) { toast.add({severity:'warn',summary:'Konfirmasi tidak cocok',life:2000}); return }
  try { await api.post('/auth/password', { old: oldPw.value, new: newPw.value }); toast.add({severity:'success',summary:'Password diganti',life:2000}) } catch(e:any){toast.add({severity:'error',summary:'Gagal',detail:e.response?.data?.error,life:3000})}
}
</script>
<template>
  <Card class="shadow-sm max-w-md"><template #title><i class="pi pi-key text-emerald-500 mr-2"></i>Ganti Password</template><template #content>
    <form @submit.prevent="submit" class="flex flex-col gap-3">
      <div><label class="block text-sm font-medium mb-1">Password Lama</label><Password v-model="oldPw" :feedback="false" toggleMask class="w-full" inputClass="w-full" /></div>
      <div><label class="block text-sm font-medium mb-1">Password Baru</label><Password v-model="newPw" :feedback="false" toggleMask class="w-full" inputClass="w-full" /></div>
      <div><label class="block text-sm font-medium mb-1">Ulangi</label><Password v-model="confirm" :feedback="false" toggleMask class="w-full" inputClass="w-full" /></div>
      <Button type="submit" label="Simpan" icon="pi pi-save" />
    </form>
  </template></Card>
</template>
