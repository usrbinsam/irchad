<script setup lang="ts">
import { useBufferStore } from "@/stores/bufferStore";
import { useIRCStore } from "@/stores/irc";
import { computed } from "vue";
const store = useIRCStore();
const bufferStore = useBufferStore();

const sortedUsers = computed(() => {
  const b = bufferStore.activeBuffer;
  if (!b) {
    return;
  }
  const allUsers = b.users;
  if (!allUsers) {
    return;
  }
  const u = [...allUsers.value];
  u.sort((a, b) => a.nick.localeCompare(b.nick));
  return u;
});
</script>

<template>
  <v-list density="compact">
    <v-list-item
      v-for="user in sortedUsers"
      :prepend-avatar="store.getMetadata(user.nick, 'avatar')"
      :title="user.nick"
    >
      <v-menu activator="parent">
        <v-card :title="user.nick">
          <v-card-text>
            <p v-text="store.metadata[user.nick]?.bio"></p>
          </v-card-text>
          <v-list density="compact">
            <v-list-item title="Ident"> {{ user.ident }}</v-list-item>
            <v-list-item title="Hostname"> {{ user.hostname }}</v-list-item>
          </v-list>
        </v-card>
      </v-menu>
    </v-list-item>
  </v-list>
</template>
