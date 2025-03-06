package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sentinelb51/revoltgo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN is not set in .env")
	}

	session := revoltgo.New(token)

	session.AddHandler(func(session *revoltgo.Session, r *revoltgo.EventReady) {
		fmt.Println("Ready to upload the RevoltGo logo when you type !upload")
	})

	session.AddHandler(func(session *revoltgo.Session, m *revoltgo.EventMessage) {

		if m.Content != "!upload" {
			return
		}

		// Read the logo.png file
		file, err := os.Open("logo.png")
		if err != nil {
			fmt.Printf("Failed to read the logo.png file: %s\n", err)
			fmt.Println("Don't tell me you deleted my beautiful logo...")
			return
		}

		// Create a file object with a name and the file reader
		payload := &revoltgo.File{
			Name:   "The name is arbitrary, but don't leave it empty or the media won't load",
			Reader: file,
		}

		// Upload the attachment to the server to get the attachment ID
		// This attachment ID will reference the uploaded file when we send it in a message
		attachment, err := session.AttachmentUpload(payload)
		if err != nil {
			fmt.Printf("Failed to upload attachment: %s\n", err)
			return
		}

		// Now, add the attachment ID to the Attachments []string slice in the MessageSend struct
		send := revoltgo.MessageSend{
			Content: "Here's your logo!", // You can omit this field if you want to send the attachment only
			Attachments: []string{
				attachment.ID,
			},
		}

		// Finally, send the message. Enjoy the logo.
		_, err = session.ChannelMessageSend(m.Channel, send)
		if err != nil {
			fmt.Printf("Failed to send message: %s\n", err)
		}

		fmt.Println("Logo uploaded!")
	})

	err = session.Open()
	if err != nil {
		panic(err)
	}

	sc := make(chan os.Signal, 1)

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	err = session.Close()
}
