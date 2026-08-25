package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sentinelb51/revoltgo"
)

func main() {

	// Create a session from a simple token
	session := revoltgo.New("token here")

	// Add a function to print when the bot is ready
	revoltgo.AddHandler(session, func(session *revoltgo.Session, e *revoltgo.EventReady) {
		fmt.Printf("Ready to process commands from %d user(s) across %d server(s)\n", len(e.Users), len(e.Servers))
	})

	// Add a function to handle messages, offload it to the handleBotMessage function
	revoltgo.AddHandler(session, func(session *revoltgo.Session, event *revoltgo.EventMessage) {
		handleBotMessage(session, event)
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
}

func handleBotMessage(session *revoltgo.Session, m *revoltgo.EventMessage) {

	// If the message content is not "!ping", ignore the message.
	if m.Content != "!ping" {
		return
	}

	latency := session.WS.Latency()
	content := latency.String()

	if latency.Milliseconds() == 0 {
		content = "Still calculating, keep re-trying this command in 15-second intervals."
	}

	// Construct a message to send back to the channel.
	send := revoltgo.MessageSend{
		Content: content,
	}

	// Send the message to the channel.
	message, err := session.ChannelMessageSend(m.Channel, send)
	if err != nil {
		fmt.Println("Error sending message: ", err)
		return
	}

	fmt.Println("Sent message:", message.Content)

}
