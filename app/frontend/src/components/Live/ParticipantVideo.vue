<script lang="ts" setup>
import type { Participant } from "@/stores/liveStore";
const props = defineProps<Participant>();

const videoTracks = computed(() => {
  const out = [];
  for (const track of props.tracks.values()) {
    if (track.kind.toLowerCase() !== "video") continue;
    // if (!track.show) continue;
    out.push(track);
  }
  return out;
});
</script>

<template>
  <div v-if="!!videoTracks.length">
    <v-img
      cover
      v-for="t in videoTracks"
      :key="t.id"
      :src="t.subscribeURL"
      content-class="d-flex align-end pb-2 px-2"
    >
      <v-chip
        color="rgba(0,0,0,0.6)"
        size="small"
        class="text-white font-weight-bold"
      >
        {{ identity }}</v-chip
      >
      <template #placeholder>
        <div
          class="d-flex align-center justify-center fill-height bg-grey-darken-3"
        >
          <v-progress-circular indeterminate color="primary" />
        </div>
      </template>
    </v-img>
  </div>
  <p v-else>No Video</p>
</template>

<style scoped>
.video-card {
  transform: translateZ(0);
  will-change: transform, opacity;
  contain: strict;
  aspect-ratio: 16/9;
}
</style>
