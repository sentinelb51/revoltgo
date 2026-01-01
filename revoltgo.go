/*
Package revoltgo is a Go wrapper for the Revolt API with low-level bindings

	Made by @sentinelb51
	For support, join our revolt server on the GitHub README file

Sometimes I will leave TO-DO comments for things I want to globally change later.
*/
package revoltgo

const VERSION = "v3.0.0-beta.2"

type RootData struct {
	WS       string           `msg:"ws"`
	App      string           `msg:"app"`
	VapID    string           `msg:"vapid"`
	Revolt   string           `msg:"revolt"`
	Build    RootDataBuild    `msg:"build"`
	Features RootDataFeatures `msg:"features"`
}

type RootDataFeaturesCaptcha struct {
	Enabled bool   `msg:"enabled"`
	Key     string `msg:"key"`
}

type RootDataFeaturesAutumn struct {
	Enabled bool   `msg:"enabled"`
	URL     string `msg:"url"`
}

type RootDataFeaturesJanuary struct {
	Enabled bool   `msg:"enabled"`
	URL     string `msg:"url"`
}

type RootDataFeaturesVoso struct {
	Enabled bool   `msg:"enabled"`
	URL     string `msg:"url"`
	WS      string `msg:"ws"`
}

type RootDataFeatures struct {
	Captcha    RootDataFeaturesCaptcha `msg:"captcha"`
	Email      bool                    `msg:"email"`
	InviteOnly bool                    `msg:"invite_only"`
	Autumn     RootDataFeaturesAutumn  `msg:"autumn"`
	January    RootDataFeaturesJanuary `msg:"january"`
	Voso       RootDataFeaturesVoso    `msg:"voso"`
}

type RootDataBuild struct {
	CommitSha       string `msg:"commit_sha"`
	CommitTimestamp string `msg:"commit_timestamp"`
	SemVer          string `msg:"semver"`
	OriginURL       string `msg:"origin_url"`
	Timestamp       string `msg:"timestamp"`
}
