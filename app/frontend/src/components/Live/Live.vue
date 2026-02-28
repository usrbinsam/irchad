<script setup lang="ts">
import { useLiveStore } from "@/stores/liveStore";
const liveStore = useLiveStore();
const { videoDialog, screenShareDialog } = storeToRefs(liveStore);

const tracks = computed(() => {
  const out = [];
  for (const [participantIdentity, v] of liveStore.participants.entries()) {
    for (const track of v.tracks.values()) {
      if (track.kind.toLowerCase() !== "video") continue;

      out.push({
        title: participantIdentity,
        id: track.id,
        url: track.subscribeURL,
        source: track.source,
        trackName: track.trackName,
      });
    }
  }
  return out;
});
</script>
<template>
  <v-dialog v-model="screenShareDialog" persistent>
    <ScreenShare />
  </v-dialog>

  <v-dialog v-model="videoDialog" fullscreen height="100vh">
    <v-card height="100%">
      <v-card-title>
        Live Streams
        <v-icon @click="videoDialog = false"
          >mdi-chevron-down</v-icon
        ></v-card-title
      >
      <v-card-text
        class="d-flex flex-column h-100 flex-wrap ga-8 justify-content-center"
      >
        <ParticipantVideo
          class="participant-video"
          v-for="t in tracks"
          v-bind="t"
          :key="t.id"
        />
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.participant-video {
  border-radius: 4px;
  background: #000;
}
</style>
