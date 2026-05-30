import { createRouter, createWebHistory } from 'vue-router'
import { useAdminStore } from '@/stores/admin'

const routes = [
  { path: '/login', name: 'Login', component: () => import('@/views/LoginView.vue') },
  {
    path: '/',
    component: () => import('@/views/AdminLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '',              name: 'Dashboard',          component: () => import('@/views/dashboard/DashboardView.vue') },
      { path: 'members',        name: 'Members',            component: () => import('@/views/members/MemberListView.vue') },
      { path: 'admins',          name: 'Admins',             component: () => import('@/views/admins/AdminListView.vue') },
      { path: 'events',        name: 'Events',             component: () => import('@/views/events/EventListView.vue') },
      { path: 'events/create', name: 'EventCreate',        component: () => import('@/views/events/EventFormView.vue') },
      { path: 'events/:id',    name: 'EventDetail',        component: () => import('@/views/events/EventDetailView.vue') },
      { path: 'products',      name: 'Products',           component: () => import('@/views/products/ProductListView.vue') },
      { path: 'orders',        name: 'Orders',             component: () => import('@/views/orders/OrderListView.vue') },
      { path: 'settings',      name: 'SiteSettings',       component: () => import('@/views/settings/SiteSettingsView.vue') },
      { path: 'training',      name: 'Training',           component: () => import('@/views/training/TrainingListView.vue') },
      { path: 'announcements', name: 'Announcements',      component: () => import('@/views/announcements/AnnouncementListView.vue') },
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/' }
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach((to, from, next) => {
  const store = useAdminStore()
  if (to.meta.requiresAuth && !store.isLoggedIn) {
    next({ name: 'Login' })
  } else {
    next()
  }
})

export default router
