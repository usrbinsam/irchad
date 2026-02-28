<script setup lang="ts">
import { useAccountStore } from "@/stores/accountStore";
import { useIRCStore } from "@/stores/irc";
import { storeToRefs } from "pinia";
import { useRouter } from "vue-router";
import { onBeforeMount } from "vue";

const accountStore = useAccountStore();
const ircStore = useIRCStore();

const { account, server } = storeToRefs(accountStore);
const withAccount = ref(false);
const form = ref(false);
const connecting = ref(false);

const router = useRouter();
const nickInUse = ref(false);

onBeforeMount(() => {
  ircStore.client.on("nick in use", () => {
    nickInUse.value = true;
    connecting.value = false;
    ircStore.client.quit();
  });
});
function login() {
  const s = server.value;
  ircStore.connect(s.host, s.port, s.path);
  connecting.value = true;
  router.push({ name: "Chat" });
  accountStore.saveServer();
}

function required(v: any) {
  return !!v || "This field is required";
}
</script>

<template>
  <main class="w-25 mt-5 ma-auto">
    <v-card title="Connect to a server">
      <v-form @submit.prevent="login" v-model="form">
        <v-card-text>
          <v-text-field
            v-model="account.nick"
            label="Nickname"
            :rules="[required]"
            autofocus
          />
          <v-text-field v-model="server.host" label="Server Host" />
          <v-text-field v-model="server.port" label="Port" type="number" />
          <v-text-field v-model="server.path" label="Path" />
          <v-text-field
            v-if="withAccount"
            v-model="account.account"
            label="Username"
            role="username"
            :rules="[required]"
          />
          <v-text-field
            v-model="account.password"
            v-if="withAccount"
            label="Password"
            type="password"
            :rules="[required]"
          />
          <v-alert color="error" v-if="accountStore.authError.reason">
            {{ accountStore.authError.message }}
          </v-alert>
          <v-alert v-if="nickInUse" color="error">
            That nickname is already in use or registered to an account. Try a
            different nickname, or login.
          </v-alert>
          <v-checkbox v-model="withAccount" label="Login with an account" />
        </v-card-text>
        <v-card-actions>
          <v-btn
            :loading="connecting"
            type="submit"
            color="success"
            :disabled="!form"
            >Connect</v-btn
          >
        </v-card-actions>
      </v-form>
    </v-card>
  </main>
</template>
