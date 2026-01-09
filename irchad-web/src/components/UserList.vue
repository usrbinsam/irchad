<script setup>
import { useIRCStore } from "@/stores/irc";
const props = defineProps(["users"]);
const store = useIRCStore();
</script>
<template>
  <v-list density="compact">
    <v-list-item
      v-for="user in users"
      :prepend-avatar="store.getMetadata(user.nick, 'avatar')"
    >
      {{ user.nick }}
    </v-list-item>
  </v-list>
</template>
