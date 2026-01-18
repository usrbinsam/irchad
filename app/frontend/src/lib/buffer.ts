"use strict";

import IrcChannel from "irc-framework/src/channel";
import IrcUser from "irc-framework/src/user";

export interface BufferOptions {
  channel: typeof IrcChannel;
  name: string;
  metadata?: Record<string, any>;
}

export class Buffer {
  name: string;
  options: BufferOptions;
  channel: IrcChannel;
  metadata: Record<string, any>;
  lastSeenIdx: number;
  messages: any[];
  typing: string[];
  users: IrcUser[];
  topic: string | null;

  constructor(options: BufferOptions) {
    this.options = options || null;

    this.name = options.name;

    this.channel = options.channel || null;

    this.metadata = options.metadata || {};
    this.messages = [];
    this.lastSeenIdx = 0;

    this.typing = [];
    this.topic = options.topic || null;

    if (this.channel) this.syncUsers();
  }

  syncUsers() {
    this.users = [...this.channel.users];
  }

  kind() {
    return this.name.startsWith("#") ? "channel" : "pm";
  }

  resetLastSeen() {
    this.lastSeenIdx = 0;
  }
}
