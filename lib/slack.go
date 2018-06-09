package lib

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// SendToSlack Slackへのメッセージ送信
func SendToSlack(path string, text string) (string, error) {
	slackURL := "https://hooks.slack.com"
	slackPath := path
	u, _ := url.ParseRequestURI(slackURL)
	u.Path = slackPath

	urlStr := fmt.Sprintf("%v", u)

	data := url.Values{}
	data.Set("payload", "{\"text\": \""+text+"\"}")

	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	return string(b), nil
}

// CreateHeaderLine メッセージのヘッダ部分作成
func CreateHeaderLine(emoji string, repositoryName string, title string) string {
	return fmt.Sprintf("%v *[%v] %v*\n", emoji, repositoryName, title)
}
