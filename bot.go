package revoltgo

import (
	"encoding/json"
	"net/http"
)

// Bot struct.
type Bot struct {
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

// Edit the bot.
func (b *Bot) Edit(session *Session, eb *EditBot) error {
	data, err := json.Marshal(eb)

	if err != nil {
		return err
	}

	_, err = session.handleRequest(http.MethodPatch, "/bots/"+b.ID, data)
	return err
}

// Delete the bot.
func (b *Bot) Delete(session *Session) error {
	_, err := session.handleRequest(http.MethodDelete, "/bots/"+b.ID, nil)
	return err
}
