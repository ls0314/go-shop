import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue')
  },
  {
    path: '/about',
    name: 'About',
    component: () => import('../views/About.vue')
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
    meta: { hideNav: true }
  },
  {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { hideNav: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router


