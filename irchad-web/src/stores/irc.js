import { defineStore } from "pinia";
import { Client } from "irc-framework";

function autoNick() {
  const discriminiator = Math.round(Math.random() * 100);
  return `chad-${discriminiator}`;
}

export const useIRCStore = defineStore("irc", () => {
  const clientInfo = ref({
    nick: autoNick(),
    username: "chad",
    gecos: "IrChad",
  });

  const selfAvatar = ref("https://placekittens.com/128/128");

  function setAvatar(v) {
    selfAvatar.value = v;
    client.raw(`METADATA * SET avatar ${selfAvatar.value}`);
  }

  function setNick(v) {
    client.changeNick(v);
  }

  const buffers = ref({});
  const activeBufferName = ref();
  const metadata = ref({});

  const activeBuffer = computed(() => {
    if (activeBufferName.value) {
      return buffers.value[activeBufferName.value];
    }
  });

  function getMetadata(subject, key) {
    if (metadata.value[subject]) return metadata.value[subject][key];
  }

  function setActiveBuffer(bufferName) {
    activeBufferName.value = bufferName;
    resetBufferLastSeen(bufferName);
  }

  function resetBufferLastSeen(bufferName) {
    const buffer = getBuffer(bufferName);
    if (!buffer) return;
    buffer.lastSeenIdx = buffer.messages.length;
  }

  const client = new Client({
    enable_echomessage: true,
  });

  function connect() {
    client.requestCap("draft/metadata-2");
    client.requestCap("echo-message");
    const tls = location.protocol === "https:";
    client.connect({
      ...clientInfo.value,
      host: location.hostname,
      tls,
      port: location.port,
      version: "irchad irc-framework",
      path: "/ws",
    });
  }

  function sendActiveBuffer(message) {
    client.say(activeBufferName.value, message);
  }

  function isMe(target) {
    return target === clientInfo.value.nick;
  }

  function addBuffer(bufferName) {
    buffers.value[bufferName] = {
      messages: [],
      topic: "",
      users: [],
      lastSeenIdx: 0,
    };
    return buffers.value[bufferName];
  }

  function getBuffer(bufferName) {
    return buffers.value[bufferName];
  }

  function delBuffer(bufferName) {
    if (buffers.value[bufferName]) {
      delete buffers.value[bufferName];
    }
  }

  client.on("registered", function () {
    client.list();
    client.raw("METADATA * SUB avatar");
    client.raw(`METADATA * SET avatar ${selfAvatar.value}`);
  });

  client.on("nick", function ({ nick, new_nick }) {
    if (nick === clientInfo.value.nick) {
      clientInfo.value.nick = new_nick;
    }
    for (let buff of Object.values(buffers.value)) {
      const idx = buff.users.findIndex((u) => u.nick === nick);
      if (idx === -1) continue;
      buff.users[idx].nick = new_nick;
    }
  });

  client.on("unknown command", function (ircCommand) {
    if (ircCommand.command === "METADATA") {
      const from = ircCommand.params[0];
      const target = ircCommand.params[2];
      const key = ircCommand.params[1];
      const value = ircCommand.params[3];

      let subject = target;
      if (target === "*") {
        subject = from;
      }

      if (!metadata.value[subject]) {
        metadata.value[subject] = {};
      }

      metadata.value[subject][key] = value;
    }
  });

  client.on("channel list", (channels) =>
    channels.map((ch) => client.join(ch.channel)),
  );

  client.on("message", function (message) {
    let buffer;
    if (isMe(message.target)) {
      buffer = getBuffer(message.nick);
    } else {
      buffer = getBuffer(message.target);
    }
    buffer.messages.push(message);

    if (activeBufferName.value) {
      resetBufferLastSeen(activeBufferName.value);
    }
  });

  client.on("join", (ev) => {
    const nick = ev.nick;
    const channel = ev.channel;
    if (isMe(nick)) {
      addBuffer(channel);
      if (!activeBufferName.value) {
        activeBufferName.value = channel;
      }
      return;
    }

    const buffer = getBuffer(channel);
    if (!buffer) return;
    buffer.users.push({
      nick: ev.nick,
      gecos: ev.gecos,
      hostname: ev.hostname,
      ident: ev.ident,
    });
  });

  client.on("quit", function ({ nick }) {
    for (let buff of Object.values(buffers.value)) {
      const idx = buff.users.findIndex((u) => u.nick === nick);
      if (idx === -1) continue;
      buff.users.splice(idx, 1);
    }
  });

  client.on("topic", ({ topic, channel }) => {
    const buffer = getBuffer(channel);
    if (!buffer) return;
    buffer.topic = topic;
  });

  client.on("part", ({ nick, channel }) => {
    if (isMe(nick)) {
      delBuffer(channel);
    }
    const buffer = getBuffer(channel);
    if (!buffer) return;
    const idx = buffer.users.findIndex((u) => u.nick === nick);
    if (idx === -1) return;

    buffer.users.splice(idx, 1);
  });

  client.on("userlist", (ev) => {
    const buffer = getBuffer(ev.channel);
    if (!buffer) return;
    buffer.users = ev.users;
  });

  return {
    clientInfo,
    connect,
    client,
    buffers,
    activeBufferName,
    activeBuffer,
    sendActiveBuffer,
    setActiveBuffer,
    getMetadata,
    metadata,
    selfAvatar,
    setAvatar,
    setNick,
  };
});
