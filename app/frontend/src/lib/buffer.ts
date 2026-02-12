"use strict";
import { ref, type Ref } from "vue";
import IrcChannel from "irc-framework/src/channel";
import IrcUser from "irc-framework/src/user";

export interface BufferOptions {
  channel: typeof IrcChannel;
  name: string;
  metadata?: Record<string, any>;
  topic: string;
}

type TypingInfo = {
  ts: number;
  status: "active" | "paused";
};

export class Buffer {
  name: string;
  options: BufferOptions;
  channel: IrcChannel;
  metadata: Record<string, any>;
  lastSeenIdx: Ref<number>;
  messages: Ref<any[]>;
  typingMap: Ref<Map<string, TypingInfo>>;
  users: Ref<IrcUser[]>;
  topic: string | null;
  unseenCount: Ref<number>;
  pingCount: Ref<number>;

  constructor(options: BufferOptions) {
    this.options = options || null;

    this.name = options.name;

    this.channel = options.channel || null;

    this.metadata = options.metadata || {};
    this.messages = ref([]);
    this.lastSeenIdx = ref(0);

    this.topic = options.topic || null;
    this.typingMap = ref(new Map<string, TypingInfo>());
    this.users = ref([]);

    this.pingCount = ref(0);
    this.unseenCount = ref(0);
    if (this.channel) this.syncUsers();
  }

  syncUsers() {
    this.users.value = [...this.channel.users];
  }

  kind() {
    return this.name.startsWith("#") ? "channel" : "pm";
  }

  resetLastSeen() {
    this.unseenCount.value = 0;
  }

  addMessage(message: any) {
    this.messages.value.push(message);
    if (message.nick && this.typingMap.value.has(message.nick)) {
      this.typingMap.value.delete(message.nick);
    }
    this.unseenCount.value++;
  }

  typingActive(nick: string) {
    this.typingMap.value.set(nick, {
      ts: new Date().getTime(),
      status: "active",
    });
    const now = new Date().getTime();
    setTimeout(() => {
      const typer = this.typingMap.value.get(nick);
      if (!typer) return;

      if (now - typer.ts >= 6000 && typer.status === "active") {
        this.typingMap.value.delete(nick);
      }
    }, 6 * 1000);
  }

  typingPaused(nick: string) {
    this.typingMap.value.set(nick, {
      ts: new Date().getTime(),
      status: "paused",
    });

    const now = new Date().getTime();
    setTimeout(() => {
      const typer = this.typingMap.value.get(nick);
      if (!typer) return;

      if (now - typer.ts >= 30000 && typer.status === "paused") {
        this.typingMap.value.delete(nick);
      }
    }, 30 * 1000);
  }

  typingDone(nick: string) {
    this.typingMap.value.delete(nick);
  }

  getTyping(): Record<"active" | "paused", string[]> {
    const active: string[] = [];
    const paused: string[] = [];
    this.typingMap.value.forEach((value, key) => {
      if (value.status === "active") active.push(key);
      else if (value.status === "paused") paused.push(key);
    });
    return { active, paused };
  }
}
