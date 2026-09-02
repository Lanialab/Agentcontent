package intelligence

import "time"

type Idea struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Prompt     string    `json:"prompt"`
	Body       string    `json:"body"`
	Mentions   []string  `json:"mentions"`
	TemplateID string    `json:"templateId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type ChatMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type Template struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
}

func DefaultTemplates() []Template {
	return []Template{
		{
			ID:          "outlier-remix",
			Name:        "Outlier remix",
			Description: "Đảo góc nhìn từ video đang thắng, giữ format.",
			Prompt:      "Dựa trên video outlier sau, đề xuất 3 góc remix giữ format nhưng đổi góc nhìn cho Brand Blueprint của tôi.",
		},
		{
			ID:          "ikigai-series",
			Name:        "Ikigai series",
			Description: "Chuỗi video từ giao điểm Ikigai.",
			Prompt:      "Từ Brand Blueprint (Ikigai + pillars), phác 5 tập series 8–12 phút.",
		},
		{
			ID:          "hook-lab",
			Name:        "Hook lab",
			Description: "10 hook trong 30 giây đầu.",
			Prompt:      "Viết 10 hook mở đầu (Nhanh) cho chủ đề này, mỗi hook ≤ 18 từ.",
		},
	}
}

func ValidateIdea(title, body string) error {
	if title == "" {
		return errEmpty("title")
	}
	if body == "" {
		return errEmpty("body")
	}
	return nil
}

type fieldError string

func errEmpty(field string) error { return fieldError(field + " is required") }

func (e fieldError) Error() string { return string(e) }
