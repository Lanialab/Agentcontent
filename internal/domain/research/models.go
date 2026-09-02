package research

import "time"

type Group struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
}

type Channel struct {
	ID           string    `json:"id"`
	GroupID      string    `json:"groupId"`
	YouTubeID    string    `json:"youtubeId"`
	Handle       string    `json:"handle"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	Hidden       bool      `json:"hidden"`
	Validated    bool      `json:"validated"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Video struct {
	ID              string    `json:"id"`
	ChannelID       string    `json:"channelId"`
	YouTubeID       string    `json:"youtubeId"`
	Title           string    `json:"title"`
	PublishedAt     time.Time `json:"publishedAt"`
	ViewCount       int64     `json:"viewCount"`
	ViewsPerDay     float64   `json:"viewsPerDay"`
	ThumbnailURL    string    `json:"thumbnailUrl"`
	DurationSeconds int       `json:"durationSeconds"`
	CreatedAt       time.Time `json:"createdAt"`
}

func ComputeViewsPerDay(views int64, published, now time.Time) float64 {
	days := now.Sub(published).Hours() / 24
	if days < 1 {
		days = 1
	}
	return float64(views) / days
}
