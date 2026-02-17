<script lang="ts" setup>
import { useLiveStore, type Participant } from "@/stores/liveStore";
const liveStore = useLiveStore();

function hasVideo(p: Participant) {
  for (const t of p.tracks.values()) {
    if (t.kind.toLowerCase() === "video") return true;
  }
  return false;
}
</script>

<template>
  <div class="d-flex flex-column">
    <div
      class="participant-row"
      v-for="p in liveStore.participants.values()"
      :key="p.identity"
    >
      <p>{{ p.identity }}</p>
      <v-icon v-if="hasVideo(p)">mdi-video-box</v-icon>
    </div>
  </div>
</template>

<style scoped>
.participant-row {
  margin-left: 30px;
}
</style>
