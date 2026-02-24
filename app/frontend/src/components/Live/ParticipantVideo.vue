<script lang="ts" setup>
const props = defineProps<{
  title: string;
  id: string;
  url: string;
  source: string;
  trackName: string;
}>();
</script>

<template>
  <v-img :src="url" content-class="video-feed d-flex align-end pb-2 px-2">
    <v-chip
      color="rgba(0,0,0,0.6)"
      size="small"
      class="title text-white font-weight-bold"
    >
      {{ title }}</v-chip
    >
    <v-chip v-if="source === 'SCREEN'">{{ trackName }}</v-chip>
    <template #placeholder>
      <div
        class="d-flex align-center justify-center fill-height bg-grey-darken-3"
      >
        <v-progress-circular indeterminate color="primary" />
      </div>
    </template>
  </v-img>
</template>

<style scoped>
.video-feed {
  transform: translateZ(0);
  will-change: transform, opacity;
  contain: strict;
  object-fit: contain;
  width: 100%;
  height: 100%;
}
</style>
