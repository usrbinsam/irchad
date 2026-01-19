<script setup lang="ts">
import { computed, watch, useTemplateRef } from "vue";
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

function formatTime(ts: string) {
  const date = new Date(ts);
  return timeFormatter.format(date);
}

const chatHistory = useTemplateRef("chat-scrollback");
watch(
  () => props.messages,
  () =>
    nextTick(() => {
      chatHistory.value!.scrollTop = chatHistory.value!.scrollHeight;
    }),
  { deep: true },
);
</script>

<template>
  <v-sheet ref="chat-history" class="message-list d-flex">
    <div ref="chat-scrollback">
      <v-virtual-scroll height="100%" :items="messagesReverse">
        <template #default="{ item: msg }">
          <v-list-item
            density="compact"
            :prepend-avatar="store.getMetadata(msg.nick, 'avatar')"
          >
            <v-list-item-title>
              <span
                class="message-nick font-weight-bold"
                :class="{ 'text-primary': me === msg.nick }"
              >
                {{ msg.nick }}
              </span>
              <span class="message-time" v-if="!!msg.time">{{
                formatTime(msg.time)
              }}</span>
              <v-chip
                class="ml-2"
                v-bind="props"
                label
                color="purple-lighten-2"
                v-if="msg.kind === 'notice'"
                size="x-small"
                ><v-icon class="mr-2">mdi-eye</v-icon>Only visible to you
              </v-chip>
            </v-list-item-title>
            <pre style="white-space: pre">{{ msg.message }}</pre>
          </v-list-item></template
        >
      </v-virtual-scroll>
    </div>
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
