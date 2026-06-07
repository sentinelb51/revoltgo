# Introduction

![RevoltGo logo RGO](https://github.com/sentinelb51/revoltgo/blob/main/logo.png)

### Purpose
RevoltGo is a low-level API wrapper for the [Revolt API](https://stoat.chat) focused on performance and maintainability.

### Context
Since July 2023 and until now, it has been the only mature Revolt Go library that, in my opinion, does things right.
Other projects had poor API coverage, no consistency or maintenance, and questionable code/design choices. They've since
been removed from [Awesome-Stoat's Go list](https://github.com/stoatchat/awesome-stoat#go), leaving this as the only
survivor.

### Applications
This library can be used for bots, self-bots, and clients

## Support
The fastest way to contact me is via the [Revolt support server dedicated for this project](https://rvlt.gg/R55WJBjx).
Of course, you can always create an issue or a PR.

# Features

### Performance
- **MessagePack websocket transport with code-generated serialisers**; no reflection, field look-ups, or runtime schema discovery. JSON inter-op is kept where the API requires it.
- **Pay only for what you use**; we inspect websocket event types and drop them without ever decoding it.
- **Low-level bindings, minimal opinionation**; you have full access to all the data the API/WS sends, no abstractions.
- **Zero-copy frame handling**; websocket frames are processed straight from the network buffer.
- **Lock-free event dispatch**; websocket frames are processed in parallel, and never content on a shared mutex.

### State
- **Optional, per-object caching**; track users, servers, channels, members, emojis, or none of it
- **Ergonomic reads**; slice getters, iterators, and counts for cached objects
- **Opportunistic refresh**; use HTTP responses to further synchronise the state
- **Consistent, race-protected**; the library does its own house-keeping so that your code never sees a half-updated world

### Authentication
- **Bots and self-bots**; both supported first-class.
- **Express login**; trade credentials for a re-usable token, get up and running fast.

### Developer experience
- **Targets latest Go releases and up-to-date dependencies**; no leaning on something three years stale
- **Consistent naming scheme**; every type's name builds on each-other, creating predictable patterns
- **REST API ratelimit handling** with safeguards against leaking your token
- **Utilities**; permission calculator, enums for almost everything, and helper functions
- **Debug toggles for HTTP and WebSocket** for when you need to see what's actually on the wire

# Getting started

## Installation
Assuming that you have a working Go environment ready, run one of the following commands to install the library.
If you do not have a Go environment ready, **[see how to set it up here](https://go.dev/doc/install)**

### Stable release

```bash
go get github.com/sentinelb51/revoltgo
```

### Latest release

```bash
go get github.com/sentinelb51/revoltgo@latest
```

## Usage
Now that the package is installed, you will have to import it

```go
import "github.com/sentinelb51/revoltgo"
```

Then, construct a new **session** using your bots token, and store it in a variable.
This "session" is a centralised store of all API and websocket methods at your fingertips, relevant to the bot you're
about to connect with.

```go
session := revoltgo.New("your token here")
```

Finally, open the websocket connection to Revolt API. Your bot will attempt to login, and if successful, will receive
events from the Revolt websocket about the world it's in.
Make sure to handle the error, as it can indicate any problem that could arise during the connection.

```go
err := session.Open()
```

To ensure the program keeps running, and accepts signals such as `Ctrl` + `C`, make a channel and wait for a signal from
said channel:

```go
sc := make(chan os.Signal, 1)

signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
<-sc
```

## Examples

The following examples are available in the [examples](https://github.com/sentinelb51/revoltgo/tree/main/examples)
directory:

- **ping_bot.go**: A **bot** that responds to the `!ping` command.
- **selfbot.go**: A **self-bot** that responds to the `!ping` command.
- **uploads.go**: A **bot** that uploads the RevoltGo logo using the command "!upload"

## Resource usage

The resource utilisation of the library depends on how many handlers are registered
and how many objects are cached in the state. More handlers will increase CPU usage, while
more objects in the state will increase memory usage.

For programs that need to be as lightweight as possible (and do not care about caching objects),
they can disable state tracking by passing a `revoltgo.StateConfig` to `Session.Open`:

```go
type StateConfig struct {
    TrackUsers        bool
    TrackServers      bool
    TrackChannels     bool
    TrackMembers      bool
    TrackEmojis       bool
    TrackAPICalls     bool
    TrackBulkAPICalls bool
}
```

### Windows platforms

Standalone, with state enabled, the library uses:

- ~0.00% CPU
- ~4-5 MB of RAM

The memory usage is expected to grow with state enabled as more objects get cached.

### Linux platforms

Not tested, but it's expected to be around the same.

## License: BSD 3-Clause

RevoltGo is licensed under the BSD 3-Clause License. What this means is that:

#### You are allowed to:

+ Modify the code, and distribute your own versions.
+ Use this library in personal, open-source, or commercial projects.
+ Include it in proprietary software, without making your project open-source.

#### You are not allowed to:

- Remove or alter the license and copyright notice.
- Use the name "RevoltGo" or its contributors for endorsements without permission.
- Hold the author liable for any damages arising from the use of the software.