<script setup>
import { useBufferStore } from "@/stores/bufferStore";
const { setActiveBuffer, buffers, activeBufferName } = useBufferStore();
</script>

<template>
  <v-list
    selectable
    :selected="[activeBufferName]"
    @click:select="(item) => setActiveBuffer(item.id)"
  >
    <v-list-item
      :value="bufName"
      v-for="(bufValue, bufName) in buffers"
      :key="bufName"
    >
      {{ bufName }}
      <template v-slot:append>
        <v-badge :content="bufValue.unseenCount.value" inline />
      </template>
    </v-list-item>
  </v-list>
</template>
