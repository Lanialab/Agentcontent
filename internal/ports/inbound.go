package ports

import (
	"context"

	"github.com/grokbot-2/agentcontent/internal/domain/brand"
	"github.com/grokbot-2/agentcontent/internal/domain/intelligence"
	"github.com/grokbot-2/agentcontent/internal/domain/research"
	"github.com/grokbot-2/agentcontent/internal/domain/script"
)

// Inbound ports — owned by the application core.

type AddChannelCmd struct {
	GroupID string
	Handle  string
	Notes   string
}

type EditChannelCmd struct {
	ID      string
	Title   string
	Notes   string
	Handle  string
	GroupID string
}

type FeedFilter struct {
	GroupID  string
	Query    string
	Outliers bool
	Hidden   bool
	MinScore float64
}

type ChatCmd struct {
	Message    string
	TemplateID string
	Mentions   []string
}

type IdeaCmd struct {
	Prompt     string
	TemplateID string
	Mentions   []string
}

type ScriptCmd struct {
	Mode  string
	Topic string
	Title string
	Body  string
}

type CreatorCommandPort interface {
	ListGroups(ctx context.Context) ([]research.Group, error)
	CreateGroup(ctx context.Context, name string) (research.Group, error)
	ListChannels(ctx context.Context, includeHidden bool) ([]research.Channel, error)
	AddChannel(ctx context.Context, cmd AddChannelCmd) (research.Channel, error)
	EditChannel(ctx context.Context, cmd EditChannelCmd) (research.Channel, error)
	ValidateChannel(ctx context.Context, id string) (research.Channel, error)
	HideChannel(ctx context.Context, id string, hidden bool) (research.Channel, error)
	MoveChannel(ctx context.Context, id, groupID string) (research.Channel, error)
}

type FeedQueryPort interface {
	ListFeed(ctx context.Context, f FeedFilter) ([]research.ScoredVideo, error)
}

type AIWorkflowPort interface {
	Chat(ctx context.Context, cmd ChatCmd) (intelligence.ChatMessage, error)
	ListChat(ctx context.Context) ([]intelligence.ChatMessage, error)
	GenerateIdea(ctx context.Context, cmd IdeaCmd) (intelligence.Idea, error)
	ListIdeas(ctx context.Context) ([]intelligence.Idea, error)
	Templates() []intelligence.Template
}

type BrandPort interface {
	Get(ctx context.Context) (brand.Blueprint, error)
	Save(ctx context.Context, bp brand.Blueprint) (brand.Blueprint, error)
}

type VideoScriptPort interface {
	List(ctx context.Context) ([]script.Script, error)
	Generate(ctx context.Context, cmd ScriptCmd) (script.Script, error)
	Save(ctx context.Context, s script.Script) (script.Script, error)
	Export(ctx context.Context, id, format string) (path string, err error)
}

type ScheduledScanPort interface {
	Run(ctx context.Context) (ScanResult, error)
}

type ScanResult struct {
	Channels int  `json:"channels"`
	Videos   int  `json:"videos"`
	Skipped  bool `json:"skipped"`
	Stub     bool `json:"stub"`
}
