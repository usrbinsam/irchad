# IrChad

![IrChad Screnshot](/screenshots/IrChad.png)

> [!IMPORTANT]
> Preview builds are available [here](https://github.com/usrbinsam/irchad/releases) for Linux and Windows.
> Preview builds will connect to the IrChad test network.
> Expect bugs and general unstability at this time.

IrChad is an IRC client that tries to make the IRC experience feel more like Discord.
The full power of an IRC experience for your normie friends that don't like or understand
the traditional IRC experience. For seasoned IRC users, it will feel like a bad/
weird IRC client. For most Discord users, it will feel like a sub-par Discord experience.

IrChad is NOT a drop-in replacement for Discord. You must host and administer
your own IRCd.

## Features

|              | Windows | Linux | macOS |
| ------------ | ------- | ----- | ----- |
| Chat         | x       | x     | x     |
| Voice        | x       | x     | -     |
| Camera       | -       | x     | -     |
| Screen Share | x       | x     | -     |

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
