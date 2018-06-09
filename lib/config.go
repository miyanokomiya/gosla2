package lib

import (
	"encoding/json"
	"io/ioutil"
)

// Config GithubとSlackの連携情報を格納する構造体
type Config struct {
	Accounts     map[string]Account `json:"accounts"`
	Repositories map[string]string  `json:"repositories"`
}

// Account Slackアカウント情報
type Account struct {
	ID      string `json:"id"`
	Channel string `json:"channel"`
}

// ParseFile 設定ファイルをパースする関数
func ParseFile(filename string) (*Config, error) {
	c := Config{}

	jsonString, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonString, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetEndPoint Slack投稿先を取得する関数
func GetEndPoint(repositoryName string, conf *Config) string {
	endPoint, _ := conf.Repositories[repositoryName]
	return endPoint
}
