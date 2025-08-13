<script setup lang="ts">
import { useSongStore, type Song } from '@/stores/songStore';
import { ref } from 'vue';

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

const processing = ref<boolean>(false);

const handleSubmitSong = async () => {
  if (processing.value === true || song.value === null) {
    return;
  }

  processing.value = true;

  await songStore.createSong(song.value);

  processing.value = false;
}
</script>

<template>
  <main>
    <form class="flex flex-col gap-2" @submit.prevent="() => handleSubmitSong()">
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
  </main>
</template>