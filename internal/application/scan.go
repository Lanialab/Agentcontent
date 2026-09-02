package application

import (
	"context"
	"time"

	"github.com/grokbot-2/agentcontent/internal/domain/research"
	"github.com/grokbot-2/agentcontent/internal/idgen"
	"github.com/grokbot-2/agentcontent/internal/ports"
)

type ScanService struct {
	store  ports.ResearchStore
	jobs   ports.JobStore
	source ports.VideoSourcePort
	syncer ports.SyncExportPort
	trace  ports.Tracer
	owner  string
}

func NewScanService(store ports.ResearchStore, jobs ports.JobStore, source ports.VideoSourcePort, syncer ports.SyncExportPort, trace ports.Tracer) *ScanService {
	return &ScanService{store: store, jobs: jobs, source: source, syncer: syncer, trace: trace, owner: idgen.New("lease")}
}

func (s *ScanService) Run(ctx context.Context) (ports.ScanResult, error) {
	if s.trace != nil {
		s.trace.Emit(ports.TraceEvent{
			Path:  "scan",
			Kind:  ports.TraceAsync,
			Label: "scheduled-scan",
			Nodes: []string{"ui-creators", "http-creator", "cli-job", "port-scan", "uc-scan", "dom-research", "port-video-source", "adp-youtube", "ext-youtube", "port-research-store", "adp-sqlite"},
		})
	}
	until := time.Now().UTC().Add(10 * time.Minute).Format(time.RFC3339)
	ok, err := s.jobs.TryAcquireScanLease(ctx, s.owner, until)
	if err != nil {
		return ports.ScanResult{}, err
	}
	if !ok {
		return ports.ScanResult{Skipped: true}, nil
	}
	defer s.jobs.ReleaseScanLease(ctx, s.owner, "")

	channels, err := s.store.ListChannels(ctx, false)
	if err != nil {
		return ports.ScanResult{}, err
	}
	now := time.Now().UTC()
	videosWritten := 0
	scanned := 0
	for _, ch := range channels {
		ytID := ch.YouTubeID
		if ytID == "" {
			remote, lerr := s.source.LookupChannel(ctx, ch.Handle)
			if lerr != nil {
				continue
			}
			ch.YouTubeID = remote.YouTubeID
			ch.Title = remote.Title
			ch.Description = remote.Description
			ch.ThumbnailURL = remote.ThumbnailURL
			ch.Validated = true
			ch.UpdatedAt = now
			ytID = ch.YouTubeID
			_ = s.store.UpsertChannel(ctx, ch)
		}
		remoteVideos, verr := s.source.ListRecentVideos(ctx, ytID)
		if verr != nil {
			continue
		}
		scanned++
		local := make([]research.Video, 0, len(remoteVideos))
		for _, rv := range remoteVideos {
			pub := parseTime(rv.PublishedAt, now)
			item := research.Video{
				ID:              idgen.New("vid"),
				ChannelID:       ch.ID,
				YouTubeID:       rv.YouTubeID,
				Title:           rv.Title,
				PublishedAt:     pub,
				ViewCount:       rv.ViewCount,
				ViewsPerDay:     research.ComputeViewsPerDay(rv.ViewCount, pub, now),
				ThumbnailURL:    rv.ThumbnailURL,
				DurationSeconds: rv.DurationSeconds,
				CreatedAt:       now,
			}
			local = append(local, item)
		}
		if err := s.store.ReplaceChannelVideos(ctx, ch.ID, local); err != nil {
			return ports.ScanResult{}, err
		}
		videosWritten += len(local)
	}
	if s.syncer != nil && s.syncer.Enabled() {
		_ = s.syncer.Sync(ctx)
	}
	return ports.ScanResult{
		Channels: scanned,
		Videos:   videosWritten,
		Stub:     s.source.Stub(),
	}, nil
}

func parseTime(s string, fallback time.Time) time.Time {
	if s == "" {
		return fallback
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fallback
	}
	return t
}
