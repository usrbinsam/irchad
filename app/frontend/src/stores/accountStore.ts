import { defineStore } from "pinia";
import { ref } from "vue";

export const useAccountStore = defineStore("accountStore", () => {
  const authenticated = ref(false);
  const account = ref({
    nick: "",
    account: "",
    password: "",
  });

  const authError = ref({
    reason: "",
    message: "",
  });

  function setAuthenticated(v: boolean) {
    authenticated.value = v;
  }
  function setNick(v: string) {
    account.value.nick = v;
  }

  return { account, authError, authenticated, setAuthenticated, setNick };
});
