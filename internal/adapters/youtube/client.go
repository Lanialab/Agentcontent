package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/grokbot-2/agentcontent/internal/ports"
)

type Adapter struct {
	key    string
	client *http.Client
	stub   *Stub
}

func New(apiKey string) *Adapter {
	return &Adapter{
		key:    strings.TrimSpace(apiKey),
		client: &http.Client{Timeout: 20 * time.Second},
		stub:   NewStub(),
	}
}

func (a *Adapter) Stub() bool { return a.key == "" }

func (a *Adapter) LookupChannel(ctx context.Context, handle string) (ports.RemoteChannel, error) {
	if a.Stub() {
		return a.stub.LookupChannel(ctx, handle)
	}
	handle = strings.TrimPrefix(strings.TrimSpace(handle), "@")
	u := "https://www.googleapis.com/youtube/v3/channels?part=snippet&forHandle=" + url.QueryEscape(handle) + "&key=" + url.QueryEscape(a.key)
	var payload struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				CustomURL   string `json:"customUrl"`
				Thumbnails  struct {
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
				} `json:"thumbnails"`
			} `json:"snippet"`
		} `json:"items"`
	}
	if err := a.get(ctx, u, &payload); err != nil {
		return ports.RemoteChannel{}, err
	}
	if len(payload.Items) == 0 {
		return ports.RemoteChannel{}, fmt.Errorf("youtube: channel %q not found", handle)
	}
	it := payload.Items[0]
	h := strings.TrimPrefix(it.Snippet.CustomURL, "@")
	if h == "" {
		h = handle
	}
	return ports.RemoteChannel{
		YouTubeID:    it.ID,
		Handle:       h,
		Title:        it.Snippet.Title,
		Description:  it.Snippet.Description,
		ThumbnailURL: it.Snippet.Thumbnails.Default.URL,
	}, nil
}

func (a *Adapter) ListRecentVideos(ctx context.Context, youtubeID string) ([]ports.RemoteVideo, error) {
	if a.Stub() {
		return a.stub.ListRecentVideos(ctx, youtubeID)
	}
	searchURL := "https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=" + url.QueryEscape(youtubeID) +
		"&maxResults=8&order=date&type=video&key=" + url.QueryEscape(a.key)
	var search struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
		} `json:"items"`
	}
	if err := a.get(ctx, searchURL, &search); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(search.Items))
	for _, it := range search.Items {
		if it.ID.VideoID != "" {
			ids = append(ids, it.ID.VideoID)
		}
	}
	if len(ids) == 0 {
		return []ports.RemoteVideo{}, nil
	}
	statsURL := "https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&id=" + url.QueryEscape(strings.Join(ids, ",")) + "&key=" + url.QueryEscape(a.key)
	var stats struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title       string `json:"title"`
				PublishedAt string `json:"publishedAt"`
				Thumbnails  struct {
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
				} `json:"thumbnails"`
			} `json:"snippet"`
			Statistics struct {
				ViewCount string `json:"viewCount"`
			} `json:"statistics"`
		} `json:"items"`
	}
	if err := a.get(ctx, statsURL, &stats); err != nil {
		return nil, err
	}
	out := make([]ports.RemoteVideo, 0, len(stats.Items))
	for _, it := range stats.Items {
		var views int64
		fmt.Sscanf(it.Statistics.ViewCount, "%d", &views)
		out = append(out, ports.RemoteVideo{
			YouTubeID:    it.ID,
			Title:        it.Snippet.Title,
			PublishedAt:  it.Snippet.PublishedAt,
			ViewCount:    views,
			ThumbnailURL: it.Snippet.Thumbnails.Default.URL,
		})
	}
	return out, nil
}

func (a *Adapter) get(ctx context.Context, raw string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return err
	}
	res, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("youtube http %d", res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(dest)
}
