import { defineStore, storeToRefs } from "pinia";
import { useRouter } from "vue-router";
import { Client } from "irc-framework";
import { useBufferStore } from "./bufferStore";
import { ref } from "vue";
import { useAccountStore } from "./accountStore";
import MarkdownIt from "markdown-it";
import DOMPurify from "dompurify";

export type HookFunction = (event: any) => HookStatus;

export enum HookStatus {
  HOOK_OK,
  HOOK_EAT,
}

interface Batch {
  type: string;
  target: string;
  messages: any[];
  params: string[];
}

const md = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: true,
});

function renderMessage(msgIn: string) {
  if (!msgIn) return "";
  const dirty = md.render(msgIn);
  return DOMPurify.sanitize(dirty, {
    ALLOWED_TAGS: [
      "li",
      "ul",
      "ol",
      "b",
      "i",
      "em",
      "strong",
      "a",
      "code",
      "pre",
      "br",
      "p",
      "h1",
      "h2",
      "h3",
      "h4",
      "h5",
    ],
    ALLOWED_ATTR: ["href", "target", "class"],
  });
}

export const useIRCStore = defineStore("ircStore", () => {
  const bufferStore = useBufferStore();
  const accountStore = useAccountStore();
  const connected = ref(false);
  const { authError } = storeToRefs(accountStore);
  const batches = new Map<string, Batch>();
  const hooks = {} as Record<string, HookFunction[]>;

  function registerHook(event: string, f: HookFunction) {
    if (hooks[event]) hooks[event].push(f);
    else hooks[event] = [f];
  }

  function runHook(eventName: string, eventArgs: any): HookStatus {
    if (!hooks[eventName]) return HookStatus.HOOK_OK;
    let lastRetVal = HookStatus.HOOK_OK;

    for (const hookFunction of hooks[eventName]) {
      const retVal = hookFunction(eventArgs);
      if (retVal === HookStatus.HOOK_EAT) return retVal;
      lastRetVal = retVal;
    }

    return lastRetVal;
  }

  function unregisterHook(eventName: string, f: HookFunction) {
    if (!hooks[eventName]) return;
    const idx = hooks[eventName].findIndex((item) => item === f);
    if (idx === -1) return;
    hooks[eventName].splice(idx, 1);
  }

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
    client.requestCap("draft/multiline");
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

    client.connect(connectParams);
  }

  function sendActiveBuffer(message: string) {
    if (!bufferStore.activeBuffer) {
      return;
    }
    send(bufferStore.activeBuffer?.name, message);
  }

  function genBatchId(): string {
    return Math.random().toString(36).substring(2, 10);
  }

  function send(target: string, message: string) {
    if (!message.includes("\n")) {
      client.say(target, message);
      return;
    }

    const split = message.split("\n");

    const batchId = genBatchId();
    client.raw(`BATCH +${batchId} draft/multiline ${target}`);
    split.forEach((line) => {
      client.raw(`@batch=${batchId} PRIVMSG ${target} :${line}`);
    });
    client.raw(`BATCH -${batchId}`);
  }

  function isMe(target: string) {
    return target === client.user.nick;
  }

  function setBio(v: string) {
    bio.value = v;
    client.raw(`METADATA * SET bio ${bio.value}`);
    storePrefs();
  }

  client.on("socket close", () => {
    batches.clear();
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

  const router = useRouter();
  client.on("registered", function () {
    connected.value = true;
    router.push({ name: "Chat" });
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

  function handleMessage(message: any) {
    let buffer;

    const retVal = runHook("message", message);
    if (retVal === HookStatus.HOOK_EAT) return;

    if (message.nick === "HistServ") return;
    if (isMe(message.target)) {
      buffer = bufferStore.getBuffer(message.nick);
    } else {
      buffer = bufferStore.getBuffer(message.target);
    }

    if (!buffer) {
      buffer = bufferStore.addBuffer(message.nick, {
        name: message.nick,
        channel: null,
      });
    }

    message.message = renderMessage(message.message);
    buffer.messages.push(message);

    if (
      bufferStore.activeBuffer &&
      bufferStore.activeBuffer.name === buffer.name
    ) {
      buffer.resetLastSeen();
    }
  }

  client.on(
    "message",
    function (message: {
      nick: string;
      target: string;
      tags: Record<string, string>;
    }) {
      const batchId = message.tags["batch"];

      if (batchId && batches.has(batchId)) {
        batches.get(batchId)?.messages.push(message);
        return;
      }

      handleMessage(message);
    },
  );

  client.on("notice", function (message) {
    const retVal = runHook("notice", message);
    if (retVal === HookStatus.HOOK_EAT) return;
    if (bufferStore.activeBuffer) {
      bufferStore.activeBuffer.messages.push({ ...message, kind: "notice" });
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

  client.on("batch start", (event: any) => {
    batches.set(event.id, {
      type: event.type,
      target: event.params[0],
      messages: [],
      params: event.params,
    });
  });

  function handleChatHistory(batch: Batch) {
    for (const message of batch.messages) {
      handleMessage(message);
    }
  }

  function handleMultiline(batch: Batch) {
    console.log(batch);
    let m = "";
    for (const message of batch.messages) {
      m += `${message.message}\n`;
    }
    handleMessage({ ...batch.messages[0], message: m });
  }

  client.on("batch end draft/multiline", (event: any) => {
    const batch = batches.get(event.id);
    if (!batch) return;

    handleMultiline(batch);
    batches.delete(event.id);
  });

  client.on("batch end chathistory", (event: any) => {
    const batch = batches.get(event.id);
    if (!batch) return;
    handleChatHistory(batch);
    batches.delete(event.id);
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
    registerHook,
    unregisterHook,
  };
});
