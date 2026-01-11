<script setup>
import { ref, onMounted } from "vue";
import { useIRCStore } from "@/stores/irc";

const store = useIRCStore();
const inputBuffer = ref();

onMounted(store.connect);

function send() {
  store.sendActiveBuffer(inputBuffer.value);
  inputBuffer.value = "";
}
</script>

<template>
  <div class="d-flex flex-row" style="height: 100vh">
    <v-sheet border class="buffers">
      <BufferList />
    </v-sheet>
    <div class="messages d-flex flex-column">
      <v-toolbar density="compact">
        <v-toolbar-title>
          <p>{{ store.activeBufferName }}</p>
          {{ store.activeBuffer?.topic }}
        </v-toolbar-title>
      </v-toolbar>
      <MessageList
        :messages="store.activeBuffer?.messages"
        :me="store.clientInfo.nick"
      />
      <v-sheet>
        <v-text-field
          variant="outlined"
          :placeholder="`Message ${store.activeBufferName}`"
          v-model="inputBuffer"
          hide-details
          class="ma-2"
          @keydown.enter.exact.prevent="send"
        />
      </v-sheet>
    </div>
    <v-sheet class="user-list h-100" border>
      <UserList :users="store.activeBuffer?.users" />
    </v-sheet>
  </div>
</template>

<style>
.buffers {
  height: 100%;
  flex: 1;
}

.messages {
  height: 100%;
  flex: 3;
  justify-content: space-between;
}

.user-list {
  height: 100%;
  flex: 1;
}
</style>
