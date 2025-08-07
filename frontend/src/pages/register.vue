<script setup lang="ts">
import { useUserStore } from '@/stores/userStore';
import { ref } from 'vue';

const userStore = useUserStore();

const processing = ref<boolean>(false);
const hasError = ref<boolean>(false);
const errorText = ref<string>("");

const username = ref<string>("");
const email = ref<string>("");
const password = ref<string>("");
const passwordConfirm = ref<string>("");

const handleRegistration = async () => {
    if (passwordConfirm.value !== password.value) {
        hasError.value = true;
        errorText.value = "Password confirmation does not match password!"
    }

    processing.value = true;
    try {
        await userStore.registerUser(email.value, username.value, password.value);
    } catch(error) {
        hasError.value = true;
        errorText.value = error;
    } finally {
        email.value = "";
        username.value = "";
        password.value = "";
        passwordConfirm.value = "";
        processing.value = false;
    }
}

const handlePasswordConfirmChange = (e) => {
    if (e.target.value !== password.value) {
        hasError.value = true;
        errorText.value = "Password confirmation does not match password!";
    } else {
        hasError.value = false;
        errorText.value = "";
    }
}
</script>

<template>
    <p>Registration</p>
    <form 
    class="flex flex-col gap-2 w-sm max-w-md"
    @submit.prevent="() => handleRegistration()">
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
            <label for="username">
                Username
            </label>
            <input
                class="border p-2 rounded-md"
                minlength="4"
                pattern="[a-zA-Z0-9]{4,30}"
                required
                type="text" id="username" v-model="username" />
        </fieldset>
        <fieldset class="flex flex-col gap-2">
            <label>Password</label>
            <input
                class="border p-2 rounded-md"
                minlength="8"
                maxlength="72"
                pattern="[a-zA-Z0-9!@#$%^*_]{8,72}"
                required
                type="password" id="password" v-model="password" />
        </fieldset>
        <fieldset class="flex flex-col gap-2">
            <label>Confirm Password</label>
            <input
                @input="(e) => handlePasswordConfirmChange(e)"
                class="border p-2 rounded-md"
                minlength="8"
                maxlength="72"
                required
                pattern="[a-zA-Z0-9!@#$%^*_]{8,72}"
                type="password" id="confirm-password" v-model="passwordConfirm" />
        </fieldset>
        <button type="submit">Register</button>
    </form>
</template>
