<template>
  <v-app>
    <Login v-if="!ircStore.connected" />
    <Chat v-else />
  </v-app>
</template>

<script lang="ts" setup>
import { useIRCStore } from "@/stores/irc";
import Chat from "@/components/Chat.vue";
import Login from "@/components/Login.vue";

const ircStore = useIRCStore();
</script>
