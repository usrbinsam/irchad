declare module "irc-framework" {
  export interface IrcUser {
    nick: string;
    ident?: string;
    hostname?: string;
    modes?: string[];
    [key: string]: any;
  }
  //
  //   export class Client {
  //     join(channel: string, key?: string);
  //     constructor();
  //     connect(options: any): void;
  //     on(event: string, callback: (event: any) => void): void;
  //     channel(name: string): IrcChannel;
  //     changeNick(newNick: string);
  //     say(target: string, message: string);
  //     raw(v: string);
  //     requestCap(cap: string);
  //     list();
  //   }
  // }
  //
  // declare module "irc-framework/src/channel" {
  //   export class IrcChannel {
  //     constructor(irc_client: Client, channel_name: string, key?: string);
  //     name: string;
  //     users: IrcUser[];
  //     say(message: string): void;
  //     notice(message: string): void;
  //     part(message?: string): void;
  //     join(key?: string): void;
  //     mode(mode: string, param?: string): void;
  //   }
}
