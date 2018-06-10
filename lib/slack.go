package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var slackURL = "https://hooks.slack.com"
var contentType = "application/x-www-form-urlencoded"

// sendToSlack Slackへのメッセージ送信
func sendToSlack(path string, text string) (string, error) {
	u, _ := url.ParseRequestURI(slackURL)
	u.Path = path
	urlStr := u.String()

	data := url.Values{}
	data.Set("payload", "{\"text\": \""+text+"\"}")

	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", contentType)

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	return string(b), nil
}

// CreatePostText 投稿用テキストを生成する
func CreatePostText(summary EventSummary) string {
	text := fmt.Sprintf("*[%v] %v*", summary.RepositoryName, summary.Title)
	text = fmt.Sprintf("%v\n%v", text, summary.URL)
	text = fmt.Sprintf("%v\n> %v", text, summary.Description)
	text = fmt.Sprintf("%v\n%v", text, summary.Comment)
	return text
}

// PostToAccounts Slackアカウント宛に送信
// 送信エラーは無視して続行
func PostToAccounts(text string, accounts map[string]Account) {
	for key, account := range accounts {
		log.Println("start: send to Slack: " + key + ", channel: " + account.Channel)
		_, err := sendToSlack(account.Channel, text)
		if err != nil {
			log.Println("failed: send to Slack: "+key, err)
		} else {
			log.Println("success: send to Slack: " + key)
		}
	}
}
