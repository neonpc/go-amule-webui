import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'dashboard', component: () => import('../views/Dashboard.vue') },
    { path: '/downloads', name: 'downloads', component: () => import('../views/Downloads.vue') },
    { path: '/uploads', name: 'uploads', component: () => import('../views/Uploads.vue') },
    { path: '/shared', name: 'shared', component: () => import('../views/SharedFiles.vue') },
    { path: '/search', name: 'search', component: () => import('../views/Search.vue') },
    { path: '/servers', name: 'servers', component: () => import('../views/Servers.vue') },
    { path: '/kad', name: 'kad', component: () => import('../views/Kad.vue') },
    { path: '/stats', name: 'stats', component: () => import('../views/Statistics.vue') },
    { path: '/log', name: 'log', component: () => import('../views/Log.vue') },
    { path: '/prefs', name: 'prefs', component: () => import('../views/Preferences.vue') },
  ],
})

export default router
