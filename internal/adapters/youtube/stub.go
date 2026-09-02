package youtube

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/grokbot-2/agentcontent/internal/ports"
)

// Stub catalogs seed handles so the product demos without a YouTube key.
type Stub struct {
	channels map[string]ports.RemoteChannel
}

func NewStub() *Stub {
	list := []ports.RemoteChannel{
		{YouTubeID: "UCF8", Handle: "f8official", Title: "F8 Official", Description: "Học lập trình để đi làm."},
		{YouTubeID: "UCWEB", Handle: "webdevvn", Title: "Web Dev VN", Description: "Frontend thực chiến."},
		{YouTubeID: "UCDINO", Handle: "dinostudio", Title: "Dino Studio", Description: "Motion + storytelling."},
		{YouTubeID: "UCWR", Handle: "presentwriter", Title: "The Present Writer", Description: "Viết & tư duy."},
		{YouTubeID: "UCHIDE", Handle: "oldvault", Title: "Kho cũ", Description: "Ẩn khỏi feed."},
	}
	m := map[string]ports.RemoteChannel{}
	for _, c := range list {
		m[strings.ToLower(c.Handle)] = c
		m[c.YouTubeID] = c
	}
	return &Stub{channels: m}
}

func (s *Stub) Stub() bool { return true }

func (s *Stub) LookupChannel(_ context.Context, handle string) (ports.RemoteChannel, error) {
	h := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(handle), "@"))
	if c, ok := s.channels[h]; ok {
		return c, nil
	}
	return ports.RemoteChannel{
		YouTubeID:   "UC_" + h,
		Handle:      h,
		Title:       handle,
		Description: "Stub channel (no YOUTUBE_API_KEY).",
	}, nil
}

func (s *Stub) ListRecentVideos(_ context.Context, youtubeID string) ([]ports.RemoteVideo, error) {
	now := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	ch, ok := s.channels[youtubeID]
	name := youtubeID
	if ok {
		name = ch.Title
	}
	return []ports.RemoteVideo{
		vid(youtubeID+"_1", name+" — baseline A", now.Add(-14*24*time.Hour), 80_000),
		vid(youtubeID+"_2", name+" — baseline B", now.Add(-9*24*time.Hour), 70_000),
		vid(youtubeID+"_3", name+" — baseline C", now.Add(-6*24*time.Hour), 75_000),
		vid(youtubeID+"_x", fmt.Sprintf("%s — 3x spike this week?", name), now.Add(-2*24*time.Hour), 90_000),
	}, nil
}

func vid(id, title string, pub time.Time, views int64) ports.RemoteVideo {
	return ports.RemoteVideo{
		YouTubeID:       id,
		Title:           title,
		PublishedAt:     pub.Format(time.RFC3339),
		ViewCount:       views,
		DurationSeconds: 540,
	}
}
