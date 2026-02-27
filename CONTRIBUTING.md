# Contributing

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
- [GStreamer](https://gstreamer.freedesktop.org/) (for MinGW for Windows)

1. Rename `ircd.test.yaml` -> `ircd.yaml`
2. Start Ergo and LiveKit Server with `docker compose up -d`
3. Connect to ergo with an IRC client and join some channels
4. Start wails dev with `wails3 dev` in the `app/` folder

## Windows

If you're developing on Windows, you'll need some extra dependencies.
You don't need to do this just to run IrChad.

- Install [MSYS2](https://www.msys2.org/)
- Launch the "MSYS2 MINGW64" terminal (the blue one)
- Install dependencies:

```sh
pacman -S --needed \
  mingw-w64-x86_64-binutils \
  mingw-w64-x86_64-crt-git \
  mingw-w64-x86_64-gcc \
  mingw-w64-x86_64-gdb \
  mingw-w64-x86_64-gdb-multiarch \
  mingw-w64-x86_64-go \
  mingw-w64-x86_64-headers-git \
  mingw-w64-x86_64-libmangle-git \
  mingw-w64-x86_64-libwinpthread \
  mingw-w64-x86_64-make \
  mingw-w64-x86_64-pkgconf \
  mingw-w64-x86_64-tools-git \
  mingw-w64-x86_64-winpthreads \
  mingw-w64-x86_64-gstreamer \
  mingw-w64-x86_64-gst-plugins-base \
  mingw-w64-x86_64-gst-plugins-good \
  mingw-w64-x86_64-gst-plugins-bad \
  mingw-w64-x86_64-gst-plugins-rs \
  mingw-w64-x86_64-gst-plugins-ugly \
  mingw-w64-x86_64-gst-libav \
  mingw-w64-x86_64-nodejs
```

- Add add `$GOPATH/bin` to the MSYS2 path
- Do a little health check:
  - `node -v`
  - `npm -v`
  - `wails3 doctor`

- Get started with `wails3 dev` in the `app/` folder

## Useful links

- [ergo manual](https://github.com/ergochat/ergo/blob/stable/docs/MANUAL.md)
- [IRCv3](https://ircv3.net/irc/)
- [Wails](https://v3alpha.wails.io/)
