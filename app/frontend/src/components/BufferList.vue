<script setup>
import { useBufferStore } from "@/stores/bufferStore";
import { useLiveStore } from "@/stores/liveStore";
import { connect } from "@/live/liveProxy";
const { setActiveBuffer, buffers, activeBufferName } = useBufferStore();
const liveStore = useLiveStore();
</script>

<template>
  <v-list
    selectable
    :selected="[activeBufferName]"
    @click:select="(item) => setActiveBuffer(item.id)"
  >
    <v-list-item
      v-for="(bufValue, bufName) in buffers"
      :key="bufName"
      :value="bufName"
    >
      <div class="d-flex flex-direction-row">
        <v-icon
          class="mr-2"
          v-if="liveStore.connected === bufName"
          color="success"
          >mdi-volume-high</v-icon
        >
        <v-icon v-else label class="mr-2" @click="connect(bufName)" color="grey"
          >mdi-volume-high</v-icon
        >
        <p>{{ bufName }}</p>
      </div>
      <template v-slot:append>
        <v-badge :content="bufValue.unseenCount.value" inline />
      </template>

      <ParticipantList v-if="liveStore.connected === bufName" />
    </v-list-item>
  </v-list>
</template>
