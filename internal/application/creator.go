package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/grokbot-2/agentcontent/internal/domain/research"
	"github.com/grokbot-2/agentcontent/internal/idgen"
	"github.com/grokbot-2/agentcontent/internal/ports"
)

type CreatorService struct {
	store  ports.ResearchStore
	source ports.VideoSourcePort
	trace  ports.Tracer
}

func NewCreatorService(store ports.ResearchStore, source ports.VideoSourcePort, trace ports.Tracer) *CreatorService {
	return &CreatorService{store: store, source: source, trace: trace}
}

func (s *CreatorService) pulse(label string, nodes ...string) {
	if s.trace == nil {
		return
	}
	s.trace.Emit(ports.TraceEvent{
		Path:  "creator",
		Kind:  ports.TraceCommand,
		Label: label,
		Nodes: nodes,
	})
}

func (s *CreatorService) ListGroups(ctx context.Context) ([]research.Group, error) {
	s.pulse("list-groups", "ui-creators", "http-creator", "port-creator", "uc-creator", "dom-research", "port-research-store", "adp-sqlite")
	return s.store.ListGroups(ctx)
}

func (s *CreatorService) CreateGroup(ctx context.Context, name string) (research.Group, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return research.Group{}, errors.New("group name is required")
	}
	g := research.Group{ID: idgen.New("grp"), Name: name, CreatedAt: time.Now().UTC()}
	s.pulse("create-group", "ui-creators", "http-creator", "port-creator", "uc-creator", "dom-research", "port-research-store", "adp-sqlite")
	if err := s.store.CreateGroup(ctx, g); err != nil {
		return research.Group{}, err
	}
	return g, nil
}

func (s *CreatorService) ListChannels(ctx context.Context, includeHidden bool) ([]research.Channel, error) {
	s.pulse("list-channels", "ui-creators", "http-creator", "port-creator", "uc-creator", "dom-research", "port-research-store", "adp-sqlite")
	return s.store.ListChannels(ctx, includeHidden)
}

func (s *CreatorService) AddChannel(ctx context.Context, cmd ports.AddChannelCmd) (research.Channel, error) {
	handle := normalizeHandle(cmd.Handle)
	if handle == "" {
		return research.Channel{}, errors.New("handle is required")
	}
	if cmd.GroupID == "" {
		return research.Channel{}, errors.New("groupId is required")
	}
	if _, err := s.store.GetGroup(ctx, cmd.GroupID); err != nil {
		return research.Channel{}, err
	}
	now := time.Now().UTC()
	ch := research.Channel{
		ID:        idgen.New("ch"),
		GroupID:   cmd.GroupID,
		Handle:    handle,
		Title:     handle,
		YouTubeID: "",
		Notes:     cmd.Notes,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.pulse("add-channel", "ui-creators", "http-creator", "port-creator", "uc-creator", "dom-research", "port-research-store", "adp-sqlite")
	if err := s.store.UpsertChannel(ctx, ch); err != nil {
		return research.Channel{}, err
	}
	return ch, nil
}

func (s *CreatorService) EditChannel(ctx context.Context, cmd ports.EditChannelCmd) (research.Channel, error) {
	ch, err := s.store.GetChannel(ctx, cmd.ID)
	if err != nil {
		return research.Channel{}, err
	}
	if cmd.Title != "" {
		ch.Title = cmd.Title
	}
	if cmd.Notes != "" || cmd.Notes == "" && cmd.Title != "" {
		ch.Notes = cmd.Notes
	}
	if cmd.Handle != "" {
		ch.Handle = normalizeHandle(cmd.Handle)
	}
	if cmd.GroupID != "" {
		ch.GroupID = cmd.GroupID
	}
	ch.UpdatedAt = time.Now().UTC()
	s.pulse("edit-channel", "ui-creators", "http-creator", "port-creator", "uc-creator", "dom-research", "port-research-store", "adp-sqlite")
	if err := s.store.UpsertChannel(ctx, ch); err != nil {
		return research.Channel{}, err
	}
	return ch, nil
}

func (s *CreatorService) ValidateChannel(ctx context.Context, id string) (research.Channel, error) {
	ch, err := s.store.GetChannel(ctx, id)
	if err != nil {
		return research.Channel{}, err
	}
	s.pulse("validate-channel", "ui-creators", "http-creator", "port-creator", "uc-creator", "dom-research", "port-video-source", "adp-youtube", "ext-youtube")
	remote, err := s.source.LookupChannel(ctx, ch.Handle)
	if err != nil {
		return research.Channel{}, err
	}
	ch.YouTubeID = remote.YouTubeID
	ch.Title = remote.Title
	ch.Description = remote.Description
	ch.ThumbnailURL = remote.ThumbnailURL
	ch.Handle = remote.Handle
	ch.Validated = true
	ch.UpdatedAt = time.Now().UTC()
	s.tracePersist("validate-persist")
	if err := s.store.UpsertChannel(ctx, ch); err != nil {
		return research.Channel{}, err
	}
	return ch, nil
}

func (s *CreatorService) tracePersist(label string) {
	if s.trace == nil {
		return
	}
	s.trace.Emit(ports.TraceEvent{
		Path:  "creator",
		Kind:  ports.TracePersist,
		Label: label,
		Nodes: []string{"uc-creator", "port-research-store", "adp-sqlite"},
	})
}

func (s *CreatorService) HideChannel(ctx context.Context, id string, hidden bool) (research.Channel, error) {
	ch, err := s.store.GetChannel(ctx, id)
	if err != nil {
		return research.Channel{}, err
	}
	ch.Hidden = hidden
	ch.UpdatedAt = time.Now().UTC()
	s.pulse("hide-channel", "ui-creators", "http-creator", "port-creator", "uc-creator", "dom-research", "port-research-store", "adp-sqlite")
	if err := s.store.UpsertChannel(ctx, ch); err != nil {
		return research.Channel{}, err
	}
	return ch, nil
}

func (s *CreatorService) MoveChannel(ctx context.Context, id, groupID string) (research.Channel, error) {
	if _, err := s.store.GetGroup(ctx, groupID); err != nil {
		return research.Channel{}, err
	}
	ch, err := s.store.GetChannel(ctx, id)
	if err != nil {
		return research.Channel{}, err
	}
	ch.GroupID = groupID
	ch.UpdatedAt = time.Now().UTC()
	s.pulse("move-channel", "ui-creators", "http-creator", "port-creator", "uc-creator", "dom-research", "port-research-store", "adp-sqlite")
	if err := s.store.UpsertChannel(ctx, ch); err != nil {
		return research.Channel{}, err
	}
	return ch, nil
}

func normalizeHandle(h string) string {
	h = strings.TrimSpace(h)
	h = strings.TrimPrefix(h, "https://www.youtube.com/")
	h = strings.TrimPrefix(h, "https://youtube.com/")
	h = strings.TrimPrefix(h, "@")
	return strings.TrimSpace(h)
}
