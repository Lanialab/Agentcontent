package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/grokbot-2/agentcontent/internal/domain/brand"
	"github.com/grokbot-2/agentcontent/internal/domain/intelligence"
	"github.com/grokbot-2/agentcontent/internal/domain/research"
	"github.com/grokbot-2/agentcontent/internal/domain/script"
	"github.com/grokbot-2/agentcontent/internal/idgen"
)

func SeedIfEmpty(ctx context.Context, db *sql.DB) error {
	store := NewStore(db)
	empty, err := store.IsEmpty(ctx)
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}
	now := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)

	groups := []research.Group{
		{ID: "grp_laptrinh", Name: "Lập trình", SortOrder: 1, CreatedAt: now},
		{ID: "grp_sangtao", Name: "Sáng tạo", SortOrder: 2, CreatedAt: now},
		{ID: "grp_giaoduc", Name: "Giáo dục", SortOrder: 3, CreatedAt: now},
	}
	for _, g := range groups {
		if err := store.CreateGroup(ctx, g); err != nil {
			return err
		}
	}

	channels := []research.Channel{
		ch("ch_f8", "grp_laptrinh", "UCF8", "f8official", "F8 Official", "Học lập trình để đi làm.", now, true, false),
		ch("ch_webvn", "grp_laptrinh", "UCWEB", "webdevvn", "Web Dev VN", "Frontend thực chiến.", now, true, false),
		ch("ch_dino", "grp_sangtao", "UCDINO", "dinostudio", "Dino Studio", "Motion + storytelling.", now, true, false),
		ch("ch_writer", "grp_giaoduc", "UCWR", "presentwriter", "The Present Writer", "Viết & tư duy.", now, true, false),
		ch("ch_hidden", "grp_sangtao", "UCHIDE", "oldvault", "Kho cũ", "Ẩn khỏi feed.", now, true, true),
	}
	for _, c := range channels {
		if err := store.UpsertChannel(ctx, c); err != nil {
			return err
		}
	}

	type spec struct {
		chID  string
		title string
		days  int
		views int64
		yt    string
	}
	// Same-channel baselines around 100–150k views / ~10 days ≈ 10–15k vpd.
	// One video per active channel is 2.8–4x so default 2.5x multiplier flags them.
	// Diagram *100 is documented in README — not used as the product threshold.
	videos := []spec{
		{"ch_f8", "JS cơ bản — vòng lặp", 18, 180_000, "f8_base_1"},
		{"ch_f8", "React trong 1 giờ", 14, 160_000, "f8_base_2"},
		{"ch_f8", "SQL cho người mới", 11, 140_000, "f8_base_3"},
		{"ch_f8", "Tôi thử 7 hook — 1 cái đi 4x?", 4, 220_000, "f8_viral"},

		{"ch_webvn", "CSS Grid layout thực tế", 20, 90_000, "web_base_1"},
		{"ch_webvn", "Debounce vs throttle", 12, 70_000, "web_base_2"},
		{"ch_webvn", "Vite config tối thiểu", 9, 65_000, "web_base_3"},
		{"ch_webvn", "Tại sao outlier feed cần 5 tín hiệu?", 3, 95_000, "web_viral"},

		{"ch_dino", "Storyboard 30s", 16, 40_000, "dino_base_1"},
		{"ch_dino", "Color script cyberpunk", 10, 35_000, "dino_base_2"},
		{"ch_dino", "Particle lines trong After Effects", 7, 32_000, "dino_base_3"},
		{"ch_dino", "Architecture map sống — 1 take", 2, 48_000, "dino_viral"},

		{"ch_writer", "Viết mỗi sáng 20 phút", 21, 55_000, "wr_base_1"},
		{"ch_writer", "Ikigai không phải slogan", 13, 48_000, "wr_base_2"},
		{"ch_writer", "Pillar content 90 ngày", 8, 40_000, "wr_base_3"},
		{"ch_writer", "Brand Blueprint cho creator Việt", 3, 72_000, "wr_viral"},

		{"ch_hidden", "Video kho cũ", 40, 8_000, "hid_1"},
	}

	for _, sp := range videos {
		pub := now.Add(-time.Duration(sp.days) * 24 * time.Hour)
		v := research.Video{
			ID:              idgen.New("vid"),
			ChannelID:       sp.chID,
			YouTubeID:       sp.yt,
			Title:           sp.title,
			PublishedAt:     pub,
			ViewCount:       sp.views,
			ViewsPerDay:     research.ComputeViewsPerDay(sp.views, pub, now),
			DurationSeconds: 480,
			CreatedAt:       now,
		}
		if err := insertVideo(ctx, db, v); err != nil {
			return err
		}
	}

	bp := brand.Blueprint{
		Love:        "Nghiên cứu kênh, bóc tách format thắng, viết kịch bản sạch.",
		GoodAt:      "Hexagonal systems, outlier math, editorial taste.",
		WorldNeeds:  "Creator Việt cần OS riêng — không phải spreadsheet + Apps Script.",
		PaidFor:     "Hệ thống research → idea → script có thể tái sử dụng.",
		Positioning: "AgentContent là personal creator OS: research YouTube, outlier feed, AI ideas, Brand Blueprint, kịch bản Nhanh/Auto/Sâu.",
		Voice:       "Rõ, kỹ thuật, tiếng Việt tự nhiên, không hype rỗng.",
		Pillars:     []string{"Research & outliers", "Intelligence / ideas", "Brand Blueprint", "Kịch bản Video"},
		Topics:      []string{"views/day vs channel baseline", "Ikigai thực chiến", "Hook lab", "Architecture live view"},
		UpdatedAt:   now,
	}
	if err := store.SaveBlueprint(ctx, bp); err != nil {
		return err
	}

	ideas := []intelligence.Idea{
		{
			ID: idgen.New("idea"), Title: "Remix 4x hook của F8", Prompt: "outlier-remix",
			Body:     "Mượn format 'tôi thử N thứ' nhưng áp vào Brand Blueprint: thử 5 pillar, chỉ 1 cái sống.",
			Mentions: []string{"@f8official"}, TemplateID: "outlier-remix", CreatedAt: now,
		},
		{
			ID: idgen.New("idea"), Title: "Ikigai series 5 tập", Prompt: "ikigai-series",
			Body:     "Tập 1 Love, 2 Good at, 3 World needs, 4 Paid for, 5 Intersection + CTA script Sâu.",
			Mentions: []string{"@presentwriter"}, TemplateID: "ikigai-series", CreatedAt: now.Add(-time.Hour),
		},
	}
	for _, idea := range ideas {
		if err := store.SaveIdea(ctx, idea); err != nil {
			return err
		}
	}

	scripts := []script.Script{
		{ID: idgen.New("sc"), Mode: script.ModeNhanh, Title: "Hook 15s — outlier", Body: script.OutlineFor(script.ModeNhanh, "Outlier 2.5x"), Status: "draft", CreatedAt: now, UpdatedAt: now},
		{ID: idgen.New("sc"), Mode: script.ModeAuto, Title: "Auto cut — architecture map", Body: script.OutlineFor(script.ModeAuto, "Architecture live view"), Status: "draft", CreatedAt: now, UpdatedAt: now},
		{ID: idgen.New("sc"), Mode: script.ModeSau, Title: "Sâu — baseline vs spike", Body: script.OutlineFor(script.ModeSau, "viewsPerDay vs channel average"), Status: "draft", CreatedAt: now, UpdatedAt: now},
	}
	for _, sc := range scripts {
		if err := store.SaveScript(ctx, sc); err != nil {
			return err
		}
	}

	_, _ = db.ExecContext(ctx, `INSERT INTO chat_messages(id, role, content, created_at) VALUES (?,?,?,?)`,
		idgen.New("msg"), "assistant", "Chào. Mình là lớp Intelligence của AgentContent. Mention một kênh, chọn template, hoặc bảo mình generate idea — dây kiến trúc sẽ sáng.", now.Format(time.RFC3339))
	return nil
}

func ch(id, gid, yt, handle, title, desc string, now time.Time, validated, hidden bool) research.Channel {
	return research.Channel{
		ID: id, GroupID: gid, YouTubeID: yt, Handle: handle, Title: title, Description: desc,
		Validated: validated, Hidden: hidden, CreatedAt: now, UpdatedAt: now,
	}
}

func insertVideo(ctx context.Context, db *sql.DB, v research.Video) error {
	_, err := db.ExecContext(ctx, `
INSERT INTO videos(id, channel_id, youtube_id, title, published_at, view_count, views_per_day, thumbnail_url, duration_seconds, created_at)
VALUES (?,?,?,?,?,?,?,?,?,?)`,
		v.ID, v.ChannelID, v.YouTubeID, v.Title, v.PublishedAt.UTC().Format(time.RFC3339),
		v.ViewCount, v.ViewsPerDay, v.ThumbnailURL, v.DurationSeconds, v.CreatedAt.UTC().Format(time.RFC3339))
	return err
}
