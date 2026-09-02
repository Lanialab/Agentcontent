package script

import "time"

type Mode string

const (
	ModeNhanh Mode = "nhanh"
	ModeAuto  Mode = "auto"
	ModeSau   Mode = "sau"
)

func ParseMode(s string) Mode {
	switch Mode(s) {
	case ModeNhanh, ModeAuto, ModeSau:
		return Mode(s)
	default:
		return ModeNhanh
	}
}

type Script struct {
	ID        string    `json:"id"`
	Mode      Mode      `json:"mode"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func OutlineFor(mode Mode, topic string) string {
	switch mode {
	case ModeNhanh:
		return "# Nhanh — " + topic + "\n\n## Hook (0–15s)\nCâu hỏi + số liệu outlier.\n\n## Promise\nBạn sẽ có 1 framework mang về dùng ngay.\n\n## 3 beats\n1. Quan sát\n2. Sai lầm\n3. Công thức\n\n## CTA\nComment format bạn muốn remix tuần sau."
	case ModeAuto:
		return "# Auto — " + topic + "\n\n## Cold open\nCắt 3 shot outlier + text overlay.\n\n## A-roll\nGiải thích baseline vs spike (views/day).\n\n## B-roll list\n- Dashboard feed\n- Channel list\n- Script editor\n\n## End screen\nBrand pillar + next video."
	case ModeSau:
		return "# Sâu — " + topic + "\n\n## Thesis\nOutlier không phải luck — là lệch baseline của chính kênh đó.\n\n## Research\n- Lấy 8 video gần nhất\n- Loại video đang xét khỏi average\n- So sánh multiplier 2.5x (không dùng *100 của diagram seed)\n\n## Case\nPhân tích 1 video thắng và 1 video chết.\n\n## Playbook\nÁp vào Brand Blueprint / pillars.\n\n## Export\nMarkdown → HTML → PDF."
	default:
		var _ = mode
		return "# Draft — " + topic
	}
}
