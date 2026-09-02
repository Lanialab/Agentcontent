package research

import "testing"

func TestScoreAndOutlier(t *testing.T) {
	got := Score(300, 100)
	if got != 3 {
		t.Fatalf("score = %v, want 3", got)
	}
	if !IsOutlier(got, 2.5) {
		t.Fatal("3x should be an outlier at 2.5x")
	}
	if IsOutlier(Score(200, 100), 2.5) {
		t.Fatal("2x should not be an outlier at 2.5x")
	}
	if IsOutlier(Score(249, 100), 2.5) {
		t.Fatal("2.49x should not be an outlier at 2.5x")
	}
	if !IsOutlier(Score(250, 100), 2.5) {
		t.Fatal("2.5x should be an outlier at the default multiplier")
	}
}

func TestScoreZeroBaseline(t *testing.T) {
	if Score(100, 0) != 0 {
		t.Fatal("zero baseline must yield 0")
	}
}

func TestChannelBaselineExcludesCandidate(t *testing.T) {
	videos := []Video{
		{ID: "a", ViewsPerDay: 100},
		{ID: "b", ViewsPerDay: 100},
		{ID: "c", ViewsPerDay: 100},
		{ID: "viral", ViewsPerDay: 400},
	}
	avg := ChannelBaseline(videos, "viral")
	if avg != 100 {
		t.Fatalf("baseline = %v, want 100", avg)
	}
	score := Score(400, avg)
	if score != 4 {
		t.Fatalf("score = %v, want 4", score)
	}
	if !IsOutlier(score, DefaultMultiplier) {
		t.Fatal("4x seed-style outlier should flag at default 2.5x")
	}
}

func TestViewsPerDayMinimumDay(t *testing.T) {
	// Same timestamp → treat as 1 day so we never divide by zero.
	now := mustParse("2026-09-01T12:00:00Z")
	vpd := ComputeViewsPerDay(1000, now, now)
	if vpd != 1000 {
		t.Fatalf("vpd = %v, want 1000", vpd)
	}
}

func TestHookScore(t *testing.T) {
	low := HookScore("hi")
	high := HookScore("I Tried 7 Tools — Did Any Survive?")
	if high <= low {
		t.Fatalf("expected question+number title to score higher: low=%v high=%v", low, high)
	}
}
