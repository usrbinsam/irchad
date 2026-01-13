<script setup>
import { useIRCStore } from "@/stores/irc";
import { computed } from "vue";
const props = defineProps(["users"]);
const store = useIRCStore();

const sortedUsers = computed(() => {
  if (!props.users) return [];
  const u = [...props.users];
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
