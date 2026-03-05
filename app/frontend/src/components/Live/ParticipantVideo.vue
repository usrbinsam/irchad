<script lang="ts" setup>
const props = withDefaults(
  defineProps<{
    title: string;
    // id: string;
    url: string;
    source: string;
    trackName: string;
    tag: "video" | "img";
  }>(),
  {
    tag: "video",
  },
);
const vol = ref(1.0);
const videoRef = useTemplateRef("video");

function syncVolume() {
  videoRef.value!.volume = vol.value;
}
</script>

<template>
  <v-img
    v-if="tag === 'img'"
    :src="url"
    content-class="video-feed d-flex align-end pb-2 px-2"
  >
    <v-chip
      color="rgba(0, 0, 0, 0.6)"
      size="small"
      class="title text-white font-weight-bold"
    >
      {{ title }}</v-chip
    >
    <template #placeholder>
      <div
        class="d-flex align-center justify-center fill-height bg-grey-darken-3"
      >
        <v-progress-circular indeterminate color="primary" />
      </div>
    </template>
  </v-img>
  <div v-else class="video-container">
    <video ref="video" :src="url" autoplay playsinline></video>
    <div class="video-overlay">
      <v-chip size="small" dark class="glass-chip">
        {{ trackName || `${title}'s Screen Share` }}</v-chip
      >
      <v-slider
        :max="1.0"
        :step="0.01"
        v-model="vol"
        @update:model-value="syncVolume"
        color="primary"
        thumb-label
        class="volume-slider"
      ></v-slider>
    </div>
  </div>
</template>

<style scoped>
.video-container {
  position: relative;
  width: 100%;
  background-color: #000;
}

.glass-chip {
  background-color: rgba(0, 0, 0, 0.4) !important;
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  text-shadow: 0px 1px 3px rgba(0, 0, 0, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.15) !important;
}

.video-container video {
  width: 100%;
  height: auto;
  display: block;
}

.video-overlay {
  position: absolute;
  top: 10px;
  left: 10px;
  z-index: 10;
  opacity: 0;
}

.video-container:hover .video-overlay {
  opacity: 1;
}

.video-feed {
  transform: translateZ(0);
  will-change: transform, opacity;
  contain: strict;
  object-fit: contain;
  width: 100%;
  height: 100%;
  aspect-ratio: 16/9;
}

.volume-slider {
  left: 10px;
  bottom: 10px;
}
</style>
