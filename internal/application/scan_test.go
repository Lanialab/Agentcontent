package application

import (
	"context"
	"testing"
	"time"

	"github.com/grokbot-2/agentcontent/internal/domain/research"
	"github.com/grokbot-2/agentcontent/internal/ports"
)

type memResearch struct {
	groups   []research.Group
	channels []research.Channel
	videos   []research.Video
}

func (m *memResearch) ListGroups(context.Context) ([]research.Group, error) { return m.groups, nil }
func (m *memResearch) CreateGroup(_ context.Context, g research.Group) error {
	m.groups = append(m.groups, g)
	return nil
}
func (m *memResearch) GetGroup(_ context.Context, id string) (research.Group, error) {
	for _, g := range m.groups {
		if g.ID == id {
			return g, nil
		}
	}
	return research.Group{}, errNotFound
}
func (m *memResearch) ListChannels(_ context.Context, includeHidden bool) ([]research.Channel, error) {
	out := []research.Channel{}
	for _, c := range m.channels {
		if c.Hidden && !includeHidden {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}
func (m *memResearch) GetChannel(_ context.Context, id string) (research.Channel, error) {
	for _, c := range m.channels {
		if c.ID == id {
			return c, nil
		}
	}
	return research.Channel{}, errNotFound
}
func (m *memResearch) UpsertChannel(_ context.Context, c research.Channel) error {
	for i, x := range m.channels {
		if x.ID == c.ID {
			m.channels[i] = c
			return nil
		}
	}
	m.channels = append(m.channels, c)
	return nil
}
func (m *memResearch) ListVideosByChannel(_ context.Context, channelID string) ([]research.Video, error) {
	out := []research.Video{}
	for _, v := range m.videos {
		if v.ChannelID == channelID {
			out = append(out, v)
		}
	}
	return out, nil
}
func (m *memResearch) ListAllVideos(context.Context) ([]research.Video, error) { return m.videos, nil }
func (m *memResearch) ReplaceChannelVideos(_ context.Context, channelID string, videos []research.Video) error {
	kept := m.videos[:0]
	for _, v := range m.videos {
		if v.ChannelID != channelID {
			kept = append(kept, v)
		}
	}
	m.videos = append(kept, videos...)
	return nil
}
func (m *memResearch) IsEmpty(context.Context) (bool, error) {
	return len(m.channels) == 0, nil
}

type memJobs struct{ held bool }

func (j *memJobs) TryAcquireScanLease(context.Context, string, string) (bool, error) {
	if j.held {
		return false, nil
	}
	j.held = true
	return true, nil
}
func (j *memJobs) ReleaseScanLease(context.Context, string, string) error {
	j.held = false
	return nil
}

type stubSource struct{}

func (stubSource) Stub() bool { return true }
func (stubSource) LookupChannel(context.Context, string) (ports.RemoteChannel, error) {
	return ports.RemoteChannel{YouTubeID: "UC_demo", Handle: "demo", Title: "Demo Channel"}, nil
}
func (stubSource) ListRecentVideos(context.Context, string) ([]ports.RemoteVideo, error) {
	now := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	return []ports.RemoteVideo{
		{YouTubeID: "v1", Title: "Baseline 1", PublishedAt: now.Add(-10 * 24 * time.Hour).Format(time.RFC3339), ViewCount: 1000},
		{YouTubeID: "v2", Title: "Baseline 2", PublishedAt: now.Add(-8 * 24 * time.Hour).Format(time.RFC3339), ViewCount: 800},
		{YouTubeID: "v3", Title: "Baseline 3", PublishedAt: now.Add(-6 * 24 * time.Hour).Format(time.RFC3339), ViewCount: 900},
		{YouTubeID: "viral", Title: "I Tested 7 Hooks — 1 Went 4x?", PublishedAt: now.Add(-2 * 24 * time.Hour).Format(time.RFC3339), ViewCount: 2400},
	}, nil
}

type noopSync struct{}

func (noopSync) Sync(context.Context) error { return nil }
func (noopSync) Enabled() bool              { return false }

var errNotFound = errString("not found")

type errString string

func (e errString) Error() string { return string(e) }

func TestScanThenFeedMarksOutlier(t *testing.T) {
	store := &memResearch{
		groups: []research.Group{{ID: "g1", Name: "Demo"}},
		channels: []research.Channel{{
			ID:        "c1",
			GroupID:   "g1",
			Handle:    "demo",
			Title:     "Demo",
			YouTubeID: "UC_demo",
		}},
	}
	scan := NewScanService(store, &memJobs{}, stubSource{}, noopSync{}, nil)
	res, err := scan.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Videos != 4 {
		t.Fatalf("videos = %d, want 4", res.Videos)
	}
	feed := NewFeedService(store, 2.5, nil)
	items, err := feed.ListFeed(context.Background(), ports.FeedFilter{Outliers: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("outliers = %d, want 1: %+v", len(items), items)
	}
	if items[0].YouTubeID != "viral" {
		t.Fatalf("got %s, want viral", items[0].YouTubeID)
	}
	if items[0].Score < 2.5 {
		t.Fatalf("score %v want >= 2.5", items[0].Score)
	}
}

func TestScanLeaseSkip(t *testing.T) {
	store := &memResearch{channels: []research.Channel{{ID: "c1"}}}
	jobs := &memJobs{held: true}
	scan := NewScanService(store, jobs, stubSource{}, noopSync{}, nil)
	res, err := scan.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !res.Skipped {
		t.Fatal("expected skipped when lease held")
	}
}
