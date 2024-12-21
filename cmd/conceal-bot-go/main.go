package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/pipexlul/conceal-bot-go/internal/commands"
	"github.com/pipexlul/conceal-bot-go/internal/mappers"
	"github.com/pipexlul/conceal-bot-go/internal/pkg/env"
)

const (
	funnyStatus = "Are we some kind of tenet?"

	hubURL      = "http://localhost:50000/hub"
	hubTopic    = "earthquakes"
	callbackURL = "http://localhost:50001/earthquakes"
)

func subscribeToRecentEarthquakes() {
	query := url.Values{}
	query.Set("hub.mode", "subscribe")
	query.Set("hub.topic", hubTopic)
	query.Set("hub.callback", callbackURL)

	url := hubURL + "?" + query.Encode()

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Fatalf("Could not create request to subscribe to earthquakes: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Could not subscribe to earthquakes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Fatalf("Wrong status code when subscribing to earthquakes: %v", resp.StatusCode)
	}

	log.Print("Successfully subscribed to earthquakes\n")
}

func websubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// TODO: Add challenge verifications
		return
	}

	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to ready post request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		log.Printf("Received message from WebSub hub: \n%s", string(body))
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Invalid method: "+r.Method, http.StatusMethodNotAllowed)
}

var mongoClient *mongo.Client

// Unused until we need mongo
func connectMongo() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("Missing MONGO_URI environment variable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB")
	mongoClient = client
}

// Unused until we need mongo
func disconnectMongo() {
	if mongoClient != nil {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}
}

func main() {
	botToken := env.GetBotToken()
	if botToken.Token == "" {
		log.Fatal("Missing all BOT TOKEN environment variables, at least one is required")
	}

	log.Printf("Starting bot in %s mode", mappers.MapTokenTypeToString(botToken.TokenType))

	dg, err := discordgo.New("Bot " + botToken.Token)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	dg.AddHandler(onReady)

	if err = dg.Open(); err != nil {
		log.Fatalf("Failed to open Discord session: %v", err)
	}
	log.Println("Discord session opened :)")

	defer func() {
		closeErr := dg.Close()
		if closeErr != nil {
			log.Fatalf("Failed to close Discord session: %v", closeErr)
		}
		log.Println("Discord session closed")
	}()

	registererCtx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	// TODO: Use a more centralized command registerer
	cmd := &commands.TimeDifferenceCmd{}
	registerCmdErr := cmd.Register(registererCtx, dg)
	if registerCmdErr != nil {
		log.Fatalf("Failed to register command: %v", cmd.GetCommandName())
	}

	log.Print("All commands registered!")

	http.HandleFunc("/earthquakes", websubCallbackHandler)
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "50001"
	}
	log.Printf("Listening on port %s", httpPort)
	httpPort = ":" + httpPort
	go func() {
		if err := http.ListenAndServe(httpPort, nil); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	subscribeToRecentEarthquakes()

	log.Println("Bot is now running. Press CTRL-C to exit.")
	select {}
}

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	if err := s.UpdateGameStatus(0, funnyStatus); err != nil {
		log.Printf("Error updating game status at ready: %v", err)
	}
	log.Printf("Ready! Logged in as: %v#%v with status: %v",
		s.State.User.Username,
		s.State.User.Discriminator,
		funnyStatus,
	)
}
