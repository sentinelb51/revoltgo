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
	VERSION        = "v3.0.0-beta.11"
	MainCommitsURL = "https://api.github.com/repos/sentinelb51/revoltgo/commits/main"
)

var COMMIT = "23abc181da6c510a1320c6952d1929d5b6cf2e5c"

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
