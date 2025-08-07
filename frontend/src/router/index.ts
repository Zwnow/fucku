import { createRouter, createWebHistory } from 'vue-router';
import HomePage from '@/pages/home.vue';
import LoginPage from '@/pages/login.vue';
import RegisterPage from '@/pages/register.vue';
import ProfilePage from '@/pages/profile.vue';
import VerifyEmailPage from '@/pages/email-verification.vue';
import { isAuthenticated, isVerified } from '@/utils/utils';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "Home",
      component: HomePage,
      meta: {
        requiresAuth: false,
        requiresVerified: false,
      }
    },
    {
      path: "/login",
      name: "Login",
      component: LoginPage,
      meta: {
        requiresAuth: false,
        requiresVerified: false,
      }
    },
    {
      path: "/register",
      name: "Register",
      component: RegisterPage,
      meta: {
        requiresAuth: false,
        requiresVerified: false,
      }
    },
    {
      path: "/verify-email",
      name: "Email Verification",
      component: VerifyEmailPage,
      meta: {
        requiresAuth: true,
        requiresVerified: false,
      }
    },
    {
      path: "/profile",
      name: "Profile",
      component: ProfilePage,
      meta: {
        requiresVerified: true,
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

    if (to.meta.requiresVerified) {
        const verified = await isVerified();
        if (!verified) {
            return next({ path: "/verify-email"})
        }
    }

    next();
});

export default router
