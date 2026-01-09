<script setup>
import { computed } from "vue";
import { useIRCStore } from "@/stores/irc";
const props = defineProps(["messages", "me"]);

const store = useIRCStore();
const messagesReverse = computed(() => {
  if (props.messages) {
    return [...props.messages];
  }
});

const timeFormatter = new Intl.DateTimeFormat("en-US", {
  hour: "numeric",
  minute: "2-digit",
  hour12: true,
});
function formatTime(ts) {
  const date = new Date(ts);
  return timeFormatter.format(date);
}
</script>

<template>
  <v-sheet class="message-list d-flex">
    <v-list>
      <v-list-item
        v-for="msg in messagesReverse"
        density="compact"
        :prepend-avatar="store.getMetadata(msg.nick, 'avatar')"
      >
        <v-list-item-title>
          <span
            class="message-nick"
            :class="{ 'text-primary': me === msg.nick }"
          >
            {{ msg.nick }}
          </span>
          <span class="message-time" v-if="!!msg.time">{{
            formatTime(msg.time)
          }}</span>
        </v-list-item-title>
        <v-list-item-subtitle>
          {{ msg.message }}
        </v-list-item-subtitle>
      </v-list-item>
    </v-list>
  </v-sheet>
</template>

<style scoped>
.message-list {
  height: 100%;
  overflow-y: auto;
  flex-direction: column-reverse;
}
.message-time {
  font-size: 0.65em;
  margin-left: 4px;
}
</style>
