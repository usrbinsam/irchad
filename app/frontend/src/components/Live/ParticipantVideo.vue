<script lang="ts" setup>
import type { Participant } from "@/stores/liveStore";
const props = defineProps<{ participant: Participant }>();

const videoTracks = computed(() => {
  const out = [];
  for (const track of props.participant.tracks.values()) {
    if (track.kind.toLowerCase() !== "video") continue;
    // if (!track.show) continue;
    out.push(track);
  }
  return out;
});
</script>

<template>
  <v-card>
    <v-card-text>
      <div class="d-flex">
        <img v-for="t in videoTracks" :key="t.id" :src="t.subscribeURL" />
      </div>
    </v-card-text>
  </v-card>
</template>
