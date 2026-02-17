<script setup lang="ts">
import { type Participant } from "@/stores/liveStore";
const props = defineProps<{ participant: Participant }>();

const audioTracks: ComputedRef<string[]> = computed(() => {
  const out = [];
  for (const track of props.participant.tracks.values()) {
    if (track.kind === "audio") out.push(track.subscribeURL);
  }
  return out;
});
</script>
<template>
  <div>
    <div style="display: none">
      <audio v-for="url in audioTracks" :src="url" autoplay></audio>
    </div>
    <ParticipantVideo v-bind="props" />
  </div>
</template>
