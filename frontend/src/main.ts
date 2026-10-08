import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import Aura from '@primeuix/themes/aura'
import { definePreset } from '@primeuix/themes'
import Tooltip from 'primevue/tooltip'
import ConfirmationService from 'primevue/confirmationservice'
import ToastService from 'primevue/toastservice'
import router from './router'
import './assets/main.css'
import 'primeicons/primeicons.css'
import App from './App.vue'

/*
 * "Industrial Precision" — dark shell, single amber accent.
 * All component sizing flows through PrimeVue's own semantic tokens:
 *   form.field.*  → input/button padding, height, border-radius
 *   button.*      → gap, iconOnlyWidth, label weight
 *   surface.*     → mapped to our ink scale (not zinc) for warm-dark cohesion
 */
const Industrial = definePreset(Aura, {
  semantic: {
    primary: {
      50:  '#fff8e6', 100: '#ffefc4', 200: '#ffe49c', 300: '#ffd166',
      400: '#faba46', 500: '#f5a524', 600: '#d98806', 700: '#a96305',
      800: '#7c4904', 900: '#5c3603', 950: '#3a2202',
    },
    /* form field tokens — shared by Button, InputText, Select, Textarea, etc.
       normal: 7px Y + 10px X + 13px font → ~36px tall
       small:  5px Y +  8px X + 12px font → ~30px tall */
    formField: {
      paddingX: '0.625rem',
      paddingY: '0.4375rem',
      sm: {
        fontSize: '0.75rem',
        paddingX: '0.5rem',
        paddingY: '0.3125rem',
      },
      lg: {
        fontSize: '1rem',
        paddingX: '0.875rem',
        paddingY: '0.5625rem',
      },
      borderRadius: '{border.radius.md}',
      transitionDuration: '0.16s',
    },
    /* button-specific tokens */
    button: {
      gap: '0.375rem',
      iconOnlyWidth: '2.25rem',
      roundedBorderRadius: '1.5rem',
      label: { fontWeight: '600' },
      sm: {
        iconOnlyWidth: '1.875rem',
      },
    },
    /* surface mapped to ink scale — warm dark, not zinc cool */
    colorScheme: {
      light: {
        surface: {
          0:   '#ffffff',
          50:  '{stone.50}',
          100: '{stone.100}',
          200: '{stone.200}',
          300: '{stone.300}',
          400: '{stone.400}',
          500: '{stone.500}',
          600: '{stone.600}',
          700: '{stone.700}',
          800: '{stone.800}',
          900: '{stone.900}',
          950: '{stone.950}',
        },
      },
      dark: {
        surface: {
          0:   '#ffffff',
          50:  '#f7f8fa',
          100: '#eceff4',
          200: '#d5dae3',
          300: '#b3bbc9',
          400: '#8791a3',
          500: '#5b6678',
          600: '#3d4759',
          700: '#2a3242',
          800: '#1d2331',
          900: '#111520',
          950: '#0b0e13',
        },
        primary: {
          color: '{primary.500}',
          contrastColor: '#0b0e13',
          hoverColor: '{primary.600}',
          activeColor: '{primary.700}',
        },
        formField: {
          background: '{surface.900}',
          borderColor: '{surface.700}',
          hoverBorderColor: '{surface.600}',
          focusBorderColor: '{primary.500}',
          color: '{surface.100}',
          placeholderColor: '{surface.400}',
          floatLabelColor: '{surface.400}',
          floatLabelFocusColor: '{primary.500}',
          iconColor: '{surface.400}',
        },
        text: {
          color: '{surface.100}',
          hoverColor: '{surface.0}',
          mutedColor: '{surface.400}',
          hoverMutedColor: '{surface.300}',
        },
        content: {
          background: '{surface.900}',
          hoverBackground: '{surface.800}',
          borderColor: '{surface.700}',
          color: '{text.color}',
        },
        overlay: {
          select:  { background: '{surface.900}', borderColor: '{surface.700}', color: '{text.color}' },
          popover: { background: '{surface.900}', borderColor: '{surface.700}', color: '{text.color}' },
          modal:   { background: '{surface.900}', borderColor: '{surface.700}', color: '{text.color}' },
        },
      },
    },
  },
})

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(PrimeVue, {
  theme: {
    preset: Industrial,
    options: {
      darkModeSelector: 'body.dark',
      cssLayer: false,
    },
  },
  ripple: true,
})
app.use(ConfirmationService)
app.use(ToastService)
app.directive('tooltip', Tooltip)
app.mount('#app')

// Automatically reload when a new deployment creates new chunk hashes
window.addEventListener('vite:preloadError', (event) => {
  event.preventDefault()
  window.location.reload()
})

window.addEventListener(
  'error',
  (event) => {
    const target = event.target as HTMLElement
    if (target && target.tagName === 'LINK' && (target as HTMLLinkElement).rel === 'stylesheet') {
      if (!sessionStorage.getItem('css_reloaded')) {
        sessionStorage.setItem('css_reloaded', 'true')
        window.location.reload()
      }
    }
  },
  true
)