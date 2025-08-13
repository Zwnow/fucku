import { defineStore } from "pinia";
import { ref } from "vue";
import { useConfigStore } from "./configStore";
import { getCSRFToken } from "@/utils/utils";

export type Song = {
    id: string;
    song_name: string;
    album_name: string;
    artist: string;
    featuring_artist: string | null;
    spotify_embed_url: string;
    reason: string;
    genres: Genre[]|null;
    special_tags: SpecialTags[] | null;
    created_at: string|null;
    updated_at: string|null;
}

export type Genre = {
    id: number;
    genre_name: string;
    created_at: string;
    updated_at: string;
}

export type SpecialTags = {
    id: number;
    name: string;
    description: string | null;
    created_at: string;
    updated_at: string;
}

export const useSongStore = defineStore("songs", () => {
    const songs = ref<Song[]>([]);

    const configStore = useConfigStore();

    const createSong = async (song: Song) => {
        try {
            const response = await fetch(`${configStore.baseUrl}/songs`, {
                credentials: "include",
                method: "POST",
                headers: {
                    "X-CSRF-Token": getCSRFToken(),
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(song)
            });

            console.log(response);

        } catch(err) {
            console.error(err);
        }
    }

    const getSongs = async () => {
        try {
            const response = await fetch(`${configStore.baseUrl}/songs`, {
                //credentials: "include",
                method: "GET",
            });

            if (response.status === 200) {
                songs.value = await response.json();
            }
        } catch(err) {
            console.error(err);
        }
    }

    return {
        songs,
        createSong,
        getSongs,
    }
});