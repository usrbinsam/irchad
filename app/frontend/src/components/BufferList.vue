<script setup>
import { useBufferStore } from "@/stores/bufferStore";
import { useLiveStore } from "@/stores/liveStore";
import { connect } from "@/live/liveProxy";
const { setActiveBuffer, buffers, activeBufferName } = useBufferStore();
const liveStore = useLiveStore();
const { videoDialog } = storeToRefs(liveStore);
</script>

<template>
  <v-list
    selectable
    density="compact"
    :selected="[activeBufferName]"
    @click:select="(item) => setActiveBuffer(item.id)"
  >
    <article
      v-for="(bufValue, bufName) in buffers"
      :key="bufName"
      :value="bufName"
    >
      <v-list-item density="compact">
        <div class="d-flex flex-direction-row">
          <v-icon
            class="mr-2"
            v-if="liveStore.connected === bufName"
            color="success"
            >mdi-volume-high</v-icon
          >
          <v-icon
            v-else
            label
            class="mr-2"
            @click="connect(bufName)"
            color="grey"
            >mdi-volume-high</v-icon
          >
          <p>{{ bufName }}</p>
        </div>
        <template v-slot:append>
          <v-badge :content="bufValue.unseenCount.value" inline />
        </template>
      </v-list-item>
      <article v-if="liveStore.connected === bufName">
        <v-btn
          variant="text"
          icon="mdi-monitor"
          size="small"
          @click="videoDialog = true"
        />
        <ParticipantList />
      </article>
    </article>
  </v-list>
</template>
