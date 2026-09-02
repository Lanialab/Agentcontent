package application

import (
	"context"
	"strings"
	"time"

	"github.com/grokbot-2/agentcontent/internal/domain/research"
	"github.com/grokbot-2/agentcontent/internal/ports"
)

type FeedService struct {
	store      ports.ResearchStore
	multiplier float64
	trace      ports.Tracer
}

func NewFeedService(store ports.ResearchStore, multiplier float64, trace ports.Tracer) *FeedService {
	if multiplier <= 0 {
		multiplier = research.DefaultMultiplier
	}
	return &FeedService{store: store, multiplier: multiplier, trace: trace}
}

func (s *FeedService) ListFeed(ctx context.Context, f ports.FeedFilter) ([]research.ScoredVideo, error) {
	if s.trace != nil {
		s.trace.Emit(ports.TraceEvent{
			Path:  "feed",
			Kind:  ports.TraceCommand,
			Label: "list-feed",
			Nodes: []string{"ui-feed", "http-feed", "port-feed", "uc-feed", "dom-research", "port-research-store", "adp-sqlite"},
		})
	}
	channels, err := s.store.ListChannels(ctx, f.Hidden)
	if err != nil {
		return nil, err
	}
	groups, err := s.store.ListGroups(ctx)
	if err != nil {
		return nil, err
	}
	groupName := map[string]string{}
	for _, g := range groups {
		groupName[g.ID] = g.Name
	}
	chByID := map[string]research.Channel{}
	for _, c := range channels {
		chByID[c.ID] = c
	}
	videos, err := s.store.ListAllVideos(ctx)
	if err != nil {
		return nil, err
	}
	byCh := map[string][]research.Video{}
	for _, v := range videos {
		byCh[v.ChannelID] = append(byCh[v.ChannelID], v)
	}
	now := time.Now().UTC()
	out := make([]research.ScoredVideo, 0, len(videos))
	q := strings.ToLower(strings.TrimSpace(f.Query))
	for _, v := range videos {
		ch, ok := chByID[v.ChannelID]
		if !ok {
			continue
		}
		if f.GroupID != "" && ch.GroupID != f.GroupID {
			continue
		}
		if q != "" {
			blob := strings.ToLower(v.Title + " " + ch.Title + " " + ch.Handle)
			if !strings.Contains(blob, q) {
				continue
			}
		}
		avg := research.ChannelBaseline(byCh[v.ChannelID], v.ID)
		score := research.Score(v.ViewsPerDay, avg)
		sv := research.ScoredVideo{
			Video:         v,
			ChannelTitle:  ch.Title,
			ChannelHandle: ch.Handle,
			GroupName:     groupName[ch.GroupID],
			ChannelAvg:    avg,
			Score:         score,
			Outlier:       research.IsOutlier(score, s.multiplier),
			Signals: research.Signals{
				OutlierScore: score,
				Velocity:     v.ViewsPerDay,
				RecencyHours: now.Sub(v.PublishedAt).Hours(),
				BaselineGap:  v.ViewsPerDay - avg,
				Hook:         research.HookScore(v.Title),
			},
		}
		if f.MinScore > 0 && sv.Score < f.MinScore {
			continue
		}
		if f.Outliers && !sv.Outlier {
			continue
		}
		out = append(out, sv)
	}
	// Rank: outlier score, then velocity.
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Score > out[i].Score || (out[j].Score == out[i].Score && out[j].ViewsPerDay > out[i].ViewsPerDay) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out, nil
}
