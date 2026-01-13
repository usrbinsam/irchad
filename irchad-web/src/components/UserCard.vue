<script setup>
import { ref } from "vue";
import { useIRCStore } from "@/stores/irc";
import { storeToRefs } from "pinia";

const { selfAvatar } = storeToRefs(useIRCStore());
const { client, clientInfo, setAvatar, setNick } = useIRCStore();
const avatarDialog = ref(false);
const newNick = ref();

function changeAvatar() {
  avatarDialog.value = true;
}

function submitAvatar() {
  setAvatar(selfAvatar.value);
  avatarDialog.value = false;
  if (newNick.value && clientInfo.nick !== newNick.value) {
    console.log("nick changed");
    client.changeNick(newNick.value);
  }
}
</script>

<template>
  <v-dialog v-model="avatarDialog">
    <v-card title="Edit Profile">
      <v-card-text>
        <v-text-field v-model="selfAvatar" label="Avatar URL" />
        <v-text-field v-model="newNick" label="Nick" />
      </v-card-text>
      <v-card-actions>
        <v-btn text="OK" @click="submitAvatar" />
      </v-card-actions>
    </v-card>
  </v-dialog>
  <v-card>
    <v-card-title>
      <v-avatar @click="changeAvatar" v-if="selfAvatar" :image="selfAvatar" />
      {{ clientInfo.nick }}
    </v-card-title>
    <v-card-text> </v-card-text>
  </v-card>
</template>
