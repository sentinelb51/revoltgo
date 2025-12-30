/*
Package revoltgo is a Go wrapper for the Revolt API with low-level bindings

	Made by @sentinelb51
	For support, join our revolt server on the GitHub README file

Sometimes I will leave TO-DO comments for things I want to globally change later.
*/
package revoltgo

const VERSION = "v3.0.0-beta.2"

type RootData struct {
	WS       string           `json:"ws"`
	App      string           `json:"app"`
	VapID    string           `json:"vapid"`
	Revolt   string           `json:"revolt"`
	Build    RootDataBuild    `json:"build"`
	Features RootDataFeatures `json:"features"`
}

type RootDataFeaturesCaptcha struct {
	Enabled bool   `json:"enabled"`
	Key     string `json:"key"`
}

type RootDataFeaturesAutumn struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
}

type RootDataFeaturesJanuary struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
}

type RootDataFeaturesVoso struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
	WS      string `json:"ws"`
}

type RootDataFeatures struct {
	Captcha    RootDataFeaturesCaptcha `json:"captcha"`
	Email      bool                    `json:"email"`
	InviteOnly bool                    `json:"invite_only"`
	Autumn     RootDataFeaturesAutumn  `json:"autumn"`
	January    RootDataFeaturesJanuary `json:"january"`
	Voso       RootDataFeaturesVoso    `json:"voso"`
}

type RootDataBuild struct {
	CommitSha       string `json:"commit_sha"`
	CommitTimestamp string `json:"commit_timestamp"`
	SemVer          string `json:"semver"`
	OriginURL       string `json:"origin_url"`
	Timestamp       string `json:"timestamp"`
}
