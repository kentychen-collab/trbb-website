import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  // Public
  { path: '/',           name: 'Home',         component: () => import('@/views/HomeView.vue') },
  { path: '/login',      name: 'Login',        component: () => import('@/views/auth/LoginView.vue') },
  { path: '/register',   name: 'Register',     component: () => import('@/views/auth/RegisterView.vue') },
  { path: '/events',     name: 'Events',       component: () => import('@/views/events/EventListView.vue') },
  { path: '/events/:id', name: 'EventDetail',  component: () => import('@/views/events/EventDetailView.vue') },
  { path: '/training',    name: 'PublicTraining', component: () => import('@/views/training/PublicTrainingView.vue') },
  { path: '/training/:id', name: 'TrainingDetail', component: () => import('@/views/training/TrainingDetailView.vue') },
  { path: '/shop',         name: 'Shop',         component: () => import('@/views/shop/ShopView.vue') },
  { path: '/shop/:id',   name: 'ProductDetail',component: () => import('@/views/shop/ProductDetailView.vue') },
  { path: '/secondhand', name: 'Secondhand',   component: () => import('@/views/secondhand/SecondhandView.vue') },
  { path: '/news',       name: 'News',         component: () => import('@/views/NewsView.vue') },

  // Protected
  {
    path: '/me',
    meta: { requiresAuth: true },
    component: () => import('@/views/me/MeLayout.vue'),
    children: [
      { path: '',          name: 'Profile',    component: () => import('@/views/me/ProfileView.vue') },
      { path: 'orders',    name: 'Orders',     component: () => import('@/views/me/OrdersView.vue') },
      { path: 'training',  name: 'Training',   component: () => import('@/views/me/TrainingView.vue') },
      { path: 'garmin',    name: 'Garmin',     component: () => import('@/views/me/GarminView.vue') },
      { path: 'my-items',  name: 'MyItems',    component: () => import('@/views/me/MyItemsView.vue') },
      { path: 'membership',name: 'Membership', component: () => import('@/views/me/MembershipView.vue') },
    ]
  },

  // Checkout
  { path: '/checkout',   name: 'Checkout',    meta: { requiresAuth: true }, component: () => import('@/views/CheckoutView.vue') },

  // Fallback
  { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('@/views/NotFoundView.vue') },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) return savedPosition
    return { top: 0 }
  }
})

router.beforeEach((to, from, next) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else {
    next()
  }
})

export default router
