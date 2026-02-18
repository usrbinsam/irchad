<script setup lang="ts">
import { useIRCStore } from "@/stores/irc";
import { useBufferStore } from "@/stores/bufferStore";
import { useAccountStore } from "@/stores/accountStore";
import Live from "@/components/Live/Live.vue";
import { useLiveStore } from "@/stores/liveStore";

const bufferStore = useBufferStore();
const ircStore = useIRCStore();
const accountStore = useAccountStore();
const liveStore = useLiveStore();
</script>

<template>
  <div class="d-flex flex-row" style="height: 100vh">
    <router-view />
    <v-sheet border class="left-pane">
      <UserCard class="user-card" />
      <v-divider />
      <BufferList class="buffer-list" />
      <v-divider />
      <LiveControls v-if="!!liveStore.connected" class="live-controls" />
    </v-sheet>
    <Live />
    <div class="messages d-flex flex-column">
      <v-card-title>
        <v-row>
          <v-col cols="2">
            {{ bufferStore.activeBufferName }}
          </v-col>
          <v-col cols="3">
            {{ bufferStore.activeBuffer?.topic }}
          </v-col>
        </v-row>
      </v-card-title>
      <MessageList
        :messages="bufferStore.activeBuffer?.messages.value"
        :me="accountStore.account.nick"
      />
      <v-sheet>
        <InputBuffer
          @raw="(txt: string) => ircStore.client.raw(txt)"
          @send="ircStore.sendActiveBuffer"
        />
      </v-sheet>
    </div>
    <v-sheet class="user-list h-100" border>
      <UserList :users="bufferStore.activeBuffer?.users.value" />
    </v-sheet>
  </div>
</template>

<style>
.left-pane {
  display: flex;
  flex-direction: column;
  flex-shrink: 1;
}

.messages {
  height: 100%;
  flex: 3;
  justify-content: space-between;
}

.user-list {
  height: 100%;
  flex-shrink: 1;
}

.buffer-list {
  overflow-y: auto;
  height: 100%;
}
.live-controls {
  flex-shrink: 1;
}
</style>
