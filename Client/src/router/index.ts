import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '@/views/LoginView.vue'
import RegisterView from '@/views/RegisterView.vue'
import HomeView from '@/views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/im' },
    { path: '/login', component: LoginView },
    { path: '/register', component: RegisterView },
    { path: '/im', component: HomeView },
  ],
})

export default router
