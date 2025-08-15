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
    genres: Genre[] | null;
    special_tags: SpecialTag[] | null;
    created_at: string | null;
    updated_at: string | null;
}

export type Genre = {
    id: number | null;
    genre_name: string;
    created_at: string | null;
    updated_at: string | null;
}

export type SpecialTag = {
    id: number | null;
    name: string;
    description: string | null;
    created_at: string | null;
    updated_at: string | null;
}

export const useSongStore = defineStore("songs", () => {
    const songs = ref<Song[]>([]);
    const genres = ref<Genre[]>([]);
    const tags = ref<SpecialTag[]>([]);

    const configStore = useConfigStore();

    const createSong = async (song: Song): Promise<string | null> => {
        try {
            const response = await POST_REQUEST(song, "/songs");

            if (response.status === 201) {
                return await response.text();
            } else {
                return null;
            }
        } catch(err) {
            console.error(err);
            return null;
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

    const massAssignGenres = async (id: number, genres: number[]) => {
        try {
            const response = await POST_REQUEST(genres, `/genres/${id}?song=${id}`);
            console.log(response);
        } catch (err) {
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
                console.log(genres.value);
            }
        } catch(err) {
            console.error(err);
        }
    }

    const createTag = async (tag: SpecialTag) => {
        try {
            const response = await POST_REQUEST(tag, "/tags");
            console.log(response);
        } catch(err) {
            console.error(err);
        }
    }

    const getTags = async () => {
        try {
            const response = await fetch(`${configStore.baseUrl}/tags`, {
                method: "GET",
            });

            if (response.status === 200) {
                tags.value = await response.json();
            }
        } catch(err) {
            console.error(err);
        }
    }

    const POST_REQUEST = async (item: Song|Genre|SpecialTag|number[], path: string) => {
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
        createTag,
        getTags,
        massAssignGenres,
    }
});
