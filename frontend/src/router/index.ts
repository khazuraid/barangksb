import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue') },
    {
      path: '/',
      component: () => import('@/views/LayoutView.vue'),
      meta: { auth: true },
      children: [
        { path: '', name: 'dashboard', component: () => import('@/views/DashboardView.vue') },
        { path: 'items', name: 'items', component: () => import('@/views/ItemsView.vue') },
        { path: 'items/new', name: 'item-new', component: () => import('@/views/ItemFormView.vue') },
        { path: 'items/:id/edit', name: 'item-edit', component: () => import('@/views/ItemFormView.vue') },
        { path: 'categories', name: 'categories', component: () => import('@/views/CategoriesView.vue') },
        { path: 'locations', name: 'locations', component: () => import('@/views/LocationsView.vue') },
        { path: 'movement', name: 'movement', component: () => import('@/views/MovementView.vue') },
        { path: 'history', name: 'history', component: () => import('@/views/HistoryView.vue') },
        { path: 'barcode', name: 'barcode', component: () => import('@/views/BarcodeView.vue') },
        { path: 'users', name: 'users', component: () => import('@/views/UsersView.vue'), meta: { admin: true } },
        { path: 'audit', name: 'audit', component: () => import('@/views/AuditView.vue'), meta: { admin: true } },
        { path: 'password', name: 'password', component: () => import('@/views/PasswordView.vue') },
      ],
    },
    { path: '/scan/:id', name: 'scan', component: () => import('@/views/ScanView.vue') },
  ],
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.auth && !auth.isLoggedIn) {
    next('/login')
  } else if (to.meta.admin && !auth.isAdmin) {
    next('/')
  } else {
    next()
  }
})

export default router
