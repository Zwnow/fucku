import { defineStore } from "pinia";
import { ref } from "vue";
import { useConfigStore } from "./configStore";
import { getCSRFToken } from "@/utils/utils";

export const useCharacterStore = defineStore("character", () => {
    const configStore = useConfigStore();

    const character = ref<Object|null>(null);

    const fetchCharacter = async () => {
        try {
            const response = await fetch(`${configStore.baseUrl}/character`, {
                credentials: "include",
                method: "GET"
            });

            console.log(response);
            return true;
        } catch(err) {
            console.error(err);
            return false;
        } 
    }

    const createCharacter = async (name: string) => {
        try {
            const response = await fetch(`${configStore.baseUrl}/character`, {
                credentials: "include",
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "X-CSRF-Token": getCSRFToken(),
                },
                body: JSON.stringify({
                    name: name
                }),
            })

            console.log(response);

        } catch(err) {
            console.error(err);
            return false;
        }
    }

    return {
        character,
        fetchCharacter,
        createCharacter,
    }
});
