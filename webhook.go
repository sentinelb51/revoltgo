package revoltgo

type Webhook struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Avatar    *Attachment `json:"avatar"`
	ChannelId string      `json:"channel_id"`
	Token     string      `json:"token"`
}

type WebhookCreate struct {
	Name   string
	Avatar string
}
