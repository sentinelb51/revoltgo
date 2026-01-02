/*
Package revoltgo is a Go wrapper for the Revolt API with low-level bindings

	Made by @sentinelb51
	For support, join our revolt server on the GitHub README file

Sometimes I will leave TO-DO comments for things I want to globally change later.
*/
package revoltgo

const VERSION = "v3.0.0-beta.4"

type RootData struct {
	WS       string           `msg:"ws" json:"ws,omitempty"`
	App      string           `msg:"app" json:"app,omitempty"`
	VapID    string           `msg:"vapid" json:"vapid,omitempty"`
	Revolt   string           `msg:"revolt" json:"revolt,omitempty"`
	Build    RootDataBuild    `msg:"build" json:"build,omitempty"`
	Features RootDataFeatures `msg:"features" json:"features,omitempty"`
}

type RootDataFeaturesCaptcha struct {
	Enabled bool   `msg:"enabled" json:"enabled,omitempty"`
	Key     string `msg:"key" json:"key,omitempty"`
}

type RootDataFeaturesAutumn struct {
	Enabled bool   `msg:"enabled" json:"enabled,omitempty"`
	URL     string `msg:"url" json:"url,omitempty"`
}

type RootDataFeaturesJanuary struct {
	Enabled bool   `msg:"enabled" json:"enabled,omitempty"`
	URL     string `msg:"url" json:"url,omitempty"`
}

type RootDataFeaturesVoso struct {
	Enabled bool   `msg:"enabled" json:"enabled,omitempty"`
	URL     string `msg:"url" json:"url,omitempty"`
	WS      string `msg:"ws" json:"ws,omitempty"`
}

type RootDataFeatures struct {
	Captcha    RootDataFeaturesCaptcha `msg:"captcha" json:"captcha,omitempty"`
	Email      bool                    `msg:"email" json:"email,omitempty"`
	InviteOnly bool                    `msg:"invite_only" json:"invite_only,omitempty"`
	Autumn     RootDataFeaturesAutumn  `msg:"autumn" json:"autumn,omitempty"`
	January    RootDataFeaturesJanuary `msg:"january" json:"january,omitempty"`
	Voso       RootDataFeaturesVoso    `msg:"voso" json:"voso,omitempty"`
}

type RootDataBuild struct {
	CommitSha       string `msg:"commit_sha" json:"commit_sha,omitempty"`
	CommitTimestamp string `msg:"commit_timestamp" json:"commit_timestamp,omitempty"`
	SemVer          string `msg:"semver" json:"semver,omitempty"`
	OriginURL       string `msg:"origin_url" json:"origin_url,omitempty"`
	Timestamp       string `msg:"timestamp" json:"timestamp,omitempty"`
}
