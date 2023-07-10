
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

## Example

### Listening to events
Standalone, your bot will be pretty useless if it doesn't react to any events. The `revoltgo.Session` struct contains slices of event listener handlers, which you may append your functions to. For instance, here is an example of a bot that responds to `!ping` with the websocket latency. Make sure to invite your bot to the server if you don't seem to be receiving events:

```go
package main

import (
	"fmt"
	"github.com/sentinelb51/revoltgo"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// Create a new session using our bots token
	session := revoltgo.New("your token here")

	// Append a function that handles authenticated events.
	// This is just to see when the authentication is complete.
	session.HandlersAuthenticated = append(session.HandlersAuthenticated, func(session *revoltgo.Session, r *revoltgo.EventAuthenticated) {
		fmt.Println("Authentication complete")
	})

	// Append a function that handles ready events.
	// We will print some details from the event to the console when we receive EventReady.
	session.HandlersReady = append(session.HandlersReady, func(session *revoltgo.Session, r *revoltgo.EventReady) {
		fmt.Printf("Ready to process commands from %d user(s) across %d server(s)\n", len(r.Users), len(r.Servers))
	})

	// Append a function that handles message events. We will process any message that is "!ping"
	// and respond with the latency of the websocket connection, if possible.
	session.HandlersMessage = append(session.HandlersMessage, func(session *revoltgo.Session, m *revoltgo.EventMessage) {

		// If the message content is not "!ping", ignore the message.
		if m.Content != "!ping" {
			return
		}

		// Simulate a typing event for a second
		err := session.ChannelBeginTyping(m.Channel)
		if err != nil {
			fmt.Println("Error sending typing event: ", err)
		}

		time.Sleep(1 * time.Second)

		// Construct a message to send back to the channel.
		var send revoltgo.MessageSend

		// If the last heartbeat ack is zero, we can't do maths to get the latency.
		if !session.LastHeartbeatAck.IsZero() {
			latency := session.LastHeartbeatAck.Sub(session.LastHeartbeatSent)
			send.Content = fmt.Sprintf("Latency: %s", latency)
		} else {
			send.Content = "Latency data unavailable"
		}

		// Send the message to the channel.
		message, err := session.ChannelMessageSend(m.Channel, send)
		if err != nil {
			fmt.Println("Error sending message: ", err)
			return
		}

		fmt.Println("Sent message:", message.Content)
	})

	// Open the session.
	err := session.Open()
	if err != nil {
		panic(err)
	}

	// Wait for a signal; keep the bot running
	sc := make(chan os.Signal, 1)

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Close session.
	err = session.Close()
}
```

### Authenticating as a user (self-bot)
The logic is exactly the same as the above example, except the authentication flow is slightly different. In this case, you create a token by logging in using your credentials, then set it to the session, and finally open the websocket connection:

```go
data := revoltgo.LoginData{
	Email: "youremail@company.com",
	Password: "yourpassword",
	FriendlyName: "RevoltGo",
}

mfa, err := session.Login(data)
if err != nil {
	panic(err)
}

session.Token = mfa.Token

// ... the rest of the logic is exactly the same

session.Open()
```
