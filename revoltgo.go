/*
Package revoltgo is a wrapper for the Revolt API with low-level bindings

		Made by @sentinelb51
		For support, join our revolt server on the GitHub README file
		To compile correctly, always run beforehand:
			/tools/msgp_codegen.py  (ensures all msgp code is generated: revoltgo_msgp_gen.go)

	   Todo: do we need state.go to track VoiceStates?
*/

package revoltgo

import (
	json "encoding/json/v2"
	"log"
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
		log.Printf("Update check skipped: this build has no resolvable module version")
		return false
	}

	response, err := http.Get(MainCommitsURL)
	if err != nil {
		log.Printf("Update check failed whilst fetching: %v", err)
		return false
	}

	defer response.Body.Close()

	var repo GithubRepos
	err = json.UnmarshalRead(response.Body, &repo)
	if err != nil {
		log.Printf("Update check failed whilst decoding: %v", err)
		return false
	}

	if !strings.HasPrefix(repo.Sha, commit) {
		days := time.Now().Sub(repo.Commits.Author.Date).Hours() / 24
		log.Printf("A new update is available (%.0f days ago)", days)
		log.Printf("To update, run: go get -u github.com/sentinelb51/revoltgo")
		return true
	}

	log.Printf("Update check complete; you are using the latest version of revoltgo")
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

type InstanceConfigFeaturesVoso struct {
	Enabled bool   `msg:"enabled" json:"enabled,omitzero"`
	URL     string `msg:"url" json:"url,omitzero"`
	WS      string `msg:"ws" json:"ws,omitzero"`
}

type InstanceConfigFeatures struct {
	Captcha    InstanceConfigFeaturesCaptcha `msg:"captcha" json:"captcha,omitzero"`
	Email      bool                          `msg:"email" json:"email,omitzero"`
	InviteOnly bool                          `msg:"invite_only" json:"invite_only,omitzero"`
	Autumn     InstanceConfigFeaturesAutumn  `msg:"autumn" json:"autumn,omitzero"`
	January    InstanceConfigFeaturesJanuary `msg:"january" json:"january,omitzero"`
	Voso       InstanceConfigFeaturesVoso    `msg:"voso" json:"voso,omitzero"`
}

type InstanceConfigBuild struct {
	CommitSha       string `msg:"commit_sha" json:"commit_sha,omitzero"`
	CommitTimestamp string `msg:"commit_timestamp" json:"commit_timestamp,omitzero"`
	SemVer          string `msg:"semver" json:"semver,omitzero"`
	OriginURL       string `msg:"origin_url" json:"origin_url,omitzero"`
	Timestamp       string `msg:"timestamp" json:"timestamp,omitzero"`
}
