package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/miyanokomiya/gosla2/lib"
)

func main() {
	// Heroku環境を考慮してポートを取得
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("use port: " + port)

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/", root),
		rest.Post("/github/events", postGithubEvents),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":"+port, api.MakeHandler()))
}

// 生存確認用ルート画面
func root(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(`{"res": "Hello!"}`)
}

// postGithubEvents Githubイベント連携関数
func postGithubEvents(w rest.ResponseWriter, r *rest.Request) {
	conf, err := lib.ParseConfigFile("./config.toml")
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hc := lib.HookContext{}
	err = hc.ParseHook(r, conf.Secret)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	summary := lib.EventSummary{}
	err = summary.ParseEventSummary(hc)
	if err == lib.ErrUnhandledEvent || err == lib.ErrUnhandledAction {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	accounts := lib.FindAccounts(summary.Comment, conf)
	summary.ReplaceComment(accounts)
	text := lib.CreatePostText(summary)
	lib.PostToAccounts(text, accounts)
	w.WriteJson(`{"res": "finished"}`)
}
