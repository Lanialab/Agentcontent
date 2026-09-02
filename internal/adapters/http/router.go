package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/grokbot-2/agentcontent/internal/config"
	"github.com/grokbot-2/agentcontent/internal/domain/brand"
	"github.com/grokbot-2/agentcontent/internal/domain/script"
	"github.com/grokbot-2/agentcontent/internal/ports"
)

type Server struct {
	Cfg         config.Config
	Creator     ports.CreatorCommandPort
	Feed        ports.FeedQueryPort
	AI          ports.AIWorkflowPort
	Brand       ports.BrandPort
	Script      ports.VideoScriptPort
	Scan        ports.ScheduledScanPort
	Hub         *Hub
	YouTubeStub bool
	AIStub      bool
	UIDir       string
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
	}))

	r.Get("/api/health", s.health)
	r.Get("/api/meta", s.meta)
	r.Get("/api/events", s.events)

	r.Get("/api/groups", s.listGroups)
	r.Post("/api/groups", s.createGroup)
	r.Get("/api/channels", s.listChannels)
	r.Post("/api/channels", s.addChannel)
	r.Patch("/api/channels/{id}", s.editChannel)
	r.Post("/api/channels/{id}/validate", s.validateChannel)
	r.Post("/api/channels/{id}/hide", s.hideChannel)
	r.Post("/api/channels/{id}/move", s.moveChannel)

	r.Get("/api/feed", s.feed)
	r.Post("/api/scan", s.scan)

	r.Get("/api/templates", s.templates)
	r.Get("/api/chat", s.listChat)
	r.Post("/api/chat", s.chat)
	r.Get("/api/ideas", s.listIdeas)
	r.Post("/api/ideas", s.generateIdea)

	r.Get("/api/brand", s.getBrand)
	r.Put("/api/brand", s.saveBrand)

	r.Get("/api/scripts", s.listScripts)
	r.Post("/api/scripts", s.generateScript)
	r.Put("/api/scripts/{id}", s.saveScript)
	r.Post("/api/scripts/{id}/export", s.exportScript)

	s.mountUI(r)
	return r
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "name": "AgentContent"})
}

func (s *Server) meta(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":              "AgentContent",
		"outlierMultiplier": s.Cfg.OutlierMultiplier,
		"youtubeStub":       s.YouTubeStub,
		"aiStub":            s.AIStub,
		"sheetsEnabled":     s.Cfg.SheetsEnabled,
		"outlierNote":       "Diagram annotates *100 as visual seed exaggeration. Product default is 2.5x (configurable 2–3x) vs same-channel average viewsPerDay.",
	})
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "sse unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ch := s.Hub.Subscribe()
	defer s.Hub.Unsubscribe(ch)
	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case ev := <-ch:
			w.Write([]byte("data: "))
			w.Write(encodeEvent(ev))
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}

func (s *Server) listGroups(w http.ResponseWriter, r *http.Request) {
	items, err := s.Creator.ListGroups(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) createGroup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, err)
		return
	}
	g, err := s.Creator.CreateGroup(r.Context(), body.Name)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, g)
}

func (s *Server) listChannels(w http.ResponseWriter, r *http.Request) {
	hidden := r.URL.Query().Get("hidden") == "1" || r.URL.Query().Get("hidden") == "true"
	items, err := s.Creator.ListChannels(r.Context(), hidden)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) addChannel(w http.ResponseWriter, r *http.Request) {
	var body ports.AddChannelCmd
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, err)
		return
	}
	ch, err := s.Creator.AddChannel(r.Context(), body)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, ch)
}

func (s *Server) editChannel(w http.ResponseWriter, r *http.Request) {
	var body ports.EditChannelCmd
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, err)
		return
	}
	body.ID = chi.URLParam(r, "id")
	ch, err := s.Creator.EditChannel(r.Context(), body)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ch)
}

func (s *Server) validateChannel(w http.ResponseWriter, r *http.Request) {
	ch, err := s.Creator.ValidateChannel(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ch)
}

func (s *Server) hideChannel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Hidden *bool `json:"hidden"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	hidden := true
	if body.Hidden != nil {
		hidden = *body.Hidden
	}
	ch, err := s.Creator.HideChannel(r.Context(), chi.URLParam(r, "id"), hidden)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ch)
}

func (s *Server) moveChannel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		GroupID string `json:"groupId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, err)
		return
	}
	ch, err := s.Creator.MoveChannel(r.Context(), chi.URLParam(r, "id"), body.GroupID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ch)
}

func (s *Server) feed(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	min, _ := strconv.ParseFloat(q.Get("minScore"), 64)
	items, err := s.Feed.ListFeed(r.Context(), ports.FeedFilter{
		GroupID:  q.Get("groupId"),
		Query:    q.Get("q"),
		Outliers: q.Get("outliers") == "1" || q.Get("outliers") == "true",
		Hidden:   q.Get("hidden") == "1",
		MinScore: min,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) scan(w http.ResponseWriter, r *http.Request) {
	res, err := s.Scan.Run(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) templates(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.AI.Templates())
}

func (s *Server) listChat(w http.ResponseWriter, r *http.Request) {
	items, err := s.AI.ListChat(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) chat(w http.ResponseWriter, r *http.Request) {
	var body ports.ChatCmd
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, err)
		return
	}
	msg, err := s.AI.Chat(r.Context(), body)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, msg)
}

func (s *Server) listIdeas(w http.ResponseWriter, r *http.Request) {
	items, err := s.AI.ListIdeas(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) generateIdea(w http.ResponseWriter, r *http.Request) {
	var body ports.IdeaCmd
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, err)
		return
	}
	idea, err := s.AI.GenerateIdea(r.Context(), body)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, idea)
}

func (s *Server) getBrand(w http.ResponseWriter, r *http.Request) {
	bp, err := s.Brand.Get(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bp)
}

func (s *Server) saveBrand(w http.ResponseWriter, r *http.Request) {
	var bp brand.Blueprint
	if err := json.NewDecoder(r.Body).Decode(&bp); err != nil {
		writeErr(w, err)
		return
	}
	saved, err := s.Brand.Save(r.Context(), bp)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (s *Server) listScripts(w http.ResponseWriter, r *http.Request) {
	items, err := s.Script.List(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) generateScript(w http.ResponseWriter, r *http.Request) {
	var body ports.ScriptCmd
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, err)
		return
	}
	item, err := s.Script.Generate(r.Context(), body)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) saveScript(w http.ResponseWriter, r *http.Request) {
	var item script.Script
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeErr(w, err)
		return
	}
	item.ID = chi.URLParam(r, "id")
	saved, err := s.Script.Save(r.Context(), item)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (s *Server) exportScript(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Format string `json:"format"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Format == "" {
		body.Format = "md"
	}
	path, err := s.Script.Export(r.Context(), chi.URLParam(r, "id"), body.Format)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": path})
}

func (s *Server) mountUI(r chi.Router) {
	dir := s.UIDir
	if dir == "" {
		dir = "web/dist"
	}
	index := filepath.Join(dir, "index.html")
	if _, err := os.Stat(index); err != nil {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.WriteString(w, `<!doctype html><meta charset="utf-8"><title>AgentContent API</title>
<body style="font-family:sans-serif;background:#070b14;color:#c8d4e8;padding:2rem">
<h1>AgentContent API</h1>
<p>UI not built. Run <code>cd web && npm install && npm run build</code> then restart.
API is live at <code>/api/health</code>.</p>`)
		})
		return
	}
	fileServer := http.FileServer(http.Dir(dir))
	r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/api/") {
			http.NotFound(w, req)
			return
		}
		rel := strings.TrimPrefix(filepath.Clean(req.URL.Path), "/")
		p := filepath.Join(dir, rel)
		if absDir, err := filepath.Abs(dir); err == nil {
			if absP, err2 := filepath.Abs(p); err2 == nil && !strings.HasPrefix(absP, absDir) {
				http.NotFound(w, req)
				return
			}
		}
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, req)
			return
		}
		http.ServeFile(w, req, index)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if err != nil && (strings.Contains(err.Error(), "not found")) {
		status = http.StatusNotFound
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
