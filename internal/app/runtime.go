package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/grokbot-2/agentcontent/internal/adapters/ai"
	fileexport "github.com/grokbot-2/agentcontent/internal/adapters/export"
	httpapi "github.com/grokbot-2/agentcontent/internal/adapters/http"
	"github.com/grokbot-2/agentcontent/internal/adapters/secrets"
	"github.com/grokbot-2/agentcontent/internal/adapters/sheets"
	"github.com/grokbot-2/agentcontent/internal/adapters/sqlite"
	"github.com/grokbot-2/agentcontent/internal/adapters/transcript"
	"github.com/grokbot-2/agentcontent/internal/adapters/youtube"
	"github.com/grokbot-2/agentcontent/internal/application"
	"github.com/grokbot-2/agentcontent/internal/config"
)

type Runtime struct {
	Cfg    config.Config
	DB     *sql.DB
	Scan   *application.ScanService
	Server *httpapi.Server
}

func New(cfg config.Config) (*Runtime, error) {
	db, err := sqlite.Open(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if cfg.SeedOnEmpty {
		if err := sqlite.SeedIfEmpty(ctx, db); err != nil {
			db.Close()
			return nil, fmt.Errorf("seed: %w", err)
		}
	}
	store := sqlite.NewStore(db)
	hub := httpapi.NewHub()
	yt := youtube.New(cfg.YouTubeAPIKey)
	aiAdp := ai.New(cfg.OpenAIKey, cfg.OpenAIBaseURL, cfg.OpenAIModel)
	ex := fileexport.New(filepath.Join(filepath.Dir(cfg.DatabasePath), "exports"))
	syncer := sheets.New(cfg.SheetsEnabled)
	_ = secrets.New()
	_ = transcript.New()

	creator := application.NewCreatorService(store, yt, hub)
	feed := application.NewFeedService(store, cfg.OutlierMultiplier, hub)
	assistant := application.NewAssistantService(store, store, aiAdp, hub)
	brand := application.NewBrandService(store, hub)
	scripts := application.NewScriptService(store, aiAdp, ex, hub)
	scan := application.NewScanService(store, store, yt, syncer, hub)

	uiDir := os.Getenv("UI_DIR")
	if uiDir == "" {
		uiDir = "web/dist"
	}

	srv := &httpapi.Server{
		Cfg:         cfg,
		Creator:     creator,
		Feed:        feed,
		AI:          assistant,
		Brand:       brand,
		Script:      scripts,
		Scan:        scan,
		Hub:         hub,
		YouTubeStub: yt.Stub(),
		AIStub:      aiAdp.Stub(),
		UIDir:       uiDir,
	}
	return &Runtime{Cfg: cfg, DB: db, Scan: scan, Server: srv}, nil
}

func (rt *Runtime) Close() error { return rt.DB.Close() }

func (rt *Runtime) HTTP() http.Handler { return rt.Server.Router() }

func (rt *Runtime) ListenAndServe() error {
	addr := ":" + rt.Cfg.Port
	s := &http.Server{
		Addr:              addr,
		Handler:           rt.HTTP(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	fmt.Fprintf(os.Stderr, "AgentContent listening on http://localhost%s (sqlite=%s stub_yt=%v stub_ai=%v)\n",
		addr, rt.Cfg.DatabasePath, rt.Server.YouTubeStub, rt.Server.AIStub)
	return s.ListenAndServe()
}
