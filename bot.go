package revoltgo

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"
)

// Bot struct.
type Bot struct {
	Client    *Session
	CreatedAt time.Time

	ID              string `json:"_id"`
	OwnerID         string `json:"owner"`
	Token           string `json:"token"`
	IsPublic        bool   `json:"public"`
	InteractionsUrl string `json:"interactionsURL"`
}

// Fetched bots struct.
type FetchedBots struct {
	Bots  []*Bot  `json:"bots"`
	Users []*User `json:"users"`
}

// Calculate creation date and edit the struct.
func (b *Bot) CalculateCreationDate() error {
	ulid, err := ulid.Parse(b.ID)

	if err != nil {
		return err
	}

	b.CreatedAt = time.UnixMilli(int64(ulid.Time()))
	return nil
}

// Edit the bot.
func (b *Bot) Edit(eb *EditBot) error {
	data, err := json.Marshal(eb)

	if err != nil {
		return err
	}

	_, err = b.Client.request(http.MethodPatch, "/bots/"+b.ID, data)
	return err
}

// Delete the bot.
func (b *Bot) Delete() error {
	_, err := b.Client.request("DELETE", "/bots/"+b.ID, nil)
	return err
}
