import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import Aura from '@primevue/themes/aura'
import Tooltip from 'primevue/tooltip'
import ConfirmationService from 'primevue/confirmationservice'
import ToastService from 'primevue/toastservice'
import router from './router'
import './assets/main.css'
import 'primeicons/primeicons.css'
import App from './App.vue'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(PrimeVue, {
  theme: {
    preset: Aura,
    options: {
      darkModeSelector: '.dark',
      cssLayer: { name: 'primevue', order: 'theme, base, primevue' },
    },
  },
  ripple: true,
})
app.use(ConfirmationService)
app.use(ToastService)
app.directive('tooltip', Tooltip)
app.mount('#app')
