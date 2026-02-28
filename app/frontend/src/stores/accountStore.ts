import { defineStore } from "pinia";
import { ref } from "vue";

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

  function loadServer() {
      const v = localStorage.getItem('lastServer')
      if (v)
          server.value = JSON.parse(v)
  }

  function saveServer() {
      localStorage.setItem('lastServer', JSON.stringify(server.value))
  }

  loadServer()

  return {
    account,
    authError,
    authenticated,
    showRegistration,
    server,
    saveServer,
    setAuthenticated,
    setNick,
  };
});
