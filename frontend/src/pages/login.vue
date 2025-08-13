<script setup lang="ts">
import { useUserStore } from '@/stores/userStore';
import { ref } from 'vue';

const userStore = useUserStore();

const processing = ref<boolean>(false);
const hasError = ref<boolean>(false);
const errorText = ref<string>("");

const email = ref<string>("");
const password = ref<string>("");

const handleLogin = async () => {
    processing.value = true;
    const success = await userStore.loginUser(email.value, password.value);
    if (!success) {
        hasError.value = true;
        errorText.value = "Something went wrong. Please check your credentials!";
    } else {
        email.value = "";
        password.value = "";
    }
    processing.value = false;
}
</script>

<template>
    <p>Login</p>
    <form 
    class="flex flex-col gap-2 w-sm max-w-md"
    @submit.prevent="() => handleLogin()">
        <p v-if="hasError" class="text-sm text-red-700">{{ errorText }}</p>
        <fieldset class="flex flex-col gap-2">
            <label for="email">
                Email
            </label>
            <input
                class="border p-2 rounded-md"
                required
                type="email" id="email" v-model="email" />
        </fieldset>
        <fieldset class="flex flex-col gap-2">
            <label>Password</label>
            <input
                class="border p-2 rounded-md"
                minlength="8"
                maxlength="72"
                required
                type="password" id="password" v-model="password" />
        </fieldset>
        <button type="submit">Login</button>
    </form>
</template>