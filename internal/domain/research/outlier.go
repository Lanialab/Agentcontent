package research

import "math"

// DefaultMultiplier is the product default. The architecture diagram annotates
// `views/day + avg(channel views/day) * 100` as a visual exaggeration used in
// early seed sketches so outliers pop on a dense map. Runtime scoring uses a
// configurable 2–3x multiplier (default 2.5x) so seed data still surfaces
// outliers without a 100x threshold.
const DefaultMultiplier = 2.5

// Five feed signals, matching the Outlier Feed contract.
const (
	SignalOutlierScore = "outlier_score"
	SignalVelocity     = "velocity"
	SignalRecency      = "recency"
	SignalBaselineGap  = "baseline_gap"
	SignalHook         = "hook"
)

type Signals struct {
	OutlierScore float64 `json:"outlierScore"`
	Velocity     float64 `json:"velocity"`
	RecencyHours float64 `json:"recencyHours"`
	BaselineGap  float64 `json:"baselineGap"`
	Hook         float64 `json:"hook"`
}

type ScoredVideo struct {
	Video
	ChannelTitle  string  `json:"channelTitle"`
	ChannelHandle string  `json:"channelHandle"`
	GroupName     string  `json:"groupName"`
	ChannelAvg    float64 `json:"channelAvgViewsPerDay"`
	Score         float64 `json:"score"`
	Outlier       bool    `json:"outlier"`
	Signals       Signals `json:"signals"`
}

// Score compares a video's viewsPerDay against the average viewsPerDay of
// recent same-channel videos (the candidate itself is excluded from the baseline).
func Score(videoVPD, channelAvgVPD float64) float64 {
	if channelAvgVPD <= 0 || math.IsNaN(channelAvgVPD) {
		return 0
	}
	return videoVPD / channelAvgVPD
}

func IsOutlier(score, multiplier float64) bool {
	if multiplier <= 0 {
		multiplier = DefaultMultiplier
	}
	return score >= multiplier
}

// ChannelBaseline is the mean viewsPerDay of peers. excludeID is omitted so a
// viral video is not compared against a baseline that already includes itself.
func ChannelBaseline(videos []Video, excludeID string) float64 {
	var sum float64
	var n int
	for _, v := range videos {
		if v.ID == excludeID {
			continue
		}
		sum += v.ViewsPerDay
		n++
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func HookScore(title string) float64 {
	score := 0.0
	if len(title) >= 24 && len(title) <= 72 {
		score += 0.4
	}
	for _, r := range title {
		if r == '?' || r == '？' {
			score += 0.25
			break
		}
	}
	hasDigit := false
	for _, r := range title {
		if r >= '0' && r <= '9' {
			hasDigit = true
			break
		}
	}
	if hasDigit {
		score += 0.2
	}
	upper := 0
	letters := 0
	for _, r := range title {
		if r >= 'A' && r <= 'Z' {
			upper++
			letters++
		} else if r >= 'a' && r <= 'z' {
			letters++
		}
	}
	if letters > 0 && float64(upper)/float64(letters) > 0.3 {
		score += 0.15
	}
	if score > 1 {
		return 1
	}
	return score
}
