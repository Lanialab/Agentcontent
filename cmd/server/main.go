package main

import (
	"log"

	"github.com/grokbot-2/agentcontent/internal/app"
	"github.com/grokbot-2/agentcontent/internal/config"
)

func main() {
	rt, err := app.New(config.Load())
	if err != nil {
		log.Fatal(err)
	}
	defer rt.Close()
	if err := rt.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
