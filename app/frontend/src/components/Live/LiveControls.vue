<script setup lang="ts">
import { useLiveStore } from "@/stores/liveStore";
import {
  disconnect,
  publishCamera,
  unpublishCamera,
  unpublishScreenShare,
  setMuted,
} from "@/live/liveProxy";

const liveStore = useLiveStore();
const { screenShareDialog } = storeToRefs(liveStore);
</script>

<template>
  <v-card>
    <v-card-text>
      <v-sheet rounded="30" v-if="liveStore.screenShareEnabled">
        <v-icon color="error" class="mr-1" @click="unpublishScreenShare"
          >mdi-broadcast-off</v-icon
        >
        Screen sharing is active
      </v-sheet>
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
          @click="setMuted(true)"
        />

        <v-btn
          size="small"
          icon="mdi-microphone-off"
          v-else
          variant="text"
          color="error"
          @click="setMuted(false)"
        />

        <v-btn
          size="small"
          icon="mdi-webcam-off"
          v-if="!liveStore.camEnabled"
          color="error"
          variant="text"
          @click="publishCamera"
        ></v-btn>
        <v-btn
          size="small"
          icon="mdi-webcam"
          v-else
          variant="text"
          @click="unpublishCamera"
        ></v-btn>

        <v-btn
          size="small"
          icon="mdi-broadcast"
          variant="text"
          :color="liveStore.screenShareEnabled ? 'success' : 'grey'"
          @click="screenShareDialog = true"
        ></v-btn>
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
