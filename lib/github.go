package lib

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
)

// HookContext はGithubから受け取るJSONデータを格納する構造体
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
	ErrNotSignature     = errors.New("No signature!")
	ErrNotEvent         = errors.New("No event!")
	ErrNotEventID       = errors.New("No event id")
	ErrInvalidSignature = errors.New("Invalid signature")
)

// ParseHook Githubのリクエストをパースする関数
func (hc *HookContext) ParseHook(req *rest.Request) error {
	secret := []byte(key)

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

	if !verifySignature(secret, hc.Signature, body) {
		return ErrInvalidSignature
	}

	hc.Payload = body

	return nil
}

// FindAccounts コメントに含まれるメンションに対応するSlackアカウント情報一覧を取得する
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
func ReplaceComment(comment string, conf *Config) string {
	replaced := comment
	matches := r.FindAllStringSubmatch(comment, -1)
	for _, val := range matches {
		if account, ok := conf.Accounts[val[0]]; ok {
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
