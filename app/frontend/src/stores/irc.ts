import { defineStore, storeToRefs } from "pinia";
import { Client } from "irc-framework";
import { useBufferStore } from "./bufferStore";
import { ref } from "vue";
import { useAccountStore } from "./accountStore";

export const useIRCStore = defineStore("ircStore", () => {
  const bufferStore = useBufferStore();
  const accountStore = useAccountStore();
  const connected = ref(false);
  const { authError } = storeToRefs(accountStore);

  const selfAvatar = ref("https://placekittens.com/128/128");
  const bio = ref();

  loadPrefs();

  function setAvatar(v: string) {
    selfAvatar.value = v;
    client.raw(`METADATA * SET avatar ${selfAvatar.value}`);
    storePrefs();
  }

  function loadPrefs() {
    // const v = localStorage.getItem("prefs");
    // if (v === null) return;
    //
    // const prefs = JSON.parse(v);
    // if (!prefs) return;
    //
    // if (prefs.avatar) {
    //   selfAvatar.value = prefs.avatar;
    // }
    //
    // if (prefs.nick) {
    //   clientInfo.value.nick = prefs.nick;
    // }
    //
    // if (prefs.bio) {
    //   bio.value = prefs.bio;
    // }
  }

  function storePrefs() {
    // localStorage.setItem(
    //   "prefs",
    //   JSON.stringify({
    //     nick: clientInfo.value.nick,
    //     avatar: selfAvatar.value,
    //     bio: bio.value,
    //   }),
    // );
  }

  function setNick(v: string) {
    client.changeNick(v);
  }

  const metadata = ref({} as Record<string, any>);

  function getMetadata(subject: string, key: string) {
    if (metadata.value[subject]) return metadata.value[subject][key];
  }

  const client = markRaw(new Client());

  function connect() {
    client.requestCap("draft/metadata-2");
    client.requestCap("echo-message");
    client.requestCap("chathistory");
    const tls = location.protocol === "https:";
    const connectParams = {
      host: location.hostname,
      port: location.port,
      tls,
      version: "irchad on irc-framework",
      path: "/ws",
      account: accountStore.account.account
        ? {
            account: accountStore.account.account,
            password: accountStore.account.password,
          }
        : undefined,
      sasl_disconect_on_fail: true,
      nick: accountStore.account.nick,
    };

    console.log(connectParams);
    client.connect(connectParams);
  }

  function sendActiveBuffer(message: string) {
    if (!bufferStore.activeBuffer) {
      return;
    }
    bufferStore.activeBuffer.channel.say(message);
  }

  function isMe(target: string) {
    console.log(client.user.nick);
    return target === client.user.nick;
  }

  function setBio(v: string) {
    bio.value = v;
    client.raw(`METADATA * SET bio ${bio.value}`);
    storePrefs();
  }

  client.on("socket close", () => {
    connected.value = false;
  });

  client.on("loggedin", () => {
    accountStore.setAuthenticated(true);
    authError.value = {
      reason: "",
      message: "",
    };
  });

  client.on(
    "sasl failed",
    ({ reason, message }: { reason: string; message: string }) => {
      authError.value = {
        reason,
        message,
      };
    },
  );

  client.on("registered", function () {
    connected.value = true;
    client.list();
    client.raw("METADATA * SUB avatar");
    client.raw("METADATA * SUB bio");
    client.raw(`METADATA * SET avatar ${selfAvatar.value}`);
    client.raw(`METADATA * SET bio ${bio.value}`);
  });

  client.on(
    "nick",
    function ({ nick, new_nick }: { nick: string; new_nick: string }) {
      if (nick === client.user.nick) {
        accountStore.setNick(new_nick);
      }
      for (let buffName in bufferStore.buffers.value) {
        const buff = bufferStore.getBuffer(buffName);
        if (!buff) {
          console.log(`${buffName} not found`);
          continue;
        }
        const idx = buff.users.findIndex((u) => u.nick === nick);
        if (idx === -1) {
          console.log(`${nick} not found in ${buffName}`);
          continue;
        }
        buff.users[idx].nick = new_nick;
      }
      metadata.value[new_nick] = { ...metadata.value[nick] };
      delete metadata.value[nick];
    },
  );

  client.on(
    "unknown command",
    function (ircCommand: { command: string; params: string[] }) {
      if (ircCommand.command === "METADATA") {
        const from = ircCommand.params[0];
        const target = ircCommand.params[2];
        const key = ircCommand.params[1];
        const value = ircCommand.params[3];

        let subject = target;
        if (target === "*") {
          subject = from;
        }
        if (!subject || !key || !value) return;

        if (!metadata.value[subject]) {
          metadata.value[subject] = {};
        }

        metadata.value[subject][key] = value;
      }
    },
  );

  client.on("channel list", (channels: { channel: string }[]) =>
    channels.map((ch) => client.join(ch.channel)),
  );

  client.on(
    "tagmsg",
    ({
      nick,
      tags,
      target,
    }: {
      nick: string;
      tags: string[];
      target: string;
    }) => {
      console.log(nick, tags, target);
    },
  );

  client.on("message", function (message: { nick: string; target: string }) {
    let buffer;
    if (message.nick === "HistServ") return;
    if (isMe(message.target)) {
      buffer = bufferStore.getBuffer(message.nick);
    } else {
      buffer = bufferStore.getBuffer(message.target);
    }
    if (!buffer) {
      return;
    }

    buffer.messages.push(message);

    if (
      bufferStore.activeBuffer &&
      bufferStore.activeBuffer.name === buffer.name
    ) {
      buffer.resetLastSeen();
    }
  });

  client.on("join", ({ nick, channel }: { nick: string; channel: string }) => {
    if (isMe(nick)) {
      bufferStore.addBuffer(channel, {
        name: channel,
        channel: client.channel(channel),
      });
      if (!bufferStore.activeBuffer) {
        bufferStore.setActiveBuffer(channel);
      }
      client.raw("CHATHISTORY LATEST " + channel + " * 200");
      return;
    }

    const buffer = bufferStore.getBuffer(channel);
    if (!buffer) return;
    buffer.syncUsers();
    buffer.users.push({
      nick: nick,
    });
  });

  client.on("quit", function ({ nick }: { nick: string }) {
    for (let buff of Object.values(bufferStore.buffers)) {
      const idx = buff.users.findIndex((u) => u.nick === nick);
      if (idx === -1) continue;
      buff.users.splice(idx, 1);
    }
  });

  client.on(
    "topic",
    ({ topic, channel }: { topic: string; channel: string }) => {
      const buffer = bufferStore.getBuffer(channel);
      if (!buffer) return;
      buffer.topic = topic;
    },
  );
  client.on("part", ({ nick, channel }: { nick: string; channel: string }) => {
    if (isMe(nick)) {
      bufferStore.delBuffer(channel);
    }
    const buffer = bufferStore.getBuffer(channel);
    if (!buffer) return;
    const idx = buffer.users.findIndex((u) => u.nick === nick);
    if (idx === -1) return;

    buffer.users.splice(idx, 1);
  });

  client.on("userlist", (ev: { channel: string; users: any[] }) => {
    const buffer = bufferStore.getBuffer(ev.channel);
    if (!buffer) return;
    buffer.users = ev.users;
  });

  return {
    connect,
    client,
    sendActiveBuffer,
    getMetadata,
    metadata,
    selfAvatar,
    setAvatar,
    setNick,
    setBio,
    bio,
    connected,
  };
});
