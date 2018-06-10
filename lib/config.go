package lib

import (
	"github.com/BurntSushi/toml"
)

// Config GithubとSlackの連携情報を格納する構造体
type Config struct {
	Secret   string             `toml:"secret"`
	Accounts map[string]Account `toml:"accounts"`
}

// Account Slackアカウント情報
type Account struct {
	ID      string `toml:"id"`
	Channel string `toml:"channel"`
}

// ParseConfigFile 設定ファイルをパースする関数
func ParseConfigFile(filename string) (*Config, error) {
	c := Config{}
	_, err := toml.DecodeFile(filename, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
