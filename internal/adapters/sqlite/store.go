package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/grokbot-2/agentcontent/internal/domain/brand"
	"github.com/grokbot-2/agentcontent/internal/domain/intelligence"
	"github.com/grokbot-2/agentcontent/internal/domain/research"
	"github.com/grokbot-2/agentcontent/internal/domain/script"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) IsEmpty(ctx context.Context) (bool, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM channels`).Scan(&n); err != nil {
		return false, err
	}
	return n == 0, nil
}

func (s *Store) ListGroups(ctx context.Context) ([]research.Group, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, sort_order, created_at FROM groups ORDER BY sort_order, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []research.Group
	for rows.Next() {
		var g research.Group
		var created string
		if err := rows.Scan(&g.ID, &g.Name, &g.SortOrder, &created); err != nil {
			return nil, err
		}
		g.CreatedAt = parseTS(created)
		out = append(out, g)
	}
	if out == nil {
		out = []research.Group{}
	}
	return out, rows.Err()
}

func (s *Store) CreateGroup(ctx context.Context, g research.Group) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO groups(id, name, sort_order, created_at) VALUES (?,?,?,?)`,
		g.ID, g.Name, g.SortOrder, g.CreatedAt.UTC().Format(time.RFC3339))
	return err
}

func (s *Store) GetGroup(ctx context.Context, id string) (research.Group, error) {
	var g research.Group
	var created string
	err := s.db.QueryRowContext(ctx, `SELECT id, name, sort_order, created_at FROM groups WHERE id = ?`, id).
		Scan(&g.ID, &g.Name, &g.SortOrder, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return g, errors.New("group not found")
	}
	g.CreatedAt = parseTS(created)
	return g, err
}

func (s *Store) ListChannels(ctx context.Context, includeHidden bool) ([]research.Channel, error) {
	q := `SELECT id, group_id, youtube_id, handle, title, description, thumbnail_url, hidden, validated, notes, created_at, updated_at FROM channels`
	if !includeHidden {
		q += ` WHERE hidden = 0`
	}
	q += ` ORDER BY title`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []research.Channel
	for rows.Next() {
		c, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if out == nil {
		out = []research.Channel{}
	}
	return out, rows.Err()
}

func (s *Store) GetChannel(ctx context.Context, id string) (research.Channel, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, group_id, youtube_id, handle, title, description, thumbnail_url, hidden, validated, notes, created_at, updated_at FROM channels WHERE id = ?`, id)
	c, err := scanChannel(row)
	if errors.Is(err, sql.ErrNoRows) {
		return research.Channel{}, errors.New("channel not found")
	}
	return c, err
}

func (s *Store) UpsertChannel(ctx context.Context, c research.Channel) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO channels(id, group_id, youtube_id, handle, title, description, thumbnail_url, hidden, validated, notes, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
  group_id=excluded.group_id,
  youtube_id=excluded.youtube_id,
  handle=excluded.handle,
  title=excluded.title,
  description=excluded.description,
  thumbnail_url=excluded.thumbnail_url,
  hidden=excluded.hidden,
  validated=excluded.validated,
  notes=excluded.notes,
  updated_at=excluded.updated_at
`, c.ID, c.GroupID, c.YouTubeID, c.Handle, c.Title, c.Description, c.ThumbnailURL, boolInt(c.Hidden), boolInt(c.Validated), c.Notes,
		c.CreatedAt.UTC().Format(time.RFC3339), c.UpdatedAt.UTC().Format(time.RFC3339))
	return err
}

func (s *Store) ListVideosByChannel(ctx context.Context, channelID string) ([]research.Video, error) {
	rows, err := s.db.QueryContext(ctx, videoSelect+` WHERE channel_id = ? ORDER BY published_at DESC`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectVideos(rows)
}

func (s *Store) ListAllVideos(ctx context.Context) ([]research.Video, error) {
	rows, err := s.db.QueryContext(ctx, videoSelect+` ORDER BY published_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectVideos(rows)
}

func (s *Store) ReplaceChannelVideos(ctx context.Context, channelID string, videos []research.Video) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM videos WHERE channel_id = ?`, channelID); err != nil {
		return err
	}
	for _, v := range videos {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO videos(id, channel_id, youtube_id, title, published_at, view_count, views_per_day, thumbnail_url, duration_seconds, created_at)
VALUES (?,?,?,?,?,?,?,?,?,?)`,
			v.ID, v.ChannelID, v.YouTubeID, v.Title, v.PublishedAt.UTC().Format(time.RFC3339),
			v.ViewCount, v.ViewsPerDay, v.ThumbnailURL, v.DurationSeconds, v.CreatedAt.UTC().Format(time.RFC3339)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListIdeas(ctx context.Context) ([]intelligence.Idea, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, prompt, body, mentions, template_id, created_at FROM ideas ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []intelligence.Idea
	for rows.Next() {
		var it intelligence.Idea
		var mentions, created string
		if err := rows.Scan(&it.ID, &it.Title, &it.Prompt, &it.Body, &mentions, &it.TemplateID, &created); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(mentions), &it.Mentions)
		if it.Mentions == nil {
			it.Mentions = []string{}
		}
		it.CreatedAt = parseTS(created)
		out = append(out, it)
	}
	if out == nil {
		out = []intelligence.Idea{}
	}
	return out, rows.Err()
}

func (s *Store) SaveIdea(ctx context.Context, idea intelligence.Idea) error {
	b, _ := json.Marshal(idea.Mentions)
	_, err := s.db.ExecContext(ctx, `INSERT INTO ideas(id, title, prompt, body, mentions, template_id, created_at) VALUES (?,?,?,?,?,?,?)`,
		idea.ID, idea.Title, idea.Prompt, idea.Body, string(b), idea.TemplateID, idea.CreatedAt.UTC().Format(time.RFC3339))
	return err
}

func (s *Store) ListChat(ctx context.Context) ([]intelligence.ChatMessage, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, role, content, created_at FROM chat_messages ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []intelligence.ChatMessage
	for rows.Next() {
		var m intelligence.ChatMessage
		var created string
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &created); err != nil {
			return nil, err
		}
		m.CreatedAt = parseTS(created)
		out = append(out, m)
	}
	if out == nil {
		out = []intelligence.ChatMessage{}
	}
	return out, rows.Err()
}

func (s *Store) SaveChat(ctx context.Context, msg intelligence.ChatMessage) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO chat_messages(id, role, content, created_at) VALUES (?,?,?,?)`,
		msg.ID, msg.Role, msg.Content, msg.CreatedAt.UTC().Format(time.RFC3339))
	return err
}

func (s *Store) GetBlueprint(ctx context.Context) (brand.Blueprint, error) {
	var bp brand.Blueprint
	var pillars, topics, updated string
	err := s.db.QueryRowContext(ctx, `SELECT love, good_at, world_needs, paid_for, positioning, voice, pillars, topics, updated_at FROM brand_blueprint WHERE id = 'default'`).
		Scan(&bp.Love, &bp.GoodAt, &bp.WorldNeeds, &bp.PaidFor, &bp.Positioning, &bp.Voice, &pillars, &topics, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return brand.Blueprint{Pillars: []string{}, Topics: []string{}}, nil
	}
	if err != nil {
		return bp, err
	}
	_ = json.Unmarshal([]byte(pillars), &bp.Pillars)
	_ = json.Unmarshal([]byte(topics), &bp.Topics)
	if bp.Pillars == nil {
		bp.Pillars = []string{}
	}
	if bp.Topics == nil {
		bp.Topics = []string{}
	}
	bp.UpdatedAt = parseTS(updated)
	return bp, nil
}

func (s *Store) SaveBlueprint(ctx context.Context, bp brand.Blueprint) error {
	pillars, _ := json.Marshal(bp.Pillars)
	topics, _ := json.Marshal(bp.Topics)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO brand_blueprint(id, love, good_at, world_needs, paid_for, positioning, voice, pillars, topics, updated_at)
VALUES ('default',?,?,?,?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
  love=excluded.love, good_at=excluded.good_at, world_needs=excluded.world_needs, paid_for=excluded.paid_for,
  positioning=excluded.positioning, voice=excluded.voice, pillars=excluded.pillars, topics=excluded.topics, updated_at=excluded.updated_at
`, bp.Love, bp.GoodAt, bp.WorldNeeds, bp.PaidFor, bp.Positioning, bp.Voice, string(pillars), string(topics), bp.UpdatedAt.UTC().Format(time.RFC3339))
	return err
}

func (s *Store) ListScripts(ctx context.Context) ([]script.Script, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, mode, title, body, status, created_at, updated_at FROM scripts ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []script.Script
	for rows.Next() {
		var it script.Script
		var mode, created, updated string
		if err := rows.Scan(&it.ID, &mode, &it.Title, &it.Body, &it.Status, &created, &updated); err != nil {
			return nil, err
		}
		it.Mode = script.ParseMode(mode)
		it.CreatedAt = parseTS(created)
		it.UpdatedAt = parseTS(updated)
		out = append(out, it)
	}
	if out == nil {
		out = []script.Script{}
	}
	return out, rows.Err()
}

func (s *Store) GetScript(ctx context.Context, id string) (script.Script, error) {
	var it script.Script
	var mode, created, updated string
	err := s.db.QueryRowContext(ctx, `SELECT id, mode, title, body, status, created_at, updated_at FROM scripts WHERE id = ?`, id).
		Scan(&it.ID, &mode, &it.Title, &it.Body, &it.Status, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return it, errors.New("script not found")
	}
	it.Mode = script.ParseMode(mode)
	it.CreatedAt = parseTS(created)
	it.UpdatedAt = parseTS(updated)
	return it, err
}

func (s *Store) SaveScript(ctx context.Context, it script.Script) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO scripts(id, mode, title, body, status, created_at, updated_at) VALUES (?,?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET mode=excluded.mode, title=excluded.title, body=excluded.body, status=excluded.status, updated_at=excluded.updated_at
`, it.ID, string(it.Mode), it.Title, it.Body, it.Status, it.CreatedAt.UTC().Format(time.RFC3339), it.UpdatedAt.UTC().Format(time.RFC3339))
	return err
}

func (s *Store) TryAcquireScanLease(ctx context.Context, owner, until string) (bool, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `
UPDATE jobs SET status='running', owner=?, lease_until=?, started_at=?, error=''
WHERE id='scan' AND (status != 'running' OR lease_until IS NULL OR lease_until < ?)
`, owner, until, now, now)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}

func (s *Store) ReleaseScanLease(ctx context.Context, owner, errMsg string) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE jobs SET status='idle', owner='', lease_until=NULL, finished_at=?, error=?
WHERE id='scan' AND owner=?
`, time.Now().UTC().Format(time.RFC3339), errMsg, owner)
	return err
}

const videoSelect = `SELECT id, channel_id, youtube_id, title, published_at, view_count, views_per_day, thumbnail_url, duration_seconds, created_at FROM videos`

type scanner interface {
	Scan(dest ...any) error
}

func scanChannel(sc scanner) (research.Channel, error) {
	var c research.Channel
	var hidden, validated int
	var created, updated string
	err := sc.Scan(&c.ID, &c.GroupID, &c.YouTubeID, &c.Handle, &c.Title, &c.Description, &c.ThumbnailURL, &hidden, &validated, &c.Notes, &created, &updated)
	c.Hidden = hidden != 0
	c.Validated = validated != 0
	c.CreatedAt = parseTS(created)
	c.UpdatedAt = parseTS(updated)
	return c, err
}

func collectVideos(rows *sql.Rows) ([]research.Video, error) {
	var out []research.Video
	for rows.Next() {
		var v research.Video
		var pub, created string
		if err := rows.Scan(&v.ID, &v.ChannelID, &v.YouTubeID, &v.Title, &pub, &v.ViewCount, &v.ViewsPerDay, &v.ThumbnailURL, &v.DurationSeconds, &created); err != nil {
			return nil, err
		}
		v.PublishedAt = parseTS(pub)
		v.CreatedAt = parseTS(created)
		out = append(out, v)
	}
	if out == nil {
		out = []research.Video{}
	}
	return out, rows.Err()
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func parseTS(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
