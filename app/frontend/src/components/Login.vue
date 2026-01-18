<script setup lang="ts">
import { useAccountStore } from "@/stores/accountStore";
import { useIRCStore } from "@/stores/irc";
import { storeToRefs } from "pinia";
import { useRouter } from "vue-router";
const accountStore = useAccountStore();
const ircStore = useIRCStore();

const { account } = storeToRefs(accountStore);
const withAccount = ref(false);
const form = ref(false);
const connecting = ref(false);

const router = useRouter();

function login() {
  ircStore.connect();
  connecting.value = true;
  router.push({ name: "Chat" });
}

function required(v: any) {
  return !!v || "This field is required";
}
</script>

<template>
  <main class="w-25 mt-5 ma-auto">
    <v-card title="Login to IrChad">
      <v-form @submit.prevent="login" v-model="form">
        <v-card-text>
          <v-text-field
            v-model="account.nick"
            label="Nickname"
            :rules="[required]"
          />
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
