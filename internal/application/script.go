package application

import (
	"context"
	"strings"
	"time"

	"github.com/grokbot-2/agentcontent/internal/domain/script"
	"github.com/grokbot-2/agentcontent/internal/idgen"
	"github.com/grokbot-2/agentcontent/internal/ports"
)

type ScriptService struct {
	store  ports.ScriptStore
	ai     ports.AIProviderPort
	export ports.FileExportPort
	trace  ports.Tracer
}

func NewScriptService(store ports.ScriptStore, ai ports.AIProviderPort, export ports.FileExportPort, trace ports.Tracer) *ScriptService {
	return &ScriptService{store: store, ai: ai, export: export, trace: trace}
}

func (s *ScriptService) List(ctx context.Context) ([]script.Script, error) {
	s.pulse("list-scripts")
	return s.store.ListScripts(ctx)
}

func (s *ScriptService) Generate(ctx context.Context, cmd ports.ScriptCmd) (script.Script, error) {
	mode := script.ParseMode(cmd.Mode)
	topic := strings.TrimSpace(cmd.Topic)
	if topic == "" {
		topic = strings.TrimSpace(cmd.Title)
	}
	if topic == "" {
		topic = "Outlier tuần này"
	}
	s.pulse("generate-script")
	if s.trace != nil {
		s.trace.Emit(ports.TraceEvent{
			Path:  "script",
			Kind:  ports.TraceExternal,
			Label: "draft-script",
			Nodes: []string{"uc-script", "port-ai", "adp-ai", "ext-ai"},
		})
	}
	outline := script.OutlineFor(mode, topic)
	body, err := s.ai.Complete(ctx, "You write Vietnamese YouTube scripts. Keep the requested mode (Nhanh / Auto / Sâu). Return markdown.", outline+"\n\n"+cmd.Body)
	if err != nil {
		body = outline
	}
	now := time.Now().UTC()
	item := script.Script{
		ID:        idgen.New("sc"),
		Mode:      mode,
		Title:     firstNonEmpty(cmd.Title, topic),
		Body:      body,
		Status:    "draft",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.SaveScript(ctx, item); err != nil {
		return script.Script{}, err
	}
	return item, nil
}

func (s *ScriptService) Save(ctx context.Context, item script.Script) (script.Script, error) {
	item.UpdatedAt = time.Now().UTC()
	if item.ID == "" {
		item.ID = idgen.New("sc")
		item.CreatedAt = item.UpdatedAt
	}
	if item.Status == "" {
		item.Status = "draft"
	}
	item.Mode = script.ParseMode(string(item.Mode))
	s.pulse("save-script")
	if err := s.store.SaveScript(ctx, item); err != nil {
		return script.Script{}, err
	}
	return item, nil
}

func (s *ScriptService) Export(ctx context.Context, id, format string) (string, error) {
	item, err := s.store.GetScript(ctx, id)
	if err != nil {
		return "", err
	}
	if s.trace != nil {
		s.trace.Emit(ports.TraceEvent{
			Path:  "script",
			Kind:  ports.TraceExternal,
			Label: "export-file",
			Nodes: []string{"uc-script", "port-export", "adp-export", "ext-files"},
		})
	}
	return s.export.Write(ctx, item.Title, format, item.Body)
}

func (s *ScriptService) pulse(label string) {
	if s.trace == nil {
		return
	}
	s.trace.Emit(ports.TraceEvent{
		Path:  "script",
		Kind:  ports.TraceCommand,
		Label: label,
		Nodes: []string{"ui-script", "http-script", "port-script", "uc-script", "dom-script", "port-script-store", "adp-sqlite"},
	})
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
