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

// ParseFile 設定ファイルをパースする関数
func (c *Config) ParseFile(filename string) error {
	_, err := toml.DecodeFile(filename, &c)
	if err != nil {
		return err
	}
	return nil
}
