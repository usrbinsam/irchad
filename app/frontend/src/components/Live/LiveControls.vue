<script setup lang="ts">
import { useLiveStore } from "@/stores/liveStore";
import { disconnect, publishMic } from "@/live/liveProxy";
const liveStore = useLiveStore();
</script>

<template>
  <v-card>
    <v-card-text>
      <p class="text-center">
        <v-icon class="mr-1 text-success">mdi-play-network</v-icon>
        <span class="font-weight-bold">{{ liveStore.connected }}</span>
      </p>

      <v-sheet class="d-flex flex-row justify-space-evenly align-center py-1">
        <v-btn
          size="small"
          icon="mdi-microphone"
          v-if="liveStore.micEnabled"
          variant="text"
        />

        <v-btn
          size="small"
          icon="mdi-microphone-off"
          v-else="liveStore.micEnabled"
          variant="text"
          color="error"
          @click="publishMic"
        />
        <v-btn size="small" icon="mdi-webcam" variant="text"></v-btn>

        <v-btn size="small" icon="mdi-monitor" variant="text"></v-btn>
        <v-btn
          size="small"
          icon="mdi-exit-run"
          variant="text"
          color="red"
          @click="disconnect"
        ></v-btn>
      </v-sheet>
    </v-card-text>
  </v-card>
</template>
