import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../views/MainView.vue'),
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('../views/AboutView.vue'),
    },
    {
      path: '/services',
      name: 'services',
      component: () => import('../views/Services/index.vue'),
    },
    {
      path: '/services/:app',
      name: 'servicesApp',
      component: () => import('../views/Services/app.vue'),
    },
    {
      path: '/terminal',
      name: 'terminal',
      component: () => import('../views/TerminalView.vue'),
    },
  ],
})

export default router
