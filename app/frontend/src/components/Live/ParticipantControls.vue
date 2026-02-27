<script lang="ts" setup>
import { SetParticipantVolume } from "@/bindings/IrChad/internal/live/livechat";
const props = defineProps<{ identity: string }>();

const volume = ref(1.0);

async function setVolume() {
  try {
    console.log(volume.value);
    await SetParticipantVolume(props.identity, volume.value);
  } catch (e) {
    alert(e);
  }
}
</script>
<template>
  <v-menu width="300px" activator="parent" :close-on-content-click="false">
    <v-list>
      <v-list-item>{{ identity }}</v-list-item>
      <v-list-item>
        <v-slider
          label="Volume"
          density="compact"
          color="primary"
          :max="1"
          :min="0"
          :step="0.1"
          @update:modelValue="setVolume"
          v-model="volume"
        />
      </v-list-item>
    </v-list>
  </v-menu>
</template>
