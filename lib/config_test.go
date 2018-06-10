package lib

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestParseFile(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal("failed: create tmp file", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	content := []byte("secret = \"aaa\"\n[accounts.\"bbb\"]\nid = \"ccc\"\nchannel = \"ddd\"")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatal(err)
	}

	config := Config{}
	err = config.ParseFile(tmpFile.Name())
	if err != nil {
		t.Fatal("failed: parse file", err)
	}
	if config.Secret != "aaa" {
		t.Fatal("failed: parse Secret", config.Secret)
	}
	account, ok := config.Accounts["bbb"]
	if !ok {
		t.Fatal("failed: parse Accounts")
	}
	if account.ID != "ccc" {
		t.Fatal("failed: parse account.ID", account.ID)
	}
	if account.Channel != "ddd" {
		t.Fatal("failed: parse account.Channel", account.Channel)
	}
}
