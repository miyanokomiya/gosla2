package lib

import (
	"encoding/json"
	"testing"

	"github.com/google/go-github/github"
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
	result1 := FindAccounts("@a", config)
	if account, ok := result1["@a"]; ok {
		if account.ID != "@aa" || account.Channel != "aaa" {
			t.Fatal("get invalid account")
		}
	} else {
		t.Fatal("cannot get account")
	}
	result2 := FindAccounts("@b", config)
	if account, ok := result2["@b"]; ok {
		if account.ID != "@bb" || account.Channel != "bbb" {
			t.Fatal("get invalid account")
		}
	} else {
		t.Fatal("cannot get account")
	}
	result3 := FindAccounts("@c", config)
	if _, ok := result3["@c"]; ok {
		t.Fatal("get invalid account")
	}
}

func TestReplaceComment(t *testing.T) {
	accounts := map[string]Account{
		"@a": Account{
			ID:      "@aa",
			Channel: "aaa",
		},
		"@b": Account{
			ID:      "@bb",
			Channel: "bbb",
		},
	}
	table := map[string]string{
		"@a abcd":      "<@aa> abcd",
		"abcd @a abcd": "abcd <@aa> abcd",
		"abcd @b abcd": "abcd <@bb> abcd",
	}
	for from, to := range table {
		summary := EventSummary{
			Comment: from,
		}
		summary.ReplaceComment(accounts)
		if summary.Comment != to {
			t.Fatal("get invalid text", "\nfrom: "+from, "\nexpected: "+to, "\nactual: "+summary.Comment)
		}
	}
}

func TestParseIssuesEvent(t *testing.T) {
	opened := "opened"
	edited := "edited"
	other := "other"
	repoName := "repo-name"
	issueTitle := "issue-title"
	htmlURL := "url"
	user := "user"
	body := "body"
	evt := github.IssueCommentEvent{
		Action: &opened,
		Repo: &github.Repository{
			Name: &repoName,
		},
		Issue: &github.Issue{
			Title:   &issueTitle,
			HTMLURL: &htmlURL,
			User: &github.User{
				Login: &user,
			},
			Body: &body,
		},
	}
	evtJSON, _ := json.Marshal(evt)
	summary := EventSummary{}
	err := summary.parseIssuesEvent(evtJSON)
	if err != nil {
		t.Fatal("failed: parse action: opened")
	}
	if summary.RepositoryName != repoName {
		t.Fatal("failed: RepositoryName")
	}
	if summary.Title != issueTitle {
		t.Fatal("failed: Title")
	}
	if summary.URL != htmlURL {
		t.Fatal("failed: URL")
	}
	if summary.Description != "Issue "+opened+" by: "+user {
		t.Fatal("failed: Description")
	}
	if summary.Comment != body {
		t.Fatal("failed: Comment")
	}

	evt.Action = &edited
	evtJSON, _ = json.Marshal(evt)
	summary = EventSummary{}
	err = summary.parseIssuesEvent(evtJSON)
	if err != nil {
		t.Fatal("failed: parse action: edited")
	}
	if summary.Description != "Issue "+edited+" by: "+user {
		t.Fatal("failed: Description")
	}

	evt.Action = &other
	evtJSON, _ = json.Marshal(evt)
	summary = EventSummary{}
	err = summary.parseIssuesEvent(evtJSON)
	if err != ErrUnhandledAction {
		t.Fatal("failed: unexpected error")
	}
}

func TestParseIssueCommentsEvent(t *testing.T) {
	created := "created"
	edited := "edited"
	other := "other"
	repoName := "repo-name"
	issueTitle := "issue-title"
	htmlURL := "url"
	user := "user"
	body := "body"
	evt := github.IssueCommentEvent{
		Action: &created,
		Repo: &github.Repository{
			Name: &repoName,
		},
		Issue: &github.Issue{
			Title: &issueTitle,
		},
		Comment: &github.IssueComment{
			HTMLURL: &htmlURL,
			User: &github.User{
				Login: &user,
			},
			Body: &body,
		},
	}
	evtJSON, _ := json.Marshal(evt)
	summary := EventSummary{}
	err := summary.parseIssueCommentsEvent(evtJSON)
	if err != nil {
		t.Fatal("failed: parse action: created")
	}
	if summary.RepositoryName != repoName {
		t.Fatal("failed: RepositoryName")
	}
	if summary.Title != issueTitle {
		t.Fatal("failed: Title")
	}
	if summary.URL != htmlURL {
		t.Fatal("failed: URL")
	}
	if summary.Description != "Comment "+created+" by: "+user {
		t.Fatal("failed: Description")
	}
	if summary.Comment != body {
		t.Fatal("failed: Comment")
	}

	evt.Action = &edited
	evtJSON, _ = json.Marshal(evt)
	summary = EventSummary{}
	err = summary.parseIssueCommentsEvent(evtJSON)
	if err != nil {
		t.Fatal("failed: parse action: edited")
	}
	if summary.Description != "Comment "+edited+" by: "+user {
		t.Fatal("failed: Description")
	}

	evt.Action = &other
	evtJSON, _ = json.Marshal(evt)
	summary = EventSummary{}
	err = summary.parseIssueCommentsEvent(evtJSON)
	if err != ErrUnhandledAction {
		t.Fatal("failed: unexpected error")
	}
}

func TestParsePullRequestEvent(t *testing.T) {
	opened := "opened"
	edited := "edited"
	other := "other"
	repoName := "repo-name"
	prTitle := "pr-title"
	htmlURL := "url"
	user := "user"
	body := "body"
	evt := github.PullRequestEvent{
		Action: &opened,
		Repo: &github.Repository{
			Name: &repoName,
		},
		PullRequest: &github.PullRequest{
			Title:   &prTitle,
			HTMLURL: &htmlURL,
			User: &github.User{
				Login: &user,
			},
			Body: &body,
		},
	}
	evtJSON, _ := json.Marshal(evt)
	summary := EventSummary{}
	err := summary.parsePullRequestEvent(evtJSON)
	if err != nil {
		t.Fatal("failed: parse action: opened")
	}
	if summary.RepositoryName != repoName {
		t.Fatal("failed: RepositoryName")
	}
	if summary.Title != prTitle {
		t.Fatal("failed: Title")
	}
	if summary.URL != htmlURL {
		t.Fatal("failed: URL")
	}
	if summary.Description != "PullRequest "+opened+" by: "+user {
		t.Fatal("failed: Description")
	}
	if summary.Comment != body {
		t.Fatal("failed: Comment")
	}

	evt.Action = &edited
	evtJSON, _ = json.Marshal(evt)
	summary = EventSummary{}
	err = summary.parsePullRequestEvent(evtJSON)
	if err != nil {
		t.Fatal("failed: parse action: edited")
	}
	if summary.Description != "PullRequest "+edited+" by: "+user {
		t.Fatal("failed: Description")
	}

	evt.Action = &other
	evtJSON, _ = json.Marshal(evt)
	summary = EventSummary{}
	err = summary.parsePullRequestEvent(evtJSON)
	if err != ErrUnhandledAction {
		t.Fatal("failed: unexpected error")
	}
}

func TestParsePullRequestReviewEvent(t *testing.T) {
	submitted := "submitted"
	edited := "edited"
	other := "other"
	repoName := "repo-name"
	prTitle := "pr-title"
	htmlURL := "url"
	user := "user"
	body := "body"
	evt := github.PullRequestReviewEvent{
		Action: &submitted,
		Repo: &github.Repository{
			Name: &repoName,
		},
		PullRequest: &github.PullRequest{
			Title: &prTitle,
		},
		Review: &github.PullRequestReview{
			HTMLURL: &htmlURL,
			User: &github.User{
				Login: &user,
			},
			Body: &body,
		},
	}
	evtJSON, _ := json.Marshal(evt)
	summary := EventSummary{}
	err := summary.parsePullRequestReviewEvent(evtJSON)
	if err != nil {
		t.Fatal("failed: parse action: submitted")
	}
	if summary.RepositoryName != repoName {
		t.Fatal("failed: RepositoryName")
	}
	if summary.Title != prTitle {
		t.Fatal("failed: Title")
	}
	if summary.URL != htmlURL {
		t.Fatal("failed: URL")
	}
	if summary.Description != "Review "+submitted+" by: "+user {
		t.Fatal("failed: Description")
	}
	if summary.Comment != body {
		t.Fatal("failed: Comment")
	}

	evt.Action = &edited
	evtJSON, _ = json.Marshal(evt)
	summary = EventSummary{}
	err = summary.parsePullRequestReviewEvent(evtJSON)
	if err != nil {
		t.Fatal("failed: parse action: edited")
	}
	if summary.Description != "Review "+edited+" by: "+user {
		t.Fatal("failed: Description")
	}

	evt.Action = &other
	evtJSON, _ = json.Marshal(evt)
	summary = EventSummary{}
	err = summary.parsePullRequestReviewEvent(evtJSON)
	if err != ErrUnhandledAction {
		t.Fatal("failed: unexpected error")
	}
}

func TestParsePullRequestReviewCommentEvent(t *testing.T) {
	created := "created"
	edited := "edited"
	other := "other"
	repoName := "repo-name"
	prTitle := "pr-title"
	htmlURL := "url"
	user := "user"
	body := "body"
	evt := github.PullRequestReviewCommentEvent{
		Action: &created,
		Repo: &github.Repository{
			Name: &repoName,
		},
		PullRequest: &github.PullRequest{
			Title: &prTitle,
		},
		Comment: &github.PullRequestComment{
			HTMLURL: &htmlURL,
			User: &github.User{
				Login: &user,
			},
			Body: &body,
		},
	}
	evtJSON, _ := json.Marshal(evt)
	summary := EventSummary{}
	err := summary.parsePullRequestReviewCommentEvent(evtJSON)
	if err != nil {
		t.Fatal("failed: parse action: created")
	}
	if summary.RepositoryName != repoName {
		t.Fatal("failed: RepositoryName")
	}
	if summary.Title != prTitle {
		t.Fatal("failed: Title")
	}
	if summary.URL != htmlURL {
		t.Fatal("failed: URL")
	}
	if summary.Description != "Comment "+created+" by: "+user {
		t.Fatal("failed: Description")
	}
	if summary.Comment != body {
		t.Fatal("failed: Comment")
	}

	evt.Action = &edited
	evtJSON, _ = json.Marshal(evt)
	summary = EventSummary{}
	err = summary.parsePullRequestReviewCommentEvent(evtJSON)
	if err != nil {
		t.Fatal("failed: parse action: edited")
	}
	if summary.Description != "Comment "+edited+" by: "+user {
		t.Fatal("failed: Description")
	}

	evt.Action = &other
	evtJSON, _ = json.Marshal(evt)
	summary = EventSummary{}
	err = summary.parsePullRequestReviewCommentEvent(evtJSON)
	if err != ErrUnhandledAction {
		t.Fatal("failed: unexpected error")
	}
}
