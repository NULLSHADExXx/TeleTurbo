<div align="center">

# TeleTurbo

**High-speed Telegram file downloader with parallel chunked downloads.**

Download files from any Telegram channel at maximum speed using multi-threaded parallel connections.

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Wails](https://img.shields.io/badge/Wails-v2-red?logo=webassembly)](https://wails.io)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Windows-blue)]()
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

</div>

---

## Download

| Platform | Download | Architecture |
|----------|----------|--------------|
| macOS | [TeleTurbo.dmg](https://github.com/NULLSHADExXx/TeleTurbo/releases/latest/download/TeleTurbo.dmg) | Apple Silicon (ARM64) |
| Windows | [TeleTurbo.exe](https://github.com/NULLSHADExXx/TeleTurbo/releases/latest/download/TeleTurbo.exe) | x86_64 |

## Features

- **Parallel Downloads** — 4-16 concurrent threads to saturate your bandwidth (5-10x faster than standard clients)
- **Private & Public Channels** — Download from any channel you're a member of
- **All Media Types** — Videos, documents, photos, audio — anything Telegram can send
- **Real-time Progress** — Live progress bar, speed, and file size tracking
- **Modern Dark UI** — Clean interface built with React
- **Cross-platform** — Native apps for macOS and Windows
- **Lightweight** — Single binary, no dependencies, under 10MB

## How It Works

Regular Telegram clients download files sequentially:
```
Part 1 → Wait → Part 2 → Wait → Part 3 → Wait...
```

TeleTurbo downloads in parallel:
```
Thread 1 → Part 1  ┐
Thread 2 → Part 2  │ simultaneous
Thread 3 → Part 3  │
Thread 4 → Part 4  ┘
```

## Quick Start

### 1. Get Telegram API Credentials

1. Go to [my.telegram.org](https://my.telegram.org)
2. Log in → "API development tools" → Create app
3. Copy your **App ID** and **App Hash**

### 2. Launch TeleTurbo

1. Enter your API credentials
2. Enter your phone number
3. Enter the OTP code sent to your Telegram app
4. Paste a Telegram link and download

### Supported Link Formats

```
https://t.me/c/1234567890/123     (Private channel)
https://t.me/channelname/123      (Public channel)
t.me/c/1234567890/123             (Short format)
```

## Build From Source

**Requirements:** Go 1.21+, Node.js 18+, [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
git clone https://github.com/NULLSHADExXx/TeleTurbo.git
cd TeleTurbo
wails build
```

The compiled app will be in `build/bin/`.

## Tech Stack

- **Backend** — Go + [gotd/td](https://github.com/gotd/td) (pure Go Telegram client)
- **Frontend** — React + TypeScript
- **Framework** — [Wails v2](https://wails.io) (Go + WebView)
- **Downloader** — gotd parallel chunked downloader

## License

MIT
