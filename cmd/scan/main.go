package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/grokbot-2/agentcontent/internal/app"
	"github.com/grokbot-2/agentcontent/internal/config"
)

// CLI / job entrypoint. Native OS scheduler (launchd / Task Scheduler / cron)
// should invoke this every 60 minutes. Missed ticks are skipped via SQLite lease.
func main() {
	rt, err := app.New(config.Load())
	if err != nil {
		log.Fatal(err)
	}
	defer rt.Close()
	res, err := rt.Scan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(res)
	if res.Skipped {
		fmt.Fprintln(os.Stderr, "scan skipped (lease held)")
	}
}
