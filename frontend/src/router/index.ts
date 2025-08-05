import { createRouter, createWebHistory } from 'vue-router';
import HomePage from '@/pages/home.vue';
import LoginPage from '@/pages/login.vue';
import { isAuthenticated } from '@/utils/utils';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "Home",
      component: HomePage,
      meta: {
        requiresAuth: false,
      }
    },
    {
      path: "/login",
      name: "Login",
      component: LoginPage,
      meta: {
        requiresAuth: false,
      }
    },
    {
      path: "/profile",
      name: "Profile",
      component: HomePage,
      meta: {
        requiresAuth: true,
      }
    },
  ],
});

router.beforeEach(async (to, from, next) => {
  if (to.meta.requiresAuth) {
    const isLoggedIn = await isAuthenticated();
    if (!isLoggedIn) {
      return next({ path: "/login" });
    }
  }
  next();
});

export default router
