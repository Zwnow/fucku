<script setup lang="ts">
import { useUserStore } from "@/stores/userStore.ts";
import { useCharacterStore } from "@/stores/characterStore.ts";
import { onMounted, ref } from "vue";

const userStore = useUserStore();
const characterStore = useCharacterStore();

const processing = ref<boolean>(true);

onMounted(async () => {
    processing.value = true;

    const success = await characterStore.fetchCharacter()
    console.log(characterStore.character);

    processing.value = false;
})

const handleCreateCharacter = async () => {
    if (processing.value === true) {
        return
    }

    processing.value = true;

    const success = await characterStore.createCharacter(name.value);

    processing.value = true;
}

const name = ref<string>("");
</script>

<template>
    <form v-if="!processing && characterStore.character === null"
    class="flex flex-col gap-2 w-sm max-w-md"
    @submit.prevent="() => handleCreateCharacter()">
        <p>Create Character</p>
        <input class="border p-2 rounded-md"
            v-model="name"
            type="text"
            pattern="[a-zA-Z0-9]{4,30}"
            required
        />
        <button type="submit">Create</button>
    </form>

    {{ userStore.user.username }}
</template>
