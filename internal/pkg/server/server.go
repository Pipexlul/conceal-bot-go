package server

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	spoillessvideo "github.com/pipexlul/conceal-bot-go/internal/commands/specialized/spoilless_video"
	"github.com/pipexlul/conceal-bot-go/internal/pkg/env"
	utils "github.com/pipexlul/conceal-bot-go/internal/utilities"
)

type APIServer struct {
	ConcealBot       utils.ConcealBot
	embedsCollection *mongo.Collection
	hostname         string
}

func (a *APIServer) Start(bot utils.ConcealBot) {
	a.ConcealBot = bot
	a.embedsCollection = bot.GetMongoClient().Database("command_data").Collection("embeds")
	a.hostname = env.GetString("HOSTNAME", "localhost")

	a.registerRoutes()
	port := env.GetString("PORT", "8080")
	port = fmt.Sprintf(":%s", port)

	log.Println("[INFO] Started API server at http://" + a.hostname + port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// TODO: Modularize this
func (a *APIServer) registerRoutes() {
	http.HandleFunc("/embed", func(w http.ResponseWriter, r *http.Request) {
		videoID := r.URL.Query().Get("video")
		title := r.URL.Query().Get("title")
		hideThumbnail := r.URL.Query().Get("hide_thumbnail") == "true"

		userAgent := r.Header.Get("User-Agent")
		log.Printf("[DEBUG] User Agent: %s", userAgent)

		if videoID == "" || title == "" {
			http.Error(w, "Missing required parameters", http.StatusBadRequest)
			return
		}

		thumbnailURL := "https://dummyimage.com/1280x720/000000/ffffff.png&text=Thumbnail+Hidden+Lol"
		if !hideThumbnail {
			thumbnailURL = fmt.Sprintf("https://img.youtube.com/vi/%s/0.jpg", videoID)
		}

		findQuery := bson.M{
			"video_id":     videoID,
			"custom_title": title,
		}

		var existing *spoillessvideo.Embed

		err := a.embedsCollection.FindOne(context.Background(), findQuery).Decode(&existing)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, mongo.ErrNoDocuments) {
				err = errors.New("embed not found")
				status = http.StatusNotFound
			}

			http.Error(w, err.Error(), status)
			return
		}

		if existing == nil {
			http.Error(w, "Embed not found", http.StatusNotFound)
			return
		}

		existing.Thumbnail = thumbnailURL

		a.renderEmbedHTML(w, existing)
	})
}

// TODO: This DEFINITELY should not go here, so move it where it belongs later
func (a *APIServer) renderEmbedHTML(w http.ResponseWriter, embed *spoillessvideo.Embed) {
	templatedStr := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta property="og:title" content="{{.Metadata.OGTitle}}" />
		<meta property="og:image" content="{{.Thumbnail}}" />
		<meta property="og:url" content="{{.Metadata.OGURL}}" />
		<meta property="og:description" content="{{.Metadata.Description}}" />
	</head>
	<body>
		<h1>Spoilless Video!</h1>
		<p>Click to watch</p>
	</body>
	</html>
	`

	t := template.New("embed_page")
	t, err := t.Parse(templatedStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = t.Execute(w, embed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
