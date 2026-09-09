/*
Package revoltgo is a wrapper for the Revolt API with low-level bindings

		Made by @sentinelb51
		For support, join our revolt server on the GitHub README file
		To compile correctly, always run beforehand:
			/tools/msgp_codegen.py  (ensures all msgp code is generated: revoltgo_msgp_gen.go)
*/

package revoltgo

//go:generate msgp -tests=false -io=false

import (
	json "encoding/json/v2"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// VERSION is the module version recorded in the importing build; "(devel)" inside this repo.
var VERSION = func() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "(unknown)"
	}

	for _, d := range bi.Deps {
		if d.Path == "github.com/sentinelb51/revoltgo" {
			return d.Version
		}
	}

	return "(dev)"
}()

const (
	ExpectedAPI    = "0.15.1"
	MainCommitsURL = "https://api.github.com/repos/sentinelb51/revoltgo/commits/main"
)

/* Logic related to the update checker */

// commit is the short hash trailing this module's pseudo-version. Empty when the
// importing build resolved no version, e.g. inside this repo or behind a replace.
var commit = func() string {
	short := VERSION[strings.LastIndex(VERSION, "-")+1:]
	if len(short) != 12 {
		return ""
	}

	return short
}()

type GithubRepos struct {
	Sha     string            `json:"sha"`
	Commits GithubReposCommit `json:"commit"`
}

type GithubReposCommit struct {
	Author    GithubReposCommitUserData `json:"author"`
	Committer GithubReposCommitUserData `json:"committer"`
	Message   string                    `json:"message"`
}

type GithubReposCommitUserData struct {
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

func HasUpdate() bool {
	if commit == "" {
		logf("Update check skipped: this build has no resolvable module version")
		return false
	}

	response, err := http.Get(MainCommitsURL)
	if err != nil {
		logf("Update check failed whilst fetching: %v", err)
		return false
	}

	defer response.Body.Close()

	var repo GithubRepos
	err = json.UnmarshalRead(response.Body, &repo)
	if err != nil {
		logf("Update check failed whilst decoding: %v", err)
		return false
	}

	if !strings.HasPrefix(repo.Sha, commit) {
		days := time.Now().Sub(repo.Commits.Author.Date).Hours() / 24
		logf("A new update is available (%.0f days ago)", days)
		logf("To update, run: go get -u github.com/sentinelb51/revoltgo")
		return true
	}

	logf("Update check complete; you are using the latest version of revoltgo")
	return false
}

/* Data structures for instance configuration, retrieved when you first contact apiURL */

type InstanceConfig struct {
	WS       string                 `msg:"ws" json:"ws,omitzero"`
	App      string                 `msg:"app" json:"app,omitzero"`
	VapID    string                 `msg:"vapid" json:"vapid,omitzero"`
	Revolt   string                 `msg:"revolt" json:"revolt,omitzero"`
	Build    InstanceConfigBuild    `msg:"build" json:"build,omitzero"`
	Features InstanceConfigFeatures `msg:"features" json:"features,omitzero"`
}

type InstanceConfigFeaturesCaptcha struct {
	Enabled bool   `msg:"enabled" json:"enabled,omitzero"`
	Key     string `msg:"key" json:"key,omitzero"`
}

type InstanceConfigFeaturesAutumn struct {
	Enabled bool   `msg:"enabled" json:"enabled,omitzero"`
	URL     string `msg:"url" json:"url,omitzero"`
}

type InstanceConfigFeaturesJanuary struct {
	Enabled bool   `msg:"enabled" json:"enabled,omitzero"`
	URL     string `msg:"url" json:"url,omitzero"`
}

// InstanceConfigVoiceNode is one voice server the instance publishes. Name is
// what ChannelJoinCallParams.Node has to be given; the coordinates are there so
// a client that knows its own location can pick the nearest.
type InstanceConfigVoiceNode struct {
	Name      string  `msg:"name" json:"name,omitzero"`
	Latitude  float64 `msg:"lat" json:"lat,omitzero"`
	Longitude float64 `msg:"lon" json:"lon,omitzero"`
	PublicURL string  `msg:"public_url" json:"public_url,omitzero"`
}

// InstanceConfigFeaturesLiveKit is the instance's voice configuration. It
// replaces the "voso" block, which the backend no longer sends.
type InstanceConfigFeaturesLiveKit struct {
	Enabled bool                      `msg:"enabled" json:"enabled,omitzero"`
	Nodes   []InstanceConfigVoiceNode `msg:"nodes" json:"nodes,omitzero"`
}

type InstanceConfigFeaturesLimitsGlobal struct {
	GroupSize              int64    `msg:"group_size" json:"group_size,omitzero"`
	MessageEmbeds          int64    `msg:"message_embeds" json:"message_embeds,omitzero"`
	MessageReplies         int64    `msg:"message_replies" json:"message_replies,omitzero"`
	MessageReactions       int64    `msg:"message_reactions" json:"message_reactions,omitzero"`
	ServerEmoji            int64    `msg:"server_emoji" json:"server_emoji,omitzero"`
	ServerRoles            int64    `msg:"server_roles" json:"server_roles,omitzero"`
	ServerChannels         int64    `msg:"server_channels" json:"server_channels,omitzero"`
	BodyLimitSize          int64    `msg:"body_limit_size" json:"body_limit_size,omitzero"`
	RestrictServerCreation []string `msg:"restrict_server_creation" json:"restrict_server_creation,omitzero"`
	NewUserHours           int64    `msg:"new_user_hours" json:"new_user_hours,omitzero"`
}

type InstanceConfigFeaturesLimitsUser struct {
	OutgoingFriendRequests int64           `msg:"outgoing_friend_requests" json:"outgoing_friend_requests,omitzero"`
	Bots                   int64           `msg:"bots" json:"bots,omitzero"`
	MessageLength          int64           `msg:"message_length" json:"message_length,omitzero"`
	MessageAttachments     int64           `msg:"message_attachments" json:"message_attachments,omitzero"`
	Servers                int64           `msg:"servers" json:"servers,omitzero"`
	VoiceQuality           int64           `msg:"voice_quality" json:"voice_quality,omitzero"`
	Video                  bool            `msg:"video" json:"video,omitzero"`
	VideoResolution        []int64         `msg:"video_resolution" json:"video_resolution,omitzero"`
	VideoAspectRatio       []float64       `msg:"video_aspect_ratio" json:"video_aspect_ratio,omitzero"`
	FileUploadSizeLimits   map[string]uint `msg:"file_upload_size_limits" json:"file_upload_size_limits,omitzero"`
}

type InstanceConfigFeaturesLimits struct {
	Global  InstanceConfigFeaturesLimitsGlobal `msg:"global" json:"global,omitzero"`
	NewUser InstanceConfigFeaturesLimitsUser   `msg:"new_user" json:"new_user,omitzero"`
	Default InstanceConfigFeaturesLimitsUser   `msg:"default" json:"default,omitzero"`
}

type InstanceConfigFeaturesLegalLinks struct {
	TermsOfService string `msg:"terms_of_service" json:"terms_of_service,omitzero"`
	PrivacyPolicy  string `msg:"privacy_policy" json:"privacy_policy,omitzero"`
	Guidelines     string `msg:"guidelines" json:"guidelines,omitzero"`
}

type InstanceConfigFeatures struct {
	Captcha    InstanceConfigFeaturesCaptcha `msg:"captcha" json:"captcha,omitzero"`
	Email      bool                          `msg:"email" json:"email,omitzero"`
	InviteOnly bool                          `msg:"invite_only" json:"invite_only,omitzero"`
	Autumn     InstanceConfigFeaturesAutumn  `msg:"autumn" json:"autumn,omitzero"`
	January    InstanceConfigFeaturesJanuary `msg:"january" json:"january,omitzero"`
	LiveKit    InstanceConfigFeaturesLiveKit `msg:"livekit" json:"livekit,omitzero"`

	Limits     InstanceConfigFeaturesLimits     `msg:"limits" json:"limits,omitzero"`
	LegalLinks InstanceConfigFeaturesLegalLinks `msg:"legal_links" json:"legal_links,omitzero"`
}

type InstanceConfigBuild struct {
	CommitSha       string `msg:"commit_sha" json:"commit_sha,omitzero"`
	CommitTimestamp string `msg:"commit_timestamp" json:"commit_timestamp,omitzero"`
	SemVer          string `msg:"semver" json:"semver,omitzero"`
	OriginURL       string `msg:"origin_url" json:"origin_url,omitzero"`
	Timestamp       string `msg:"timestamp" json:"timestamp,omitzero"`
}
