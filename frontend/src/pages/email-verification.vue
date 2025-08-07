<script setup lang="ts">
import { useUserStore } from "@/stores/userStore.ts";
import { ref } from "vue";

const userStore = useUserStore();

const code = ref<string>("");

const processing = ref<boolean>(false);
const hasError = ref<boolean>(false);
const errorText = ref<string>("");

const handleVerification = async () => {
    if (processing.value === true) {
        return;
    }

    processing.value = true;

    const success = await userStore.verifyEmail(code.value);
    if (!success) {
        hasError.value = true;
        errorText.value = "Something went wrong, please try again!";
    }

    processing.value = false;
}
</script>

<template>
    <form @submit.prevent="() => handleVerification()"
        class="flex flex-col gap-2 w-sm max-w-md"
    >
        <input
            class="border p-2 rounded-md"
            required
            v-model="code"
            type="text"
        />
        <button>Confirm</button>
    </form>
</template>
