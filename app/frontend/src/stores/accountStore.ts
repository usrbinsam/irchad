import { defineStore } from "pinia";
import { ref } from "vue";

import { type Config } from "@/bindings/IrChad/internal/network/models";
export const useAccountStore = defineStore("accountStore", () => {
  const authenticated = ref(false);
  const showRegistration = ref(false);
  const account = ref({
    nick: "",
    account: "",
    password: "",
  });

  const authError = ref({
    reason: "",
    message: "",
  });

  const discoveryURL = ref("http://127.0.0.1:8888");
  const config = ref({} as Config);
  const server = ref({
    host: "127.0.0.1",
    port: 8097,
    path: "/",
  });

  function setAuthenticated(v: boolean) {
    authenticated.value = v;
  }

  function setNick(v: string) {
    account.value.nick = v;
  }

  function loadAccount() {
    const v = localStorage.getItem("account");
    if (v) account.value = JSON.parse(v);
  }

  function saveAccount() {
    localStorage.setItem("account", JSON.stringify(account.value));
  }

  loadAccount();

  return {
    account,
    authError,
    authenticated,
    showRegistration,
    server,
    discoveryURL,
    config,
    saveAccount,
    setAuthenticated,
    setNick,
  };
});
