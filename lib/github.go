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

var mentionReg = regexp.MustCompile(`@[a-zA-Z0-9_\-]+`)

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
func (hc *HookContext) ParseHook(req *rest.Request, secret string) error {
	if hc.Signature = req.Header.Get("x-hub-signature"); len(hc.Signature) == 0 {
		return ErrNotSignature
	}
	if hc.Event = req.Header.Get("x-github-event"); len(hc.Event) == 0 {
		return ErrNotEvent
	}
	if hc.ID = req.Header.Get("x-github-delivery"); len(hc.ID) == 0 {
		return ErrNotEventID
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	if !verifySignature([]byte(secret), hc.Signature, body) {
		return ErrInvalidSignature
	}
	if len(body) == 0 {
		return ErrEmptyPayload
	}
	hc.Payload = body
	return nil
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

// FindAccounts コメントに含まれるメンションに対応するアカウント情報一覧を取得する
func FindAccounts(comment string, conf Config) map[string]Account {
	accounts := map[string]Account{}
	matches := mentionReg.FindAllStringSubmatch(comment, -1)
	for _, val := range matches {
		hit := val[0]
		if account, ok := conf.Accounts[hit]; ok {
			accounts[hit] = account
		}
	}
	return accounts
}

// ReplaceComment コメント内のアカウント情報を置き換える関数
func (summary *EventSummary) ReplaceComment(accounts map[string]Account) {
	for key, account := range accounts {
		summary.Comment = strings.Replace(summary.Comment, key, "<"+account.ID+">", -1)
	}
}

// parseIssuesEvent issuesイベントをパースする
func (summary *EventSummary) parseIssuesEvent(payload []byte) error {
	evt := github.IssuesEvent{}
	err := json.Unmarshal(payload, &evt)
	if err != nil {
		return err
	}
	if *evt.Action != "opened" && *evt.Action != "edited" {
		return ErrUnhandledAction
	}
	summary.RepositoryName = *evt.Repo.Name
	summary.Title = *evt.Issue.Title
	summary.URL = *evt.Issue.HTMLURL
	summary.Description = fmt.Sprintf("Issue %v by: %v", *evt.Action, *evt.Issue.User.Login)
	summary.Comment = *evt.Issue.Body
	return nil
}

// parseIssueCommentsEvent issue_commentイベントをパースする
func (summary *EventSummary) parseIssueCommentsEvent(payload []byte) error {
	evt := github.IssueCommentEvent{}
	err := json.Unmarshal(payload, &evt)
	if err != nil {
		return err
	}
	if *evt.Action != "created" && *evt.Action != "edited" {
		return ErrUnhandledAction
	}
	summary.RepositoryName = *evt.Repo.Name
	summary.Title = *evt.Issue.Title
	summary.URL = *evt.Comment.HTMLURL
	summary.Description = fmt.Sprintf("Comment %v by: %v", *evt.Action, *evt.Comment.User.Login)
	summary.Comment = *evt.Comment.Body
	return nil
}

// parsePullRequestEvent pull_requestイベントをパースする
func (summary *EventSummary) parsePullRequestEvent(payload []byte) error {
	evt := github.PullRequestEvent{}
	err := json.Unmarshal(payload, &evt)
	if err != nil {
		return err
	}
	if *evt.Action != "opened" && *evt.Action != "edited" {
		return ErrUnhandledAction
	}
	summary.RepositoryName = *evt.Repo.Name
	summary.Title = *evt.PullRequest.Title
	summary.URL = *evt.PullRequest.HTMLURL
	summary.Description = fmt.Sprintf("PullRequest %v by: %v", *evt.Action, *evt.PullRequest.User.Login)
	summary.Comment = *evt.PullRequest.Body
	return nil
}

// parsePullRequestReviewEvent pull_request_reviewイベントをパースする
func (summary *EventSummary) parsePullRequestReviewEvent(payload []byte) error {
	evt := github.PullRequestReviewEvent{}
	err := json.Unmarshal(payload, &evt)
	if err != nil {
		return err
	}
	// コードコメントを含んだレビューのサブミットを行うと submitted と edited 両方がイベントとして投げられる
	// -> github側がそうなっているので仕方ない
	if *evt.Action != "submitted" && *evt.Action != "edited" {
		return ErrUnhandledAction
	}
	summary.RepositoryName = *evt.Repo.Name
	summary.Title = *evt.PullRequest.Title
	summary.URL = *evt.Review.HTMLURL
	summary.Description = fmt.Sprintf("Review %v by: %v", *evt.Action, *evt.Review.User.Login)
	summary.Comment = *evt.Review.Body
	return nil
}

// parsePullRequestReviewCommentEvent pull_request_review_commentイベントをパースする
func (summary *EventSummary) parsePullRequestReviewCommentEvent(payload []byte) error {
	evt := github.PullRequestReviewCommentEvent{}
	err := json.Unmarshal(payload, &evt)
	if err != nil {
		return err
	}
	if *evt.Action != "created" && *evt.Action != "edited" {
		return ErrUnhandledAction
	}
	summary.RepositoryName = *evt.Repo.Name
	summary.Title = *evt.PullRequest.Title
	summary.URL = *evt.Comment.HTMLURL
	summary.Description = fmt.Sprintf("Comment %v by: %v", *evt.Action, *evt.Comment.User.Login)
	summary.Comment = *evt.Comment.Body
	return nil
}

// ParseEventSummary githubイベントサマリ生成
func (summary *EventSummary) ParseEventSummary(hc HookContext) error {
	switch hc.Event {
	case "issues":
		return summary.parseIssuesEvent(hc.Payload)
	case "issue_comment":
		return summary.parseIssueCommentsEvent(hc.Payload)
	case "pull_request":
		return summary.parsePullRequestEvent(hc.Payload)
	case "pull_request_review":
		return summary.parsePullRequestReviewEvent(hc.Payload)
	case "pull_request_review_comment":
		return summary.parsePullRequestReviewCommentEvent(hc.Payload)
	default:
		return ErrUnhandledEvent
	}
}
