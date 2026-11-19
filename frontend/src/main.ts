import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import Aura from '@primeuix/themes/aura'
import { definePreset } from '@primeuix/themes'
import Tooltip from 'primevue/tooltip'
import ConfirmationService from 'primevue/confirmationservice'
import ToastService from 'primevue/toastservice'
import router from './router'
import { useThemeStore } from './stores/theme'
import './assets/main.css'
import 'primeicons/primeicons.css'
import App from './App.vue'

/*
 * "Clinical Precision" — unified Cool-Slate surfaces + single Clinical Teal accent.
 * Subdued saturation for long operator sessions, cohesive across light & dark themes.
 */
const ClinicalPrecision = definePreset(Aura, {
  semantic: {
    primary: {
      50:  '#f0fdfa', 100: '#ccfbf1', 200: '#99f6e4', 300: '#5eead4',
      400: '#2dd4bf', 500: '#14b8a6', 600: '#0d9488', 700: '#0f766e',
      800: '#115e59', 900: '#134e4a', 950: '#042f2e',
    },
    formField: {
      paddingX: '0.625rem',
      paddingY: '0.4375rem',
      sm: {
        fontSize: '0.75rem',
        paddingX: '0.5rem',
        paddingY: '0.3125rem',
      },
      lg: {
        fontSize: '0.9375rem',
        paddingX: '0.875rem',
        paddingY: '0.5625rem',
      },
      borderRadius: '{border.radius.md}',
      transitionDuration: '0.14s',
    },
    colorScheme: {
      light: {
        surface: {
          0:   '#ffffff',
          50:  '#f8fafc',
          100: '#f1f5f9',
          200: '#e2e8f0',
          300: '#cbd5e1',
          400: '#94a3b8',
          500: '#64748b',
          600: '#475569',
          700: '#334155',
          800: '#1e293b',
          900: '#0f172a',
          950: '#020617',
        },
        primary: {
          color: '{primary.600}',
          contrastColor: '#ffffff',
          hoverColor: '{primary.700}',
          activeColor: '{primary.800}',
        },
        formField: {
          background: '#ffffff',
          borderColor: '#d5dde7',
          hoverBorderColor: '#94a3b8',
          focusBorderColor: '{primary.600}',
          color: '#0f172a',
          placeholderColor: '#64748b',
          iconColor: '#64748b',
        },
        content: {
          background: '#ffffff',
          hoverBackground: '#f4f7fa',
          borderColor: '#dfe6ee',
          color: '#0f172a',
        },
        overlay: {
          select:  { background: '#ffffff', borderColor: '#d5dde7', color: '#0f172a' },
          popover: { background: '#ffffff', borderColor: '#d5dde7', color: '#0f172a' },
          modal:   { background: '#ffffff', borderColor: '#d5dde7', color: '#0f172a' },
        },
      },
      dark: {
        surface: {
          0:   '#ffffff',
          50:  '#f8fafc',
          100: '#f1f5f9',
          200: '#e2e8f0',
          300: '#cbd5e1',
          400: '#94a3b8',
          500: '#64748b',
          600: '#475569',
          700: '#263449',
          800: '#192334',
          900: '#131b27',
          950: '#0c1118',
        },
        primary: {
          color: '{primary.400}',
          contrastColor: '#042f2e',
          hoverColor: '{primary.300}',
          activeColor: '{primary.500}',
        },
        formField: {
          background: '{surface.800}',
          borderColor: '{surface.700}',
          hoverBorderColor: '#3a4c66',
          focusBorderColor: '{primary.400}',
          color: '{surface.100}',
          placeholderColor: '{surface.500}',
          floatLabelColor: '{surface.400}',
          floatLabelFocusColor: '{primary.400}',
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
const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(PrimeVue, {
  theme: {
    preset: ClinicalPrecision,
    options: {
      darkModeSelector: 'body.dark',
      cssLayer: false,
    },
  },
  ripple: false,
})
app.use(ConfirmationService)
app.use(ToastService)
app.directive('tooltip', Tooltip)

// Sync theme immediately before mount to prevent light/dark flash
useThemeStore(pinia).apply()

app.mount('#app')
