<script setup>
import { computed } from "vue";
import { useIRCStore } from "@/stores/irc";
const { buffers } = useIRCStore();
const store = useIRCStore();
const bufferList = computed(() => {
  return Object.keys(buffers);
});
</script>

<template>
  <v-list>
    <v-list-item v-for="buf in bufferList" @click="store.setActiveBuffer(buf)">
      {{ buf }}
    </v-list-item>
  </v-list>
</template>
