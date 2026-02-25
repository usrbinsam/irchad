<script setup lang="ts">
import type {
  ScreenShareOpts,
  WindowData,
} from "@/bindings/IrChad/internal/live";
import { getWindows, shareWindow } from "@/live/liveProxy";
import { useLiveStore } from "@/stores/liveStore";
const liveStore = useLiveStore();
const { screenShareDialog } = storeToRefs(liveStore);

const windowList = ref([] as WindowData[]);

const selectedWindow = ref(0);
const screenShareOpts = ref({
  FrameRate: 30,
  BitRate: 8000,
} as ScreenShareOpts);
async function share() {
  await shareWindow(selectedWindow.value, screenShareOpts.value);
  screenShareDialog.value = false;
}

onMounted(async () => (windowList.value = await getWindows()));
</script>

<template>
  <v-card title="Screen Share">
    <v-card-text>
      <v-radio-group v-model="selectedWindow">
        <v-radio
          v-for="w in windowList"
          :key="w.ID"
          :label="w.Title"
          :value="w.ID"
        >
        </v-radio>
      </v-radio-group>
      <v-radio-group label="Frame Rate" v-model="screenShareOpts.FrameRate">
        <v-radio :value="15" label="15 FPS" />
        <v-radio :value="30" label="30 FPS" />
        <v-radio :value="60" label="60 FPS" />
      </v-radio-group>
      <v-text-field
        type="number"
        label="Bit Rate"
        v-model="screenShareOpts.BitRate"
      />
    </v-card-text>
    <v-card-actions align="end">
      <v-btn
        color="success"
        class="mr-4"
        prepend-icon="mdi-broadcast"
        :disabled="!selectedWindow"
        @click="share"
        >Share</v-btn
      >
      <v-btn color="error" @click="screenShareDialog = false">Cancel</v-btn>
    </v-card-actions>
  </v-card>
</template>
