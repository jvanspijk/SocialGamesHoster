package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

type rateWindow struct {
	start time.Time
	count int
}

type rateRule struct {
	name   string
	limit  int
	window time.Duration
}

var requestRates = struct {
	sync.Mutex
	values map[string]rateWindow
}{values: map[string]rateWindow{}}

func SecurityMiddleware(event *core.RequestEvent) error {
	traceID := newTraceID()
	event.Set(TraceIDKey, traceID)
	event.Response.Header().Set("X-Request-ID", traceID)
	// The static SPA supplies its build-specific script hashes in a CSP meta tag.
	// This response policy adds protections that meta-delivered CSP cannot express
	// without blocking SvelteKit's generated bootstrap before its hash is known.
	event.Response.Header().Set("Content-Security-Policy", "object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
	event.Response.Header().Set("X-Content-Type-Options", "nosniff")
	event.Response.Header().Set("X-Frame-Options", "DENY")
	event.Response.Header().Set("Referrer-Policy", "no-referrer")
	event.Response.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	event.Response.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

	path := event.Request.URL.Path
	if strings.HasPrefix(path, "/_/") || strings.HasPrefix(path, "/api/collections/") || strings.HasPrefix(path, "/api/batch") {
		return event.NotFoundError("Not found.", nil)
	}

	if strings.HasPrefix(path, "/api/app/v1/") {
		if !validHost(event.Request.Host) {
			return WriteError(event, result.AppError{
				Code: "request.invalid_host", Message: "The requested host address is not allowed.", Status: http.StatusBadRequest,
			})
		}
		limitRequestBody(event)
		if isStateChanging(event.Request.Method) && !validOrigin(event.Request) {
			return WriteError(event, result.Forbidden("request.invalid_origin", "The request origin is not allowed."))
		}
		if rule, ok := rateRuleFor(event.Request); ok && !allowRequest(event, rule, time.Now()) {
			event.Response.Header().Set("Retry-After", strconv.Itoa(max(1, int(rule.window.Seconds()))))
			return WriteError(event, result.AppError{
				Code: "request.rate_limited", Message: "Too many attempts. Wait a moment and try again.", Status: http.StatusTooManyRequests,
			})
		}
	}
	return event.Next()
}

func StaticCacheMiddleware(event *core.RequestEvent) error {
	path := event.Request.URL.Path
	if path == "/" || !strings.Contains(path, ".") {
		event.Response.Header().Set("Cache-Control", "no-cache")
	} else if strings.HasPrefix(path, "/_app/immutable/") {
		event.Response.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		event.Response.Header().Set("Cache-Control", "public, max-age=3600")
	}
	return event.Next()
}

func validOrigin(request *http.Request) bool {
	origin := request.Header.Get("Origin")
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(parsed.Host, request.Host) && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func validHost(value string) bool {
	host := value
	if parsed, _, err := net.SplitHostPort(value); err == nil {
		host = parsed
	} else if strings.Count(value, ":") == 1 {
		host = strings.SplitN(value, ":", 2)[0]
	}
	host = strings.TrimSuffix(strings.Trim(host, "[]"), ".")
	localHostname, _ := os.Hostname()
	return hostAllowed(host, localHostname, localInterfaceAddresses())
}

func hostAllowed(host, localHostname string, localAddresses []net.IP) bool {
	if strings.EqualFold(host, "localhost") ||
		(localHostname != "" && strings.EqualFold(host, strings.TrimSuffix(localHostname, "."))) {
		return true
	}
	ipText := host
	if zoneIndex := strings.LastIndex(ipText, "%"); zoneIndex >= 0 {
		ipText = ipText[:zoneIndex]
	}
	ip := net.ParseIP(ipText)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() {
		return true
	}
	if !ip.IsLinkLocalUnicast() && !isCarrierGradeNAT(ip) {
		return false
	}
	for _, local := range localAddresses {
		if local.Equal(ip) {
			return true
		}
	}
	return false
}

func localInterfaceAddresses() []net.IP {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	result := make([]net.IP, 0, len(addresses))
	for _, address := range addresses {
		switch value := address.(type) {
		case *net.IPNet:
			result = append(result, value.IP)
		case *net.IPAddr:
			result = append(result, value.IP)
		}
	}
	return result
}

func isCarrierGradeNAT(ip net.IP) bool {
	ipv4 := ip.To4()
	return ipv4 != nil && ipv4[0] == 100 && ipv4[1]&0b11000000 == 0b01000000
}

func limitRequestBody(event *core.RequestEvent) {
	if event.Request.Body == nil {
		return
	}
	limit := int64(2 << 20)
	path := event.Request.URL.Path
	switch {
	case path == "/api/app/v1/rulesets/import":
		limit = 26 << 20
	case strings.Contains(path, "/assets"), strings.HasSuffix(path, "/avatar"):
		limit = 6 << 20
	}
	event.Request.Body = http.MaxBytesReader(event.Response, event.Request.Body, limit)
}

func rateRuleFor(request *http.Request) (rateRule, bool) {
	if !isStateChanging(request.Method) {
		return rateRule{}, false
	}
	path := request.URL.Path
	switch {
	case path == "/api/app/v1/setup/owner",
		path == "/api/app/v1/auth/game-master/login",
		path == "/api/app/v1/auth/player/requests",
		strings.Contains(path, "/profile-requests/"):
		return rateRule{name: "identity", limit: 8, window: time.Minute}, true
	case strings.HasSuffix(path, "/join"):
		return rateRule{name: "join", limit: 20, window: time.Minute}, true
	case strings.Contains(path, "/rooms/") && strings.HasSuffix(path, "/messages"):
		return rateRule{name: "message", limit: 30, window: 10 * time.Second}, true
	default:
		return rateRule{}, false
	}
}

func allowRequest(event *core.RequestEvent, rule rateRule, now time.Time) bool {
	host, _, err := net.SplitHostPort(event.Request.RemoteAddr)
	if err != nil {
		host = event.Request.RemoteAddr
	}
	collection := ""
	actorID := ""
	if event.Auth != nil {
		// A LAN party commonly places every player behind one address. Once the
		// caller is authenticated, limit that identity instead of penalising the
		// whole party for one player's traffic.
		collection = event.Auth.Collection().Name
		actorID = event.Auth.Id
	}
	key := rule.name + ":" + rateLimitIdentity(host, collection, actorID)
	requestRates.Lock()
	defer requestRates.Unlock()
	window, allowed := advanceRateWindow(requestRates.values[key], rule, now)
	if allowed {
		requestRates.values[key] = window
	}
	if len(requestRates.values) > 4096 {
		for key, window := range requestRates.values {
			if now.Sub(window.start) > 10*time.Minute {
				delete(requestRates.values, key)
			}
		}
	}
	return allowed
}

func rateLimitIdentity(host, collection, actorID string) string {
	if collection != "" && actorID != "" {
		return "actor:" + collection + ":" + actorID
	}
	return "ip:" + host
}

func advanceRateWindow(window rateWindow, rule rateRule, now time.Time) (rateWindow, bool) {
	if window.start.IsZero() || now.Sub(window.start) >= rule.window {
		window = rateWindow{start: now}
	}
	if window.count >= rule.limit {
		return window, false
	}
	window.count++
	return window, true
}

func isStateChanging(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete
}

func newTraceID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "trace-unavailable"
	}
	return hex.EncodeToString(value[:])
}
