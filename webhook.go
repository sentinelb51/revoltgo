package revoltgo

import "log"

// Webhook is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/database/src/models/channel_webhooks/model.rs#L8
type Webhook struct {
	ID          string      `json:"_id"` // Rust: rename = "_id"
	Name        string      `json:"name"`
	Avatar      *Attachment `json:"avatar"`
	CreatorID   string      `json:"creator_id"`
	ChannelID   string      `json:"channel_id"`
	Permissions uint64      `json:"permissions"`
	Token       *string     `json:"token"`
}

func (w *Webhook) update(data PartialWebhook) {
	if data.Name != nil {
		w.Name = *data.Name
	}

	if data.Avatar != nil {
		w.Avatar = data.Avatar
	}

	if data.CreatorID != nil {
		w.CreatorID = *data.CreatorID
	}

	if data.ChannelID != nil {
		w.ChannelID = *data.ChannelID
	}

	if data.Permissions != nil {
		w.Permissions = *data.Permissions
	}

	if data.Token != nil {
		w.Token = data.Token
	}
}

func (w *Webhook) clear(fields []string) {
	for _, field := range fields {
		switch field {
		case "Avatar":
			w.Avatar = nil
		default:
			log.Printf("Webhook.clear(): unknown field %s", field)
		}
	}
}

type PartialWebhook struct {
	Name        *string     `json:"name,omitempty"`
	Avatar      *Attachment `json:"avatar,omitempty"`
	CreatorID   *string     `json:"creator_id,omitempty"`
	ChannelID   *string     `json:"channel_id,omitempty"`
	Permissions *uint64     `json:"permissions,omitempty"`
	Token       *string     `json:"token,omitempty"`
}

type WebhookCreate struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
}
