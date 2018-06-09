package lib

import (
	"testing"
)

func TestFindAccounts(t *testing.T) {
	config := Config{
		Accounts: map[string]Account{
			"@a": Account{
				ID:      "@aa",
				Channel: "aaa",
			},
			"@b": Account{
				ID:      "@bb",
				Channel: "bbb",
			},
		},
	}
	result1 := FindAccounts("@a", &config)
	if account, ok := result1["@a"]; ok {
		if account.ID != "@aa" || account.Channel != "aaa" {
			t.Fatal("get invalid account")
		}
	} else {
		t.Fatal("cannot get account")
	}
	result2 := FindAccounts("@b", &config)
	if account, ok := result2["@b"]; ok {
		if account.ID != "@bb" || account.Channel != "bbb" {
			t.Fatal("get invalid account")
		}
	} else {
		t.Fatal("cannot get account")
	}
	result3 := FindAccounts("@c", &config)
	if _, ok := result3["@c"]; ok {
		t.Fatal("get invalid account")
	}
}
