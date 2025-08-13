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
    id: number|null;
    genre_name: string;
    created_at: string|null;
    updated_at: string|null;
}

export type SpecialTags = {
    id: number;
    name: string;
    description: string | null;
    created_at: string|null;
    updated_at: string|null;
}

export const useSongStore = defineStore("songs", () => {
    const songs = ref<Song[]>([]);
    const genres = ref<Genre[]>([]);

    const configStore = useConfigStore();

    const createSong = async (song: Song) => {
        try {
            const response = await POST_REQUEST(song, "/songs");
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

    const createGenre = async (genre: Genre) => {
        try {
            const response = await POST_REQUEST(genre, "/genres");
            console.log(response);
        } catch(err) {
            console.error(err);
        }
    }

    const getGenres = async () => {
        try {
            const response = await fetch(`${configStore.baseUrl}/genres`, {
                //credentials: "include",
                method: "GET",
            });

            if (response.status === 200) {
                genres.value = await response.json();
            }
        } catch(err) {
            console.error(err);
        }
    }

    const POST_REQUEST = async (item: Song|Genre, path: string) => {
        return await fetch(`${configStore.baseUrl}${path}`, {
            credentials: "include",
            method: "POST",
            headers: {
                "X-CSRF-Token": getCSRFToken(),
                "Content-Type": "application/json"
            },
            body: JSON.stringify(item)
        })
    }

    return {
        songs,
        genres,
        createSong,
        getSongs,
        createGenre,
        getGenres,
    }
});