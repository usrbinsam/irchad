<script lang="ts" setup>
import { useLiveStore, type Participant } from "@/stores/liveStore";

const liveStore = useLiveStore();

const props = defineProps<{
  participant: Participant;
}>();

const trackVideos = computed(() => {
  const out = [];

  for (const t of props.participant.tracks.values()) {
    if (t.kind.toLowerCase() !== "video") continue;

    const icon = t.source === "camera" ? "mdi-video-box" : "mdi-monitor";
    out.push({
      icon,
      trackId: t.id,
    });
  }

  return out;
});

function showTrack(trackId: string) {
  liveStore.showTrack(props.participant.identity, trackId);
}
</script>
<template>
  <div class="d-flex flex-row">
    <p>{{ participant.identity }}</p>
    <v-icon
      v-for="tIcon in trackVideos"
      :key="tIcon.trackId"
      @click="showTrack(tIcon.trackId)"
      >{{ tIcon.icon }}</v-icon
    >
  </div>
</template>
