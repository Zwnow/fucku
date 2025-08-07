<script setup lang="ts">
import { useUserStore } from "@/stores/userStore.ts";

const userStore = useUserStore();
</script>

<template>
    {{ userStore.user.username }}
</template>
