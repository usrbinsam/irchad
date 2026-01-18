<script lang="ts" setup>
import { HookStatus, useIRCStore } from "@/stores/irc";
import { onBeforeMount, onBeforeUnmount } from "vue";
import { useRouter } from "vue-router";
const router = useRouter();
const ircStore = useIRCStore();
const emit = defineEmits(["success"]);

const RegistrationSuccess = new RegExp(/Account created/);
function interceptNickServMessage(message: { nick: string; message: string }) {
  if (message.nick !== "NickServ") {
    return HookStatus.HOOK_OK;
  }

  registering.value = false;

  if (message.message.match(RegistrationSuccess)) {
    emit("success");
  }

  return HookStatus.HOOK_EAT;
}

onBeforeMount(() => {
  ircStore.registerHook("message", interceptNickServMessage);
  ircStore.registerHook("notice", interceptNickServMessage);
});

onBeforeUnmount(() => {
  ircStore.unregisterHook("message", interceptNickServMessage);
  ircStore.unregisterHook("notice", interceptNickServMessage);
});

const newAccount = ref({
  email: "",
  username: "",
  password: "",
});

const form = ref(false);
const registering = ref(false);

function register() {
  if (!form.value) return;

  const { password, email } = newAccount.value;
  registering.value = true;
  ircStore.client.say("NickServ", `REGISTER ${password} ${email}`);
}

function cancel() {
  router.back();
}
const show = ref(true);
</script>
<template>
  <v-dialog v-model="show" max-width="400px" persistent>
    <v-card title="Account Registration">
      <v-form v-model="form" @submit.prevent="register">
        <v-card-text>
          <v-text-field
            :rules="[(v) => !!v || 'Email required']"
            v-model="newAccount.email"
            label="Email"
            role="email"
          />
          <v-text-field
            :rules="[(v) => !!v || 'Password required']"
            v-model="newAccount.password"
            label="Password"
            role="password"
            type="password"
          />
        </v-card-text>
        <v-card-actions>
          <v-btn
            :disabled="!form"
            type="submit"
            :loading="registering"
            text="Register"
            color="success"
          />
          <v-btn @click="cancel">Cancel</v-btn>
        </v-card-actions>
      </v-form>
    </v-card></v-dialog
  >
</template>
