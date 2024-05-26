
## Support server

[We have a Revolt server dedicated to this project.](https://rvlt.gg/2Qn0ctjm)

## Why use revoltgo

At the time of writing, other (few) Revolt Go packages were simply unfeasible. They had:

- Hardcoded JSON payloads
- Poor API coverage and consistency
- Interface{} shoved in fields they were too lazy to add a struct for
- Hard-to-maintain codebase and odd design choices (wrapping Client and Time for each struct)
- ... this list can go on

## Features

RevoltGo as a project provides:

- Broader, up-to-date API coverage and functionality compared to other Revolt Go projects
- Extensive customisability due to low-level bindings
- Consistent, cleaner, and maintainable codebase

Additionally, revoltgo provides quality-of-life features such as:

- Permission calculator
- Lightweight ratelimit handling
- Automatic re-connects on websocket failures
- State/cache updates for members, roles, channels, and servers

## Getting started

### Installation

Assuming that you have a working Go environment ready, all you have to do is run the following command:

```bash
go get github.com/sentinelb51/revoltgo
```

If you do not have a Go environment ready, **[see how to set it up here](https://go.dev/doc/install)**

### Usage
Now that the package is installed, you will have to import it
```go
import "github.com/sentinelb51/revoltgo"
```

Then, construct a new **session** using your bot's token, and store it in a variable.
This "session" is a centralised store of all API and websocket methods at your fingertips, relevant to the bot you're about to connect with.
```go
session := revoltgo.New("your token here")
```

Finally, open the websocket connection to Revolt API. Your bot will attempt to login, and if successful, will receive events from the Revolt websocket about the world it's in.
Make sure to handle the error, as it can indicate any problem that could arise during the connection.
```go
err := session.Open()
```

To ensure the program keeps running, and accepts signals such as `Ctrl` + `C`, make a channel and wait for a signal from said channel:
```go
sc := make(chan os.Signal, 1)

signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
<-sc
```

When it's time to close the connection, simply close the session as demonstrated below.
```go
session.Close()
```

## Examples
The following examples are available in the [examples](https://github.com/sentinelb51/revoltgo/tree/main/examples) directory:
- **ping_bot.go**: A simple **bot** that responds to the `!ping` command.
- **selfbot.go**: A simple **self-bot** that responds to the `!ping` command.
