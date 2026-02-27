<script lang="ts" setup>
import { SetParticipantVolume } from "@/bindings/IrChad/internal/live/livechat";
const props = defineProps<{ identity: string }>();

const volume = ref(1.0);

async function setVolume() {
  try {
    console.log(volume.value);
    await SetParticipantVolume(props.identity, volume.value);
  } catch (e) {}
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
          :max="2"
          :min="0"
          :step="0.01"
          @update:modelValue="setVolume"
          v-model="volume"
        >
          <template #append> {{ (volume * 100).toFixed(0) }}% </template>
        </v-slider>
      </v-list-item>
    </v-list>
  </v-menu>
</template>
