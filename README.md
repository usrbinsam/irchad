# IrChad

:warning: This software is under active development and does not have any release builds yet!

IrChad is an IRC client that tries to make the IRC experience feel more like Discord.
The full power of an IRC experience for your normie friends that don't like or understand
the traditional IRC experience. For seasoned IRC users, it will feel like a bad/
weird IRC client. For most Discord users, it will feel like a sub-par Discord experience.

IrChad is NOT a drop-in replacement for Discord. You must host and administer
your own IRCd.

## Motivation

I'm mainly building IrChad as the means of a **Discord Exit**. Age verification, obnoxious Nitro advertising, unfriendly to 3rd party SDK developers, politics, multiple security breaches, privacy concerns, and more.

IRCv3 has just enough support now that the IRC protocol can replace Discord.

## Goals

- Make IRC feel more like Discord without heavily modifying core IRC.
- Interoperate with existing IRC tech. Traditional IRC clients on the same IRC
  server should not have a degraded/more annoying experience where possible.

## Intended Environment

IrChad expects to speak IRC protocol over a WebSocket and IRCv3 support (batch,
multiline, typing, and metadata) - I recommend [ergo](https://github.com/ergochat/ergo).

You should be comfortable with IRC administration via a traditional IRC client
of your choice. IrChad has no plans to build UI elements for IRC administration.

IrChad connects to the server and joins every channel in the `LIST` output.

It currently only supports a single-network, but multi-network support should be
mostly trivial.

## Contributing

Contributions are welcome, there is a LOT to do.
See [Issues](https://github.com/usrbinsam/irchad/issues).

You will need:

- [Go](https://go.dev/doc/install)
- [Wails v3](https://v3alpha.wails.io/)
- [Node 24](https://nodejs.org/en/download) (using something like [fnm](https://github.com/Schniz/fnm) is recommended)
- [npm](https://npmjs.org/)
- [Docker](https://docs.docker.com/get-started/get-docker/) (or something to run
  containers)
- [Task](https://taskfile.dev/)
- [ffmpeg](https://www.ffmpeg.org/) + [libopus](https://opus-codec.org/downloads/) (if testing voice/video)

1. Rename `ircd.test.yaml` -> `ircd.yaml`
2. Start Ergo and LiveKit Server with `docker compose up -d`
3. Connect to ergo with an IRC client and join some channels
4. Start wails dev with `wails3 dev` in the `app/` folder

### Useful links

- [ergo manual](https://github.com/ergochat/ergo/blob/stable/docs/MANUAL.md)
- [IRCv3](https://ircv3.net/irc/)
- [Wails](https://v3alpha.wails.io/)
