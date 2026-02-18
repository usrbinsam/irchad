<script setup lang="ts">
import { useLiveStore } from "@/stores/liveStore";
const liveStore = useLiveStore();
</script>

<template>
  <div>
    <Participant
      v-for="participant in liveStore.participants.values()"
      :key="participant.identity"
      :participant="participant"
    />
  </div>
</template>
