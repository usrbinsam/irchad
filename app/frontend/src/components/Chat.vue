<script setup lang="ts">
import { useIRCStore } from "@/stores/irc";
import { useBufferStore } from "@/stores/bufferStore";
import { useAccountStore } from "@/stores/accountStore";
import Live from "@/components/Live.vue";

const bufferStore = useBufferStore();
const ircStore = useIRCStore();
const accountStore = useAccountStore();

const live = useTemplateRef<InstanceType<typeof Live>>("live");
function joinLive() {
  live.value?.connect();
}
</script>

<template>
  <div class="d-flex flex-row" style="height: 100vh">
    <router-view />
    <v-sheet border class="buffers">
      <UserCard />
      <v-divider />
      <BufferList />
    </v-sheet>
    <Live ref="live" />
    <div class="messages d-flex flex-column">
      <v-card-title>
        <v-row>
          <v-col cols="2">
            {{ bufferStore.activeBufferName }}
          </v-col>
          <v-col cols="3">
            {{ bufferStore.activeBuffer?.topic }}
          </v-col>
          <v-col align="end">
            <v-btn icon="mdi-volume-high" @click="joinLive"> </v-btn>
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
