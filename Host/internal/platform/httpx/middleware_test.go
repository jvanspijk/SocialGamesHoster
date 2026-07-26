package httpx

import (
	"net"
	"net/http"
	"testing"
	"time"
)

func TestValidHostAllowsOnlyLoopbackAndPrivateAddresses(t *testing.T) {
	t.Parallel()
	tests := map[string]bool{
		"localhost:8090":   true,
		"127.0.0.1:8090":   true,
		"192.168.1.8:8090": true,
		"10.0.0.8":         true,
		"[::1]:8090":       true,
		"8.8.8.8:8090":     false,
		"example.com:8090": false,
		"":                 false,
	}
	for value, expected := range tests {
		if actual := validHost(value); actual != expected {
			t.Errorf("validHost(%q) = %v, want %v", value, actual, expected)
		}
	}
}

func TestHostPolicyAllowsTheHostComputersLANIdentities(t *testing.T) {
	t.Parallel()
	localAddresses := []net.IP{
		net.ParseIP("100.88.1.12"),
		net.ParseIP("169.254.10.20"),
	}
	tests := map[string]bool{
		"PARTY-HOST":     true,
		"party-host":     true,
		"100.88.1.12":    true,
		"100.88.1.13":    false,
		"169.254.10.20":  true,
		"169.254.10.21":  false,
		"203.0.113.8":    false,
		"party-host.lan": false,
	}
	for host, expected := range tests {
		if actual := hostAllowed(host, "PARTY-HOST", localAddresses); actual != expected {
			t.Errorf("hostAllowed(%q) = %v, want %v", host, actual, expected)
		}
	}
}

func TestRateRuleClassification(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"/api/app/v1/setup/owner":                 "identity",
		"/api/app/v1/auth/game-master/login":      "identity",
		"/api/app/v1/auth/player/requests":        "identity",
		"/api/app/v1/games/game123/join":          "join",
		"/api/app/v1/rooms/room123/messages":      "message",
		"/api/app/v1/games/game123/announcements": "",
	}
	for path, expected := range tests {
		request, err := http.NewRequest(http.MethodPost, "http://127.0.0.1"+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		rule, ok := rateRuleFor(request)
		if ok != (expected != "") || rule.name != expected {
			t.Errorf("rateRuleFor(%q) = (%q, %v), want %q", path, rule.name, ok, expected)
		}
	}
}

func TestRateWindowsReset(t *testing.T) {
	t.Parallel()
	now := time.Now()
	rule := rateRule{name: "contract", limit: 2, window: time.Minute}
	window, allowed := advanceRateWindow(rateWindow{}, rule, now)
	if !allowed || window.count != 1 {
		t.Fatalf("first request = (%+v, %v), want count 1 allowed", window, allowed)
	}
	window, allowed = advanceRateWindow(window, rule, now.Add(time.Second))
	if !allowed || window.count != 2 {
		t.Fatalf("second request = (%+v, %v), want count 2 allowed", window, allowed)
	}
	window, allowed = advanceRateWindow(window, rule, now.Add(2*time.Second))
	if allowed || window.count != 2 {
		t.Fatalf("over-limit request = (%+v, %v), want count 2 rejected", window, allowed)
	}
	window, allowed = advanceRateWindow(window, rule, now.Add(time.Minute))
	if !allowed || window.count != 1 || !window.start.Equal(now.Add(time.Minute)) {
		t.Fatalf("new window = (%+v, %v), want fresh count 1 allowed", window, allowed)
	}
}

func TestSignedInPlayersOnOneLANReceiveIndependentRateBudgets(t *testing.T) {
	playerA := rateLimitIdentity("192.168.1.10", "player_profiles", "player-a")
	playerB := rateLimitIdentity("192.168.1.10", "player_profiles", "player-b")
	if playerA == playerB {
		t.Fatal("two signed-in players on one LAN must not share a rate budget")
	}
	if playerA != rateLimitIdentity("192.168.1.25", "player_profiles", "player-a") {
		t.Fatal("one signed-in player must retain one rate budget across connections")
	}
	if rateLimitIdentity("192.168.1.10", "", "") == rateLimitIdentity("192.168.1.11", "", "") {
		t.Fatal("anonymous requests from different addresses must not share a rate budget")
	}
}
