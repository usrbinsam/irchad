<script lang="ts" setup>
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
    <v-avatar v-if="!!avatar" class="avatar mr-2" size="22px">
      <v-img :src="avatar" />
    </v-avatar>
    <p class="nick text-grey flex-grow-1 text-left">{{ identity }}</p>
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
  </div>
</template>

<style scoped>
.avatar {
  text-align: left;
}
</style>
