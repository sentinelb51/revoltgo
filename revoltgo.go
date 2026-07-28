/*
Package revoltgo is a wrapper for the Revolt API with low-level bindings

		Made by @sentinelb51
		For support, join our revolt server on the GitHub README file
		To compile correctly, always run beforehand:
			/tools/msgp_codegen.py  (ensures all msgp code is generated: revoltgo_msgp_gen.go)
			/tools/build_hash.py    (updates the COMMIT variable in this file)

	   Todo: do we need state.go to track VoiceStates?
*/

/*
	TODO:

Full pass over the library hunting for critical bugs, logic errors, and dead weight. Hot paths (decode-skip dispatch, COW handlers, zstd pooling, ratelimiter) are already lean — no major perf rework needed. The real value i↑ a set of genuine bugs, several of which panic or silently break features.

	                                                                                                                                                                                                                                 ↑
	  Critical bugs
	                                                                                                                                                                                                                                 ↑
	  1. ServerMembers panics on request error — session.go:892: s.State.addServerMembersAndUsers(data.Users, data.Members) runs even when err != nil, and data (a *ServerMembers) is nil → nil-pointer panic on any timeout/error. Guard with if err == nil.                                                                                                                                                                                                      ↑
	  2. deleteChannel panics for non-server channels — state.go:892: *channel.Server derefs nil when a DM/Group/SavedMessages channel is deleted. Nil-check channel.Server before the server-cleanup half.
	  3. ServerPermissions has swapped args — permissions.go:75: s.Member(user.ID, server.ID) but the signature is Member(sID, uID). Member lookup always fails → permission calc errors for every non-owner. Swap to s.Member(server.ID, user.ID).
	  4. ChannelPermissions DM nil deref — permissions.go:107: *channel.Permissions without nil check (Group case below checks; DM doesn't). Nil-check like the Group branch.                                                        ↑
	  5. UserBlock/UserUnblock decode into nil pointer — session.go:564,570: pass user (nil *User) instead of &user. Result is never populated and a 200-with-body returns a decode error. Use &user.
	  6. MessageFlagsMentionsOnline = 3 — message.go:49: bitflag value should be 4; 3 equals SuppressNotifications|MentionsEveryone (the comment even says they're mutually exclusive). Change to 4.                                 ↑
	  7. Relationships requests a literal %s path — session.go:1143: uses const URLUserRelationships = "/users/%s/relationships" directly with no formatting → garbage URL. Correct API path is /users/relationships(verify); replace with a literal/Endpoint func.                                                                                                                                                                                                  ↑
	  8. AddHandler for EventMessageAppend/EventMessageRemoveReaction kills the process — events.go:179,277: these two structs don't embed Event, so they have no Type field and handlerName (session.go:303) hits log.Fatalf. Embed Event in both (msgp regen needed — user runs codegen).                                                                                                                                                                         ↑
	                                                                                                                                                                                                                                 ↑
	  Medium bugs

	  9. connect() ignores ShouldReconnect — websocket.go:128-133: dial failure unconditionally spawns reconnectLoop, even when reconnects are disabled. Also Open() returns nil for a failed first connect. Gate the reconnect on ws.ShouldReconnect; optionally have connect() return the error so Open can report the initial failure.
	  10. log.Fatalf in OnOpen — websocket.go:194: a SetDeadline error kills the host process. Log + close the socket instead (OnClose path handles reconnect).
	  11. Sync settings endpoints wrong — session.go:1216-1225: both SyncSettingsFetch and SyncSettingsSet POST /sync/settings; correct paths are /sync/settings/fetch and /sync/settings/set. The unused EndpointSyncSettings (endpoints.go:357) already exists for this — use it.
	  12. UserDefaultAvatar can't work — session.go:580: JSON-decodes a binary PNG body into []byte → always errors. Needs a raw-body path (e.g. special-case *[]byte in handleResponse to io.ReadAll).
	  13. Ratelimiter cleaner goroutine leaks — Session.Close/WriteClose never call Ratelimiter.Close(); one goroutine leaks per session forever. Call it from Session.Close.
	  14. createServer nil derefs — state.go:920: s.self.Load().ID panics if self is unknown (Ready had no users and the @me fetch failed); event.Server also unchecked before event.Server.ID. Add nil guards.

	  Cleanup / simplification

	15. Delete the dead URL* const block — endpoints.go:73-141 (~70 lines): everything except URLUserMeUsername is unused or broken (URLUserRelationships is bug #7). Keep those two as literals or Endpoint funcs.
	  16. Delete eventTypeFromJSON + jsonSkipAheadKeyType — event.go:13-26: dead since the msgpack switch.
	  17. handlerName: reject non-pointer T — registering func(*Session, EventMessage) (value, not pointer) currently registers fine, then panics at dispatch on e.(T) since constructors produce pointers. Make handlerName fatal at registration if T isn't a pointer-to-struct (it already drills pointers; just require Kind == Ptr at top instead).

	  Files touched

	  session.go, state.go, permissions.go, message.go, websocket.go, events.go, event.go, endpoints.go, http.go (item 12).

	  Verification

	  - Item 8 requires msgp regeneration (tools/msgp_codegen.py) — user runs this.
	- User verifies compilation per their workflow (no go build/vet from me).
	  - Behavioral spot-checks: ServerMembers against a failing request must return error, not panic; ServerPermissions for a non-owner member must resolve; DM channel delete event must not panic.
*/
package revoltgo

import (
	"log"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

const (
	VERSION        = "v3.0.0"
	MainCommitsURL = "https://api.github.com/repos/sentinelb51/revoltgo/commits/main"
)

/* Logic related to the update checker */

var COMMIT = "eeea8c05edbeba726ee54c09353a8cdda781a519"

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
