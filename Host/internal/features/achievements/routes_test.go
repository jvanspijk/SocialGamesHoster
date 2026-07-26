package achievements

import "testing"

func TestSpoilerAchievementVisibilityFollowsGameCompletion(t *testing.T) {
	for _, status := range []string{"running", "paused"} {
		if awardVisibleDuringStatus(true, status) {
			t.Fatalf("hidden achievement became visible while game status was %q", status)
		}
	}
	for _, status := range []string{"review", "archived"} {
		if !awardVisibleDuringStatus(true, status) {
			t.Fatalf("hidden achievement remained concealed after game status became %q", status)
		}
	}
	if !awardVisibleDuringStatus(false, "running") {
		t.Fatal("ordinary achievement must be visible during a running game")
	}
}
