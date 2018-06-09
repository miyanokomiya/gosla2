package lib

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/google/go-github/github"
)

// HookContext Githubから受け取るJSONデータを格納する構造体
type HookContext struct {
	Signature string
	Event     string
	ID        string
	Payload   []byte
}

const key = "miyanokomiya"

var r = regexp.MustCompile(`@[a-zA-Z0-9_\-]+`)

// error定義まとめ
var (
	ErrNotSignature     = errors.New("no_signature")
	ErrNotEvent         = errors.New("no_event")
	ErrNotEventID       = errors.New("no_event_id")
	ErrInvalidSignature = errors.New("invalid_signature")
	ErrEmptyPayload     = errors.New("empty_payload")
	ErrUnhandledEvent   = errors.New("unhandled_event")
	ErrUnhandledAction  = errors.New("unhandled_action")
)

// EventSummary githubイベントサマリ
type EventSummary struct {
	RepositoryName string
	Title          string
	URL            string
	Description    string
	Comment        string
}

// ParseHook Githubのリクエストをパースする関数
func ParseHook(req *rest.Request) (HookContext, error) {
	hc := HookContext{}
	secret := []byte(key)
	if hc.Signature = req.Header.Get("x-hub-signature"); len(hc.Signature) == 0 {
		return hc, ErrNotSignature
	}
	if hc.Event = req.Header.Get("x-github-event"); len(hc.Event) == 0 {
		return hc, ErrNotEvent
	}
	if hc.ID = req.Header.Get("x-github-delivery"); len(hc.ID) == 0 {
		return hc, ErrNotEventID
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return hc, err
	}
	defer req.Body.Close()
	if !verifySignature(secret, hc.Signature, body) {
		return hc, ErrInvalidSignature
	}
	if len(body) == 0 {
		return hc, ErrEmptyPayload
	}
	hc.Payload = body
	return hc, nil
}

// FindAccounts コメントに含まれるメンションに対応するアカウント情報一覧を取得する
func FindAccounts(comment string, conf *Config) map[string]Account {
	accounts := map[string]Account{}
	matches := r.FindAllStringSubmatch(comment, -1)
	for _, val := range matches {
		hit := val[0]
		if account, ok := conf.Accounts[hit]; ok {
			accounts[hit] = account
		}
	}
	return accounts
}

// ReplaceComment コメント内のアカウント情報を置き換える関数
func ReplaceComment(comment string, accounts map[string]Account) string {
	replaced := comment
	matches := r.FindAllStringSubmatch(comment, -1)
	for _, val := range matches {
		if account, ok := accounts[val[0]]; ok {
			replaced = strings.Replace(replaced, val[0], "<"+account.ID+">", -1)
		}
	}
	return replaced
}

// secretキー検証
func verifySignature(secret []byte, signature string, body []byte) bool {

	const signaturePrefix = "sha1="
	const signatureLength = 45 // len(SignaturePrefix) + len(hex(sha1))

	if len(signature) != signatureLength || !strings.HasPrefix(signature, signaturePrefix) {
		return false
	}

	actual := make([]byte, 20)
	hex.Decode(actual, []byte(signature[len(signaturePrefix):]))

	return hmac.Equal(signBody(secret, body), actual)
}

func signBody(secret, body []byte) []byte {
	computed := hmac.New(sha1.New, secret)
	computed.Write(body)
	return []byte(computed.Sum(nil))
}

// parseIssuesEvent issuesイベントをパースする
func parseIssuesEvent(hc HookContext) (EventSummary, error) {
	summary := EventSummary{}
	evt := github.IssuesEvent{}
	err := json.Unmarshal(hc.Payload, &evt)
	if err != nil {
		return summary, err
	}
	if *evt.Action != "opened" && *evt.Action != "edited" {
		return summary, ErrUnhandledAction
	}
	summary.RepositoryName = *evt.Repo.Name
	summary.Title = *evt.Issue.Title
	summary.URL = *evt.Issue.HTMLURL
	summary.Description = fmt.Sprintf("Issue %v by: %v", *evt.Action, *evt.Issue.User.Login)
	summary.Comment = *evt.Issue.Body
	return summary, nil
}

// parseIssueCommentsEvent issue_commentイベントをパースする
func parseIssueCommentsEvent(hc HookContext) (EventSummary, error) {
	summary := EventSummary{}
	evt := github.IssueCommentEvent{}
	err := json.Unmarshal(hc.Payload, &evt)
	if err != nil {
		return summary, err
	}
	if *evt.Action != "created" && *evt.Action != "edited" {
		return summary, ErrUnhandledAction
	}
	summary.RepositoryName = *evt.Repo.Name
	summary.Title = *evt.Issue.Title
	summary.URL = *evt.Comment.HTMLURL
	summary.Description = fmt.Sprintf("Comment %v by: %v", *evt.Action, *evt.Comment.User.Login)
	summary.Comment = *evt.Comment.Body
	return summary, nil
}

// parsePullRequestEvent pull_requestイベントをパースする
func parsePullRequestEvent(hc HookContext) (EventSummary, error) {
	summary := EventSummary{}
	evt := github.PullRequestEvent{}
	err := json.Unmarshal(hc.Payload, &evt)
	if err != nil {
		return summary, err
	}
	if *evt.Action != "opened" && *evt.Action != "edited" {
		return summary, ErrUnhandledAction
	}
	summary.RepositoryName = *evt.Repo.Name
	summary.Title = *evt.PullRequest.Title
	summary.URL = *evt.PullRequest.HTMLURL
	summary.Description = fmt.Sprintf("Comment %v by: %v", *evt.Action, *evt.PullRequest.User.Login)
	summary.Comment = *evt.PullRequest.Body
	return summary, nil
}

// parsePullRequestReviewCommentEvent pull_request_review_commentイベントをパースする
func parsePullRequestReviewCommentEvent(hc HookContext) (EventSummary, error) {
	summary := EventSummary{}
	evt := github.PullRequestReviewCommentEvent{}
	err := json.Unmarshal(hc.Payload, &evt)
	if err != nil {
		return summary, err
	}
	if *evt.Action != "opened" && *evt.Action != "edited" {
		return summary, ErrUnhandledAction
	}
	summary.RepositoryName = *evt.Repo.Name
	summary.Title = *evt.PullRequest.Title
	summary.URL = *evt.Comment.HTMLURL
	summary.Description = fmt.Sprintf("Comment %v by: %v", *evt.Action, *evt.Comment.User.Login)
	summary.Comment = *evt.Comment.Body
	return summary, nil
}

// CreateEventSummary githubイベントサマリ生成
func CreateEventSummary(hc HookContext) (EventSummary, error) {
	switch hc.Event {
	case "issues":
		return parseIssuesEvent(hc)
	case "issue_comment":
		return parseIssueCommentsEvent(hc)
	case "pull_request":
		return parsePullRequestEvent(hc)
	case "pull_request_review_comment":
		return parsePullRequestReviewCommentEvent(hc)
	default:
		return EventSummary{}, ErrUnhandledEvent
	}
}
