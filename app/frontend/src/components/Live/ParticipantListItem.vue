<script lang="ts" setup>
import ParticipantControls from "./ParticipantControls.vue";

const props = defineProps<{
  identity: string;
  avatar?: string;
  speaking: boolean;
  webcam: boolean;
  streaming: boolean;
  muted: boolean;
}>();
</script>
<template>
  <div class="d-flex flex-row justify-space-between">
    <v-avatar
      v-if="!!avatar"
      :class="['avatar mr-2', { 'speaking-border': speaking }]"
      size="23px"
    >
      <v-img :src="avatar" />
    </v-avatar>
    <p
      :class="[
        'nick flex-grow-1 text-left',
        { 'text-white': speaking, 'text-grey': !speaking },
      ]"
    >
      {{ identity }}
    </p>
    <v-icon size="small" color="grey" v-if="webcam">mdi-webcam</v-icon>
    <v-icon size="small" color="grey" v-if="muted">mdi-microphone-off</v-icon>
    <v-chip
      label
      color="red-accent-4"
      variant="flat"
      size="x-small"
      v-if="streaming"
      >LIVE</v-chip
    >
    <participant-controls v-bind="props" />
  </div>
</template>

<style scoped>
.avatar {
  text-align: left;
}
.speaking-border {
  outline: 2px solid #90ee90; /* LightGreen */
  outline-offset: 1px;
}
</style>
