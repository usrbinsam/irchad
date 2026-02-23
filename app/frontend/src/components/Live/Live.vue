<script setup lang="ts">
import { useLiveStore } from "@/stores/liveStore";
const liveStore = useLiveStore();
const { videoDialog, screenShareDialog } = storeToRefs(liveStore);
</script>
<template>
  <div>
    <Participant
      v-for="participant in liveStore.participants.values()"
      :key="participant.identity"
      :participant="participant"
    />
  </div>
  <v-dialog v-model="screenShareDialog" persistent>
    <ScreenShare />
  </v-dialog>

  <v-dialog v-model="videoDialog" fullscreen>
    <v-card v-if="videoDialog">
      <v-card-title>
        Live Streams
        <v-icon @click="videoDialog = false"
          >mdi-chevron-down</v-icon
        ></v-card-title
      >
      <v-card-text>
        <ParticipantVideo
          v-for="participant in liveStore.participants.values()"
          v-bind="participant"
          :key="participant.identity"
        />
      </v-card-text>
    </v-card>
  </v-dialog>
</template>
