<script setup lang="ts">
import { useLiveStore } from "@/stores/liveStore";
const liveStore = useLiveStore();
const { screenShareDialog } = storeToRefs(liveStore);
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
</template>
