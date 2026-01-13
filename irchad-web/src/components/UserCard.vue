<script setup>
import { ref } from "vue";
import { useIRCStore } from "@/stores/irc";
import { storeToRefs } from "pinia";

const { selfAvatar } = storeToRefs(useIRCStore());
const { clientInfo, setAvatar, setNick, setBio } = useIRCStore();
const avatarDialog = ref(false);
const newNick = ref();
const newBio = ref();

function changeAvatar() {
  newNick.value = clientInfo.nick;
  avatarDialog.value = true;
}

function submitAvatar() {
  setAvatar(selfAvatar.value);
  avatarDialog.value = false;
  if (newNick.value && clientInfo.nick !== newNick.value) {
    setNick(newNick.value);
  }
  if (newBio.value) {
    setBio(newBio.value);
  }
}
</script>

<template>
  <v-dialog v-model="avatarDialog" max-width="800px">
    <v-card title="Edit Profile">
      <v-card-text>
        <v-text-field v-model="selfAvatar" label="Avatar URL" />
        <v-text-field v-model="newNick" label="Nick" />
        <v-text-field v-model="newBio" label="Bio" />
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
