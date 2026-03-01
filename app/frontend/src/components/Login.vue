<script setup lang="ts">
import { useAccountStore } from "@/stores/accountStore";
import { useIRCStore } from "@/stores/irc";
import { storeToRefs } from "pinia";
import { useRouter } from "vue-router";
import { onBeforeMount } from "vue";
import { NetworkService } from "@/bindings/IrChad/internal/network";

const accountStore = useAccountStore();
const ircStore = useIRCStore();

const { account, config, discoveryURL } = storeToRefs(accountStore);
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

// onMounted(() => {
//   const a = account.value;
//   if (a.account && a.nick && a.password) {
//     login();
//   }
// });

async function discover() {
  const config = await NetworkService.Connect(
    discoveryURL.value,
    account.value.nick,
    account.value.account,
    account.value.password,
  );
  return config;
}

async function login() {
  const discoveryConfig = await discover();
  if (!discoveryConfig) return;
  const s = URL.parse(discoveryConfig.irc.server);
  if (!s) return;
  config.value = discoveryConfig;

  ircStore.connect(s.protocol === "wss:", s.hostname, s.port, s.pathname);
  router.push({ name: "Chat" });
  accountStore.saveAccount();
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
          <v-text-field
            v-if="withAccount"
            v-model="account.account"
            label="Account Name"
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
          <v-text-field v-model="discoveryURL" label="Server" />
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
