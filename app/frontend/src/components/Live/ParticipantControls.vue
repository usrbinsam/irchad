<script lang="ts" setup>
import { SetParticipantVolume } from "@/bindings/IrChad/internal/live/livechat";
const props = defineProps<{ identity: string }>();

const volume = ref(100);

async function setVolume(v: number) {
  try {
    let f = 0;
    if (v) {
      f = v / 100;
    }
    await SetParticipantVolume(props.identity, f);
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
          :max="200"
          :min="0"
          :step="1"
          @update:modelValue="setVolume"
          v-model="volume"
        >
          <template #append> {{ volume }}% </template>
        </v-slider>
      </v-list-item>
    </v-list>
  </v-menu>
</template>
