package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/google/go-github/github"
	"github.com/miyanokomiya/gosla2/lib"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("use port: " + port)

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/", home),
		rest.Post("/github/events", PostGithubEvents),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":"+port, api.MakeHandler()))
}

func home(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(`{"res": "Hello!"}`)
}

// PostGithubEvents Githubイベント連携関数
func PostGithubEvents(w rest.ResponseWriter, r *rest.Request) {

	hc := lib.HookContext{}
	err := hc.ParseHook(r)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(hc.Payload) == 0 {
		rest.Error(w, "Payload Size is 0.", http.StatusInternalServerError)
		return
	}

	conf, err := lib.ParseFile("./config.json")
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	text := ""
	repositoryName := ""

	switch hc.Event {
	case "issues":

		evt := github.IssuesEvent{}
		if err := json.Unmarshal(hc.Payload, &evt); err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
		}
		repositoryName = *evt.Repo.Name

		if *evt.Action == "opened" {
			text = fmt.Sprintf("%v :github-status-red: *【%v】%v* :github-status-red: \n", text, repositoryName, *evt.Issue.Title)
			text = fmt.Sprintf("%v%v\n", text, *evt.Issue.HTMLURL)
			text = fmt.Sprintf("%v>Issue opened by: %v\n", text, *evt.Issue.User.Login)

			comment := lib.ReplaceComment(*evt.Issue.Body, conf)
			text = fmt.Sprintf("%v\n%v\n", text, comment)
		}

	case "issue_comment":

		evt := github.IssueCommentEvent{}
		if err := json.Unmarshal(hc.Payload, &evt); err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
		}
		repositoryName = *evt.Repo.Name

		if *evt.Action == "created" {
			text = lib.CreateHeaderLine(":speaking_head_in_silhouette:", repositoryName, *evt.Issue.Title)
			text = fmt.Sprintf("%v%v\n", text, *evt.Comment.HTMLURL)
			text = fmt.Sprintf("%v>Comment created by: %v\n", text, *evt.Comment.User.Login)

			comment := lib.ReplaceComment(*evt.Comment.Body, conf)
			text = fmt.Sprintf("%v\n%v\n", text, comment)
		}

	case "pull_request":

		evt := github.PullRequestEvent{}
		if err := json.Unmarshal(hc.Payload, &evt); err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
		}
		repositoryName = *evt.Repo.Name

		if *evt.Action == "opened" {
			text = fmt.Sprintf("%v :github-status-red: *【%v】%v* :github-status-red: \n", text, repositoryName, *evt.PullRequest.Title)
			text = fmt.Sprintf("%v%v\n", text, *evt.PullRequest.HTMLURL)
			text = fmt.Sprintf("%v>PullRequest opened by: %v\n", text, *evt.PullRequest.User.Login)

			comment := lib.ReplaceComment(*evt.PullRequest.Body, conf)
			text = fmt.Sprintf("%v\n%v\n", text, comment)
		}

	case "pull_request_review_comment":
		evt := github.PullRequestReviewCommentEvent{}
		if err := json.Unmarshal(hc.Payload, &evt); err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
		}
		repositoryName = *evt.Repo.Name

		if *evt.Action == "created" {
			text = fmt.Sprintf("%v :github-status-orange: *[%v]%v* :github-status-orange: \n", text, repositoryName, *evt.PullRequest.Title)
			text = fmt.Sprintf("%v%v\n", text, *evt.Comment.HTMLURL)
			text = fmt.Sprintf("%v>ReviewComment created by: %v\n", text, *evt.Comment.User.Login)

			comment := lib.ReplaceComment(*evt.Comment.Body, conf)
			text = fmt.Sprintf("%v\n%v\n", text, comment)
		}
	default:
	}

	if text != "" && repositoryName != "" {
		endPoint := lib.GetEndPoint("myChannel", conf)
		res, err := lib.SendToSlack(endPoint, text)

		if err != nil {
			w.WriteJson(fmt.Sprintf(`{"res": "%v", "error": "%v"}`, res, err))
		} else {
			w.WriteJson(fmt.Sprintf(`{"res": "%v"}`, res))
		}
	} else {
		w.WriteJson(`{"res": "text or repository's name not exist"}`)
	}
}

// func main() {
// 	logger, _ := zap.NewDevelopment()
// 	defer logger.Sync()
// 	logger.Info("Hello zap", zap.String("key", "value"), zap.Time("now", time.Now()))
// 	conf, err := lib.ParseFile("./config.json")
// 	endPoint := lib.GetEndPoint("myChannel", conf)
// 	text := lib.ReplaceComment("@miyanokomiya aaaaaaaaaaaaaaa", conf)
// 	res, err := lib.SendToSlack(endPoint, text)

// 	if err != nil {
// 		println(fmt.Sprintf(`{"res": "%v", "error": "%v"}`, res, err))
// 	} else {
// 		println(fmt.Sprintf(`{"res": "%v"}`, res))
// 	}
// }
