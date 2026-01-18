<script setup>
import { useBufferStore } from "@/stores/bufferStore";
const { setActiveBuffer, buffers, activeBufferName } = useBufferStore();

const store = useBufferStore();
</script>

<template>
  <v-list
    selectable
    :selected="[activeBufferName]"
    @click:select="(item) => setActiveBuffer(item.id)"
  >
    <v-list-item :value="bufName" v-for="(bufValue, bufName) in buffers">
      {{ bufName }}
      <template v-slot:append>
        <v-badge
          :content="bufValue.messages.length - bufValue.lastSeenIdx"
          inline
        />
      </template>
    </v-list-item>
  </v-list>
</template>
