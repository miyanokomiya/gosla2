package lib

import (
	"testing"
)

func TestCreatePostText(t *testing.T) {
	type fromTo struct {
		from EventSummary
		to   string
	}
	table := map[string]fromTo{
		"case1": fromTo{
			from: EventSummary{
				RepositoryName: "repo",
				Title:          "tit",
				URL:            "url",
				Description:    "desc",
				Comment:        "comm",
			},
			to: "*[repo] tit*\nurl\n> desc\ncomm",
		},
		"case2": fromTo{
			from: EventSummary{
				RepositoryName: "repo aaaa",
				Title:          "tit",
				URL:            "url",
				Description:    "desc",
				Comment:        "comm\naaa\naaaeee",
			},
			to: "*[repo aaaa] tit*\nurl\n> desc\ncomm\naaa\naaaeee",
		},
	}
	for key, fromTo := range table {
		result := CreatePostText(fromTo.from)
		if result != fromTo.to {
			t.Fatal("failed: "+key, "expect: "+fromTo.to, "actual: "+result)
		}
	}
}
