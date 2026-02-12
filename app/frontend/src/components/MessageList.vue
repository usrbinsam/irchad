<script setup lang="ts">
import { computed, watch, useTemplateRef } from "vue";
import { useIRCStore } from "@/stores/irc";
import { useBufferStore } from "@/stores/bufferStore";
const props = defineProps(["messages", "me"]);

const store = useIRCStore();
const bufferStore = useBufferStore();
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

const typing = ref("");

// watch(
//   () => bufferStore.activeBuffer?typing,
//   () => {
//     if (!bufferStore.activeBuffer) {
//       typing.value = "";
//       return;
//     }
//
//     if (bufferStore.activeBuffer.typing.length === 0) {
//       typing.value = "";
//       return;
//     }
//
//     const names = bufferStore.activeBuffer?.typing.join(", ");
//     typing.value = names;
//   },
//   { deep: true },
// );
//
const currentTypers = computed(() => {
  const typers = bufferStore.activeBuffer?.typingMap.value;
  const active: string[] = [];
  const paused: string[] = [];

  if (!typers) {
    return { active, paused };
  }

  typers.value.forEach((typingInfo, nick) => {
    if (typingInfo.status === "active") active.push(nick);
    else paused.push(nick);
  });

  return { active, paused };
});
</script>

<template>
  <v-sheet ref="chat-history" class="message-list">
    <p v-if="typing" class="ma-2">
      <v-icon class="ma-1">mdi-chat-processing</v-icon>
      {{ currentTypers }}
    </p>
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
            <div v-html="msg.message" /> </v-list-item
        ></template>
      </v-virtual-scroll>
    </div>
  </v-sheet>
</template>

<style scoped>
.message-list {
  height: 100%;
  overflow-y: auto;
  display: flex;
  flex-direction: column-reverse;
}
.message-time {
  font-size: 0.65em;
  margin-left: 4px;
}
</style>
