package revoltgo

import "log"

//go:generate msgp -tests=false -io=false

type WebhookRemoveField string

// todo: why aren't we using this?

const (
	WebhookRemoveNickname WebhookRemoveField = "Nickname"
	WebhookRemoveAvatar   WebhookRemoveField = "Avatar"
)

// Webhook is derived from
// https://github.com/stoatchat/stoatchat/blob/main/crates/core/database/src/models/channel_webhooks/model.rs#L8
type Webhook struct {
	ID          string  `msg:"_id" json:"_id,omitzero"`
	Name        string  `msg:"name" json:"name,omitzero"`
	Avatar      *File   `msg:"avatar" json:"avatar,omitzero"`
	CreatorID   string  `msg:"creator_id" json:"creator_id,omitzero"`
	ChannelID   string  `msg:"channel_id" json:"channel_id,omitzero"`
	Permissions int64   `msg:"permissions" json:"permissions,omitzero"`
	Token       *string `msg:"token" json:"token,omitzero"`
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
		switch WebhookRemoveField(field) {
		case WebhookRemoveAvatar:
			w.Avatar = nil
		default:
			log.Printf("Webhook.clear(): unknown field %s", field)
		}
	}
}

type PartialWebhook struct {
	Name        *string `msg:"name" json:"name,omitzero"`
	Avatar      *File   `msg:"avatar" json:"avatar,omitzero"`
	CreatorID   *string `msg:"creator_id" json:"creator_id,omitzero"`
	ChannelID   *string `msg:"channel_id" json:"channel_id,omitzero"`
	Permissions *int64  `msg:"permissions" json:"permissions,omitzero"`
	Token       *string `msg:"token" json:"token,omitzero"`
}
