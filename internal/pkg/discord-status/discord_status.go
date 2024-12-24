package discordstatus

import (
	"log"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

var statuses = []string{
	"Epic Anya very voice!",
	"Nice projection Mellow",
	"It's time to become pedantic",
	"This is why Pipex hates you",
	"Any middle namers?!",
	"The living embodiment of yellow",
	"HEY BEBEH",
	"Gucci Gang Gucci Gang Gucci Gang Gucci Gang",
	"The most important thing in life is to be a good person",
	"I need to do my history homework",
	"Yooooooouuuuuuuuuuuuuu",
	"Go mentioned",
	"In the secret chat",
	"Hide in the bunker",
	"40% on cats lmao",
	"Current cat count: 11",
	"Current GFuel count: 42",
	"Me hear keyboard me clacky!!",
	"lol, lmao even",
	"Hes him",
	"Swizz's ceiling",
	"Swizz's floor",
	"Swizz's wall",
	"Swizz's door",
	"Swizz's window",
	"We love jeffertons",
	"Mellow's scripts EPIC!",
	"Jace is squared, Pipex is rounded",
	"Silco will never be real",
	"We simp Jinx",
	"Miquella, lord of projections",
}

type Helper struct {
	s       *discordgo.Session
	randGen *rand.Rand
}

func New(s *discordgo.Session, randGen *rand.Rand) *Helper {
	return &Helper{
		s:       s,
		randGen: randGen,
	}
}

func (dsh *Helper) SetupStatusTicker() {
	ticker := time.Tick(time.Hour * 2)
	go func() {
		for range ticker {
			if err := dsh.UpdateStatusFromRandom(); err != nil {
				log.Printf("Failed to update game status: %v", err)
			}
		}
	}()
}

func (dsh *Helper) UpdateStatusFromRandom() error {
	status := dsh.GetRandomStatus(statuses)

	log.Printf("Updating game status to '%s'", status)
	return dsh.s.UpdateGameStatus(0, status)
}

func (dsh *Helper) GetRandomStatus(statusesList []string) string {
	return statusesList[dsh.randGen.Intn(len(statusesList))]
}
