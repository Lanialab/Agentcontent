package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/grokbot-2/agentcontent/internal/domain/intelligence"
	"github.com/grokbot-2/agentcontent/internal/idgen"
	"github.com/grokbot-2/agentcontent/internal/ports"
)

type AssistantService struct {
	store ports.AssistantStore
	brand ports.BrandStore
	ai    ports.AIProviderPort
	trace ports.Tracer
}

func NewAssistantService(store ports.AssistantStore, brand ports.BrandStore, ai ports.AIProviderPort, trace ports.Tracer) *AssistantService {
	return &AssistantService{store: store, brand: brand, ai: ai, trace: trace}
}

func (s *AssistantService) Templates() []intelligence.Template {
	return intelligence.DefaultTemplates()
}

func (s *AssistantService) ListChat(ctx context.Context) ([]intelligence.ChatMessage, error) {
	return s.store.ListChat(ctx)
}

func (s *AssistantService) ListIdeas(ctx context.Context) ([]intelligence.Idea, error) {
	s.pulse("list-ideas")
	return s.store.ListIdeas(ctx)
}

func (s *AssistantService) Chat(ctx context.Context, cmd ports.ChatCmd) (intelligence.ChatMessage, error) {
	if strings.TrimSpace(cmd.Message) == "" {
		return intelligence.ChatMessage{}, fmt.Errorf("message is required")
	}
	s.pulse("chat")
	user := intelligence.ChatMessage{
		ID:        idgen.New("msg"),
		Role:      "user",
		Content:   cmd.Message,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.store.SaveChat(ctx, user); err != nil {
		return intelligence.ChatMessage{}, err
	}
	system := s.systemPrompt(ctx, cmd.TemplateID, cmd.Mentions)
	reply, err := s.ai.Complete(ctx, system, cmd.Message)
	if err != nil {
		return intelligence.ChatMessage{}, err
	}
	assistant := intelligence.ChatMessage{
		ID:        idgen.New("msg"),
		Role:      "assistant",
		Content:   reply,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.store.SaveChat(ctx, assistant); err != nil {
		return intelligence.ChatMessage{}, err
	}
	return assistant, nil
}

func (s *AssistantService) GenerateIdea(ctx context.Context, cmd ports.IdeaCmd) (intelligence.Idea, error) {
	prompt := strings.TrimSpace(cmd.Prompt)
	if prompt == "" {
		prompt = "Đề xuất 1 ý tưởng video khớp Brand Blueprint và 1 outlier đang thắng."
	}
	s.pulse("generate-idea")
	if s.trace != nil {
		s.trace.Emit(ports.TraceEvent{
			Path:  "idea",
			Kind:  ports.TraceExternal,
			Label: "ai-complete",
			Nodes: []string{"uc-ai", "port-ai", "adp-ai", "ext-ai"},
		})
	}
	system := s.systemPrompt(ctx, cmd.TemplateID, cmd.Mentions)
	body, err := s.ai.Complete(ctx, system, prompt)
	if err != nil {
		return intelligence.Idea{}, err
	}
	title := firstLine(body)
	idea := intelligence.Idea{
		ID:         idgen.New("idea"),
		Title:      title,
		Prompt:     prompt,
		Body:       body,
		Mentions:   cmd.Mentions,
		TemplateID: cmd.TemplateID,
		CreatedAt:  time.Now().UTC(),
	}
	if err := intelligence.ValidateIdea(idea.Title, idea.Body); err != nil {
		return intelligence.Idea{}, err
	}
	if s.trace != nil {
		s.trace.Emit(ports.TraceEvent{
			Path:  "idea",
			Kind:  ports.TracePersist,
			Label: "save-idea",
			Nodes: []string{"uc-ai", "port-assistant-store", "adp-sqlite"},
		})
	}
	if err := s.store.SaveIdea(ctx, idea); err != nil {
		return intelligence.Idea{}, err
	}
	return idea, nil
}

func (s *AssistantService) pulse(label string) {
	if s.trace == nil {
		return
	}
	s.trace.Emit(ports.TraceEvent{
		Path:  "idea",
		Kind:  ports.TraceCommand,
		Label: label,
		Nodes: []string{"ui-ai", "http-ai", "port-ai-in", "uc-ai", "dom-intel", "port-ai", "adp-ai", "ext-ai"},
	})
}

func (s *AssistantService) systemPrompt(ctx context.Context, templateID string, mentions []string) string {
	bp, _ := s.brand.GetBlueprint(ctx)
	tpl := ""
	for _, t := range intelligence.DefaultTemplates() {
		if t.ID == templateID {
			tpl = t.Prompt
			break
		}
	}
	return fmt.Sprintf(`You are the AgentContent intelligence layer for a Vietnamese creator OS.
Brand positioning: %s
Voice: %s
Pillars: %s
Template: %s
Mentions: %s
Reply in Vietnamese unless the user writes in English. Be concrete. Reference outlier math (viewsPerDay vs same-channel average, default 2.5x — the architecture diagram's *100 is visual seed exaggeration, not the product threshold).`,
		bp.Positioning, bp.Voice, strings.Join(bp.Pillars, ", "), tpl, strings.Join(mentions, ", "))
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i > 0 {
		s = s[:i]
	}
	s = strings.TrimLeft(s, "# ")
	if len(s) > 80 {
		return s[:80]
	}
	if s == "" {
		return "Ý tưởng mới"
	}
	return s
}
