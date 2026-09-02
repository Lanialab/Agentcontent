package fileexport

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

type Adapter struct {
	dir string
}

func New(dir string) *Adapter {
	if dir == "" {
		dir = "./data/exports"
	}
	return &Adapter{dir: dir}
}

func (a *Adapter) Write(_ context.Context, name, format, body string) (string, error) {
	if err := os.MkdirAll(a.dir, 0o755); err != nil {
		return "", err
	}
	format = strings.ToLower(strings.TrimSpace(format))
	slug := slugify(name)
	var filename, content string
	switch format {
	case "html":
		filename = slug + ".html"
		content = "<!doctype html><meta charset=\"utf-8\"><pre>" + htmlEscape(body) + "</pre>"
	case "pdf":
		// Pure-Go MVP: write a text stand-in. Real PDF rendering is a non-goal.
		filename = slug + ".pdf.txt"
		content = "PDF export stub\n\n" + body
	default:
		filename = slug + ".md"
		content = body
	}
	path := filepath.Join(a.dir, filename)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func slugify(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "script"
	}
	s = strings.ReplaceAll(s, "/", "-")
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

func (a *Adapter) Dir() string { return a.dir }
