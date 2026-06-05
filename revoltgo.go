/*
Package revoltgo is a wrapper for the Revolt API with low-level bindings

	Made by @sentinelb51
	For support, join our revolt server on the GitHub README file
	To compile correctly, always run beforehand:
		/tools/msgp_codegen.py  (ensures all msgp code is generated: revoltgo_msgp_gen.go)
		/tools/build_hash.py         (updates the COMMIT variable in this file)
*/
package revoltgo

import (
	"log"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

const (
	VERSION        = "v3.0.0-beta.13"
	MainCommitsURL = "https://api.github.com/repos/sentinelb51/revoltgo/commits/main"
)

/* Logic related to the update checker */

var COMMIT = "0c230f218fe81f9c3c488a0e409b81e7c9818491"

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
	response, err := http.Get(MainCommitsURL)
	if err != nil {
		log.Printf("Update check failed whilst fetching: %v", err)
		return false
	}

	defer response.Body.Close()

	var repo GithubRepos
	err = json.NewDecoder(response.Body).Decode(&repo)
	if err != nil {
		log.Printf("Update check failed whilst decoding: %v", err)
		return false
	}

	if repo.Sha != COMMIT {
		days := time.Now().Sub(repo.Commits.Author.Date).Hours() / 24
		log.Printf("A new nightly update is available (%.0f days ago)", days)
		log.Printf("To update, run: go get -u github.com/sentinelb51/revoltgo")
		return true
	}

	log.Printf("Update check complete; you are using the latest version of revoltgo")
	return false
}

/* Data structures for instance configuration, retrieved when you first contact apiURL */

type InstanceConfig struct {
	WS       string                 `msg:"ws" json:"ws,omitempty"`
	App      string                 `msg:"app" json:"app,omitempty"`
	VapID    string                 `msg:"vapid" json:"vapid,omitempty"`
	Revolt   string                 `msg:"revolt" json:"revolt,omitempty"`
	Build    InstanceConfigBuild    `msg:"build" json:"build,omitempty"`
	Features InstanceConfigFeatures `msg:"features" json:"features,omitempty"`
}

type InstanceConfigFeaturesCaptcha struct {
	Enabled bool   `msg:"enabled" json:"enabled,omitempty"`
	Key     string `msg:"key" json:"key,omitempty"`
}

type InstanceConfigFeaturesAutumn struct {
	Enabled bool   `msg:"enabled" json:"enabled,omitempty"`
	URL     string `msg:"url" json:"url,omitempty"`
}

type InstanceConfigFeaturesJanuary struct {
	Enabled bool   `msg:"enabled" json:"enabled,omitempty"`
	URL     string `msg:"url" json:"url,omitempty"`
}

type InstanceConfigFeaturesVoso struct {
	Enabled bool   `msg:"enabled" json:"enabled,omitempty"`
	URL     string `msg:"url" json:"url,omitempty"`
	WS      string `msg:"ws" json:"ws,omitempty"`
}

type InstanceConfigFeatures struct {
	Captcha    InstanceConfigFeaturesCaptcha `msg:"captcha" json:"captcha,omitempty"`
	Email      bool                          `msg:"email" json:"email,omitempty"`
	InviteOnly bool                          `msg:"invite_only" json:"invite_only,omitempty"`
	Autumn     InstanceConfigFeaturesAutumn  `msg:"autumn" json:"autumn,omitempty"`
	January    InstanceConfigFeaturesJanuary `msg:"january" json:"january,omitempty"`
	Voso       InstanceConfigFeaturesVoso    `msg:"voso" json:"voso,omitempty"`
}

type InstanceConfigBuild struct {
	CommitSha       string `msg:"commit_sha" json:"commit_sha,omitempty"`
	CommitTimestamp string `msg:"commit_timestamp" json:"commit_timestamp,omitempty"`
	SemVer          string `msg:"semver" json:"semver,omitempty"`
	OriginURL       string `msg:"origin_url" json:"origin_url,omitempty"`
	Timestamp       string `msg:"timestamp" json:"timestamp,omitempty"`
}
