package revoltgo

import "log"

//go:generate msgp -tests=false -io=false

// Webhook is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/database/src/models/channel_webhooks/model.rs#L8
type Webhook struct {
	ID          string      `msg:"_id" json:"_id,omitempty"` // Rust: rename = "_id"
	Name        string      `msg:"name" json:"name,omitempty"`
	Avatar      *Attachment `msg:"avatar" json:"avatar,omitempty"`
	CreatorID   string      `msg:"creator_id" json:"creator_id,omitempty"`
	ChannelID   string      `msg:"channel_id" json:"channel_id,omitempty"`
	Permissions uint64      `msg:"permissions" json:"permissions,omitempty"`
	Token       *string     `msg:"token" json:"token,omitempty"`
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
	Name        *string     `msg:"name" json:"name,omitempty"`
	Avatar      *Attachment `msg:"avatar" json:"avatar,omitempty"`
	CreatorID   *string     `msg:"creator_id" json:"creator_id,omitempty"`
	ChannelID   *string     `msg:"channel_id" json:"channel_id,omitempty"`
	Permissions *uint64     `msg:"permissions" json:"permissions,omitempty"`
	Token       *string     `msg:"token" json:"token,omitempty"`
}

type WebhookCreate struct {
	Name   string `msg:"name" json:"name,omitempty"`
	Avatar string `msg:"avatar" json:"avatar,omitempty"`
}
