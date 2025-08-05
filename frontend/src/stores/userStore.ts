import { defineStore } from "pinia";
import { ref } from "vue";
import { useConfigStore } from "./configStore";
import router from "@/router";

export const useUserStore = defineStore("user", () => {
    const configStore = useConfigStore();

    const user = ref(null);
    const loggedIn = ref(false);

    const loginUser = async (email: string, password: string) => {
        const response = await fetch(`${configStore.baseUrl}/login`, {
            credentials: "include",
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                email: email,
                password: password,
            })
        });

        if (response.status === 200) {
            loggedIn.value = true;
            router.push("/profile");
        } else {
            throw new Error("Login failed");
        }
    }

    const logoutUser = async () => {
        const response = await fetch(`${configStore.baseUrl}/logout`, {
            credentials: "include",
            method: "POST",
            headers: {
                "X-CSRF-Token": getCSRFToken(),
                "Content-Type": "application/json",
            },
        });

        if (response.status === 200) {
            loggedIn.value = false;
            router.push("/");
        }
    }

    const getCSRFToken = (): string => {
        const name = "csrf_token=";
        const decoded = decodeURIComponent(document.cookie);
        const cookies = decoded.split(';');

        for (let cookie of cookies) {
            cookie = cookie.trim();
            if (cookie.startsWith(name)) {
                return cookie.substring(name.length);
            }
        }

        return "";
    }

    return {
        user,
        loggedIn,
        loginUser,
        logoutUser,
    }
});