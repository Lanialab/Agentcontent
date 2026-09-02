package ports

import (
	"context"

	"github.com/grokbot-2/agentcontent/internal/domain/brand"
	"github.com/grokbot-2/agentcontent/internal/domain/intelligence"
	"github.com/grokbot-2/agentcontent/internal/domain/research"
	"github.com/grokbot-2/agentcontent/internal/domain/script"
)

// Outbound ports — owned by the application core.

type ResearchStore interface {
	ListGroups(ctx context.Context) ([]research.Group, error)
	CreateGroup(ctx context.Context, g research.Group) error
	GetGroup(ctx context.Context, id string) (research.Group, error)
	ListChannels(ctx context.Context, includeHidden bool) ([]research.Channel, error)
	GetChannel(ctx context.Context, id string) (research.Channel, error)
	UpsertChannel(ctx context.Context, c research.Channel) error
	ListVideosByChannel(ctx context.Context, channelID string) ([]research.Video, error)
	ListAllVideos(ctx context.Context) ([]research.Video, error)
	ReplaceChannelVideos(ctx context.Context, channelID string, videos []research.Video) error
	IsEmpty(ctx context.Context) (bool, error)
}

type AssistantStore interface {
	ListIdeas(ctx context.Context) ([]intelligence.Idea, error)
	SaveIdea(ctx context.Context, idea intelligence.Idea) error
	ListChat(ctx context.Context) ([]intelligence.ChatMessage, error)
	SaveChat(ctx context.Context, msg intelligence.ChatMessage) error
}

type BrandStore interface {
	GetBlueprint(ctx context.Context) (brand.Blueprint, error)
	SaveBlueprint(ctx context.Context, bp brand.Blueprint) error
}

type ScriptStore interface {
	ListScripts(ctx context.Context) ([]script.Script, error)
	GetScript(ctx context.Context, id string) (script.Script, error)
	SaveScript(ctx context.Context, s script.Script) error
}

type JobStore interface {
	TryAcquireScanLease(ctx context.Context, owner string, until string) (bool, error)
	ReleaseScanLease(ctx context.Context, owner string, errMsg string) error
}

type RemoteChannel struct {
	YouTubeID    string
	Handle       string
	Title        string
	Description  string
	ThumbnailURL string
}

type RemoteVideo struct {
	YouTubeID       string
	Title           string
	PublishedAt     string
	ViewCount       int64
	ThumbnailURL    string
	DurationSeconds int
}

type VideoSourcePort interface {
	LookupChannel(ctx context.Context, handle string) (RemoteChannel, error)
	ListRecentVideos(ctx context.Context, youtubeID string) ([]RemoteVideo, error)
	Stub() bool
}

type TranscriptPort interface {
	Transcript(ctx context.Context, youtubeVideoID string) (string, error)
}

type AIProviderPort interface {
	Complete(ctx context.Context, system, user string) (string, error)
	Stub() bool
}

type SyncExportPort interface {
	Sync(ctx context.Context) error
	Enabled() bool
}

type SecretPort interface {
	Get(ctx context.Context, name string) (string, error)
}

type FileExportPort interface {
	Write(ctx context.Context, name, format, body string) (string, error)
}

type TraceKind string

const (
	TraceCommand  TraceKind = "command"
	TraceAsync    TraceKind = "async"
	TracePersist  TraceKind = "persist"
	TraceExternal TraceKind = "external"
)

type TraceEvent struct {
	Path  string    `json:"path"`
	Kind  TraceKind `json:"kind"`
	Label string    `json:"label"`
	Nodes []string  `json:"nodes"`
}

type Tracer interface {
	Emit(ev TraceEvent)
}
