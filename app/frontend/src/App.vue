<template>
  <v-app>
    <router-view />

    <!-- <Login v-if="!ircStore.connected" /> -->
    <!-- <Chat v-else /> -->
    <!-- <v-dialog max-width="450px" v-model="showRegistration"> -->
    <!--   <Register @success="showRegistration = false" /> -->
    <!-- </v-dialog> -->
  </v-app>
</template>

<script lang="ts" setup>
import { useIRCStore } from "@/stores/irc";
import Chat from "@/components/Chat.vue";
import Login from "@/components/Login.vue";
import Register from "@/components/Register.vue";
import { useAccountStore } from "./stores/accountStore";

const ircStore = useIRCStore();
const accountStore = useAccountStore();
const { showRegistration } = storeToRefs(accountStore);
</script>
