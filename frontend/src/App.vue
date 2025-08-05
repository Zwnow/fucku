<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useUserStore } from '@/stores/userStore';
import { isAuthenticated } from '@/utils/utils';

const userStore = useUserStore();
const loading = ref(true);

onMounted(async () => {
  loading.value = true;

  try {
    userStore.loggedIn = await isAuthenticated();
  } catch (error) {
    userStore.loggedIn = false;
  } finally {
    loading.value = false;
  }
});

const handleLogout = async () => {
  try {
    await userStore.logoutUser();
  } catch(error) {
    console.error(error);
  } finally {

  }
}
</script>

<template>
  <nav v-if="!userStore.loggedIn">
    <RouterLink to="/">Home</RouterLink>
    <RouterLink to="/login">Login</RouterLink>
  </nav>
  <nav v-else>
    <RouterLink to="/profile">Profile</RouterLink>
    <form @submit.prevent="() => handleLogout()">
      <button 
      class="cursor-pointer"
      type="submit">Logout</button>
    </form>
  </nav>
  <main>
    <RouterView></RouterView>
  </main>
</template>

<style scoped></style>
