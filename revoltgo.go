/*
Package revoltgo is a Go wrapper for the Revolt API with low-level bindings

	Made by @sentinelb51
	For support, join our revolt server on the GitHub README file

Sometimes I will leave TO-DO comments for things I want to globally change later.
*/
package revoltgo

const VERSION = "v3.0.0-beta.2"

type RootData struct {
	Revolt   string `json:"revolt"`
	Features struct {
		Captcha struct {
			Enabled bool   `json:"enabled"`
			Key     string `json:"key"`
		} `json:"captcha"`
		Email      bool `json:"email"`
		InviteOnly bool `json:"invite_only"`
		Autumn     struct {
			Enabled bool   `json:"enabled"`
			URL     string `json:"url"`
		} `json:"autumn"`
		January struct {
			Enabled bool   `json:"enabled"`
			URL     string `json:"url"`
		} `json:"january"`
		Voso struct {
			Enabled bool   `json:"enabled"`
			URL     string `json:"url"`
			WS      string `json:"ws"`
		} `json:"voso"`
	} `json:"features"`
	WS    string `json:"ws"`
	App   string `json:"app"`
	VapID string `json:"vapid"`
	Build struct {
		CommitSha       string `json:"commit_sha"`
		CommitTimestamp string `json:"commit_timestamp"`
		SemVer          string `json:"semver"`
		OriginURL       string `json:"origin_url"`
		Timestamp       string `json:"timestamp"`
	} `json:"build"`
}
