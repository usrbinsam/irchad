<script setup>
import { useIRCStore } from "@/stores/irc";

const { setActiveBuffer, buffers, activeBufferName } = useIRCStore();
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
