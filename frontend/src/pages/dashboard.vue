<script setup lang="ts">
import { useSongStore, type Song, type Genre, type SpecialTag } from '@/stores/songStore';
import { onMounted, ref } from 'vue';

onMounted(async () => {
  await songStore.getGenres();
  await songStore.getTags();
});

const songStore = useSongStore();

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

const handleSubmitSong = async () => {
  if (processing.value === true) {
    return;
  }

  processing.value = true;

  await songStore.createSong(song.value);

  processing.value = false;
}

const handleSubmitGenre = async () => {
  if (processing.value === true) {
    return;
  }

  processing.value = true;

  await songStore.createGenre(genre.value);

  processing.value = false;
}

const handleSubmitTag = async () => {
  if (processing.value === true) {
    return;
  }

  processing.value = true;

  await songStore.createTag(tag.value);

  processing.value = false;
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
