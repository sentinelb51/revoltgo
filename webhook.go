package revoltgo

//go:generate msgp -tests=false -io=false

type WebhookRemoveField string

// The spec's FieldsWebhook enum admits only "Avatar"; WebhookRemoveNickname is
// carried for compatibility and the backend rejects it.
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

type PartialWebhook struct {
	Name        *string `msg:"name" json:"name,omitzero"`
	Avatar      *File   `msg:"avatar" json:"avatar,omitzero"`
	CreatorID   *string `msg:"creator_id" json:"creator_id,omitzero"`
	ChannelID   *string `msg:"channel_id" json:"channel_id,omitzero"`
	Permissions *int64  `msg:"permissions" json:"permissions,omitzero"`
	Token       *string `msg:"token" json:"token,omitzero"`
}
