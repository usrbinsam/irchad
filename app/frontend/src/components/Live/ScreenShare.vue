<script setup lang="ts">
import type { WindowData } from "@/bindings/IrChad/internal/live";
import { getWindows, shareWindow } from "@/live/liveProxy";
import { useLiveStore } from "@/stores/liveStore";
const liveStore = useLiveStore();
const { screenShareDialog } = storeToRefs(liveStore);

const windowList = ref([] as WindowData[]);

const selectedWindow = ref(0);
async function share() {
  await shareWindow(selectedWindow.value);
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
