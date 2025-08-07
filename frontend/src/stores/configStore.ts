import { defineStore } from "pinia";

export const useConfigStore = defineStore("config", () => {
    const baseUrl = "http://localhost:3000"

    return {
        baseUrl,
    }
});