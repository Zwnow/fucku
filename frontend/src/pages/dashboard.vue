<script setup lang="ts">
import { useSongStore, type Song, type Genre, type SpecialTag } from '@/stores/songStore';
import { onMounted, ref } from 'vue';

const songStore = useSongStore();

onMounted(async () => {
  await songStore.getGenres();
  await songStore.getTags();
});

const song = ref<Song>({
  id: "",
  song_name: "",
  album_name: "",
  artist: "",
  featuring_artist: "",
  genres: null,
  special_tags: null,
  reason: "",
  spotify_embed_url: "",
  created_at: null,
  updated_at: null,
});

const genre = ref<Genre>({
  id: null,
  genre_name: "",
  created_at: null,
  updated_at: null,
});

const tag = ref<SpecialTag>({
  id: null,
  name: "",
  description: "",
  created_at: null,
  updated_at: null,
});

const processing = ref<boolean>(false);

const genresOpen = ref<boolean>(false);
const selectedGenres = ref<Genre[]>([]);


const handleSubmitSong = async () => {
    if (processing.value === true) {
        return;
    }

    processing.value = true;
    const songId = await songStore.createSong(song.value);
    if (songId === null) {
        return;
    }

    if (selectedGenres.value.length > 0 ) {
        const ids = selectedGenres.value.map((g) => g.id);
        await songStore.massAssignGenres(songId, ids);
    }

    processing.value = false;
}

const handleSubmitGenre = async () => {
  if (processing.value === true) {
    return;
  }

  processing.value = true;

  await songStore.createGenre(genre.value);
  await songStore.getGenres();

  processing.value = false;
}

const handleSubmitTag = async () => {
  if (processing.value === true) {
    return;
  }

  processing.value = true;

  await songStore.createTag(tag.value);
  await songStore.getTags();

  processing.value = false;
}

const selectGenre = (genre) => {
    if (selectedGenres.value.indexOf(genre) === -1) {
        selectedGenres.value.push(genre);
    }
}
</script>

<template>
  <main>
        <form class="max-w-md border rounded-md flex flex-col gap-2" @submit.prevent="() => handleSubmitSong()">
            <p class="font-bold underline">Add Song</p>
            <fieldset class="flex flex-col">
                <label for="song_name">Song Name</label>
                <input 
                    id="song_name" v-model="song.song_name"
                    required
                ></input>
            </fieldset>
            <fieldset class="flex flex-col">
                <label for="album_name">Album Name</label>
                <input 
                    id="album_name" v-model="song.album_name"
                    required></input>
            </fieldset>
            <fieldset class="flex flex-col">
                <label for="artist">Artist</label>
                <input 
                    id="artist" v-model="song.artist"
                    required
                ></input>
            </fieldset>
            <fieldset class="flex flex-col">
                <label for="reason">Reason</label>
                <textarea
                    id="reason" v-model="song.reason"
                    class="resize-none"
                    required></textarea>
            </fieldset>
            <fieldset class="flex flex-col">
                <label for="spotify_embed">Spotify Embed</label>
                <input 
                    id="spotify_embed" v-model="song.spotify_embed_url"
                    required
                ></input>
            </fieldset>
            <div class="flex flex-row gap-2">
                <div v-for="genre in selectedGenres">
                    {{ genre.genre_name }}
                </div>
            </div>
            <button 
                type="button"
                @click="() => genresOpen = !genresOpen">Genres</button>
                <div v-if="songStore.genres.length > 0 && genresOpen" class="flex flex-col top-0 bg-white">
                    <button 
                    type="button"
                    @click="() => selectGenre(genre)"
                    v-for="genre in songStore.genres">{{ genre.genre_name }}</button>
                </div>
            <button type="submit">Ok</button>
        </form>



    <form class="max-w-md border rounded-md flex flex-col gap-2" @submit.prevent="() => handleSubmitGenre()">
      <p class="font-bold underline">Add Genre</p>
      <fieldset class="flex flex-col">
        <label for="genre_name">Genre Name</label>
        <input 
        id="genre_name" v-model="genre.genre_name"
        required
        ></input>
      </fieldset>
      <button type="submit">Ok</button>
    </form>

    <form class="max-w-md border rounded-md flex flex-col gap-2" @submit.prevent="() => handleSubmitTag()">
      <p class="font-bold underline">Add Tag</p>
      <fieldset class="flex flex-col">
        <label for="name">Tag Name</label>
        <input 
        id="name" v-model="tag.name"
        required
        ></input>
        <label for="tag_description">Tag Description</label>
        <textarea
        id="tag_description" v-model="tag.description"
        required
        ></textarea>
      </fieldset>
      <button type="submit">Ok</button>
    </form>
  </main>
</template>
