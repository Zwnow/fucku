import { defineStore } from "pinia";
import { ref } from "vue";
import { useConfigStore } from "./configStore";
import { type FetchResponse } from "@/types.ts";
import router from "@/router";

export const useUserStore = defineStore("user", () => {
    const configStore = useConfigStore();

    const user = ref(null);
    const loggedIn = ref(false);

    const loginUser = async (email: string, password: string): Promise<boolean> => {
        try {
            const r = await fetch(`${configStore.baseUrl}/login`, {
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

            if (r.status === 200) {
                loggedIn.value = true;

                const data = await r.json();
                user.value = data.user;

                sessionStorage.setItem("user", JSON.stringify(user.value));

                router.push("/profile");
                return true
            }
            return false;
        } catch (err) {
            console.error(err);
            return false;
        }
    }

    const registerUser = async (email: string, username: string, password: string) => {
        try {
            const response = await fetch(`${configStore.baseUrl}/register`, {
                credentials: "include",
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    email: email,
                    username: username,
                    password: password,
                })
            });

            if (response.status === 200) {
                router.push("/login");
            } else {
                const reason = await response.text();
                switch (reason.trim()) {
                    case "contains whitespace":
                        throw new Error("Illegal characters used.");
                    case "email already taken":
                        throw new Error("Email is already taken.");
                    default:
                        throw new Error("Something went wrong.");
                }
            }
        } catch (err) {
            throw err;
        }
    }

    const logoutUser = async () => {
        try {
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

                sessionStorage.clear();

                router.push("/");
            }
        } catch(err) {
            console.error(err);
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
        registerUser,
        logoutUser,
    }
});
