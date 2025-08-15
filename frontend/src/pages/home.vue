<script setup lang="ts">
import { useSongStore } from '@/stores/songStore';
import { onMounted } from 'vue';

const songStore = useSongStore();

onMounted(async () => {
  console.log("getting songs");
  await songStore.getSongs();
});
</script>

<template>
  <main>
    <div v-for="song in songStore.songs">
      <div class="w-[300px]">
<!--
        <iframe data-testid="embed-iframe" style="border-radius:12px" :src="song.spotify_embed_url" width="100%" height="100" frameBorder="0" allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture" loading="lazy"></iframe>
-->
      </div>
      <p>Album: {{ song.album_name }}</p>
      <p>Reason: {{ song.reason }}</p>
    </div>
  </main>
</template>
