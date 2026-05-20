package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

const oauthCallbackPort = "8080"

// OAuthToken holds an access token for a social platform.
type OAuthToken struct {
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token,omitempty"`
	TokenType    string            `json:"token_type"`
	ExpiresAt    time.Time         `json:"expires_at"`
	Scope        string            `json:"scope,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"` // e.g. linkedin person_id
}

func (t *OAuthToken) Valid() bool {
	return t != nil && t.AccessToken != "" && time.Now().Before(t.ExpiresAt)
}

func (t *OAuthToken) BearerHeader() string {
	return "Bearer " + t.AccessToken
}

// platformDef holds OAuth 2.0 endpoints for one social platform.
type platformDef struct {
	authURL      string
	tokenURL     string
	scopes       []string
	usePKCE      bool
	callbackPath string
	devPortal    string
}

var oauthPlatforms = map[string]platformDef{
	"twitter": {
		authURL:      "https://twitter.com/i/oauth2/authorize",
		tokenURL:     "https://api.twitter.com/2/oauth2/token",
		scopes:       []string{"tweet.read", "tweet.write", "users.read", "offline.access"},
		usePKCE:      true,
		callbackPath: "/callback/twitter",
		devPortal:    "https://developer.twitter.com/en/portal/dashboard",
	},
	"linkedin": {
		authURL:      "https://www.linkedin.com/oauth/v2/authorization",
		tokenURL:     "https://www.linkedin.com/oauth/v2/accessToken",
		scopes:       []string{"r_liteprofile", "r_emailaddress", "w_member_social"},
		usePKCE:      false,
		callbackPath: "/callback/linkedin",
		devPortal:    "https://www.linkedin.com/developers/apps",
	},
	"instagram": {
		authURL:      "https://www.facebook.com/v18.0/dialog/oauth",
		tokenURL:     "https://graph.facebook.com/v18.0/oauth/access_token",
		scopes:       []string{"instagram_basic", "instagram_content_publish", "pages_read_engagement"},
		usePKCE:      false,
		callbackPath: "/callback/instagram",
		devPortal:    "https://developers.facebook.com/apps",
	},
	"reddit": {
		authURL:      "https://www.reddit.com/api/v1/authorize",
		tokenURL:     "https://www.reddit.com/api/v1/access_token",
		scopes:       []string{"identity", "submit", "read"},
		usePKCE:      false,
		callbackPath: "/callback/reddit",
		devPortal:    "https://www.reddit.com/prefs/apps",
	},
}

// OAuthManager handles login flows and token lifecycle for all platforms.
type OAuthManager struct {
	mu     sync.RWMutex
	tokens map[string]*OAuthToken
	mem    *Memory
	client *http.Client
}

func NewOAuthManager(mem *Memory) *OAuthManager {
	m := &OAuthManager{
		tokens: make(map[string]*OAuthToken),
		mem:    mem,
		client: &http.Client{Timeout: 15 * time.Second},
	}
	// Restore tokens saved from a previous session.
	for platform, tok := range mem.Data.OAuthTokens {
		t := tok
		m.tokens[platform] = &t
	}
	return m
}

// IsConnected returns true when the platform has a non-expired access token.
func (o *OAuthManager) IsConnected(platform string) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	t, ok := o.tokens[strings.ToLower(platform)]
	return ok && t.Valid()
}

// Token returns the stored token for a platform (nil if absent or expired).
func (o *OAuthManager) Token(platform string) *OAuthToken {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.tokens[strings.ToLower(platform)]
}

// ConnectedPlatforms returns the names of all authenticated platforms.
func (o *OAuthManager) ConnectedPlatforms() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	var out []string
	for p, t := range o.tokens {
		if t.Valid() {
			out = append(out, p)
		}
	}
	return out
}

// Login runs the full OAuth 2.0 / PKCE browser flow for a platform.
// It opens the user's browser, starts a local callback server, waits for the
// authorization code, exchanges it for a token, and persists the result.
func (o *OAuthManager) Login(ctx context.Context, platform string) (string, error) {
	platform = strings.ToLower(platform)
	def, ok := oauthPlatforms[platform]
	if !ok {
		supported := make([]string, 0, len(oauthPlatforms))
		for k := range oauthPlatforms {
			supported = append(supported, k)
		}
		return "", fmt.Errorf("unknown platform %q — supported: %s", platform, strings.Join(supported, ", "))
	}

	clientID := os.Getenv(envKey(platform, "CLIENT_ID"))
	clientSecret := os.Getenv(envKey(platform, "CLIENT_SECRET"))
	if clientID == "" {
		return "", fmt.Errorf(
			"set %s in your .env to enable %s login.\n  Create an OAuth app at: %s\n  Callback URL to register: http://localhost:%s%s",
			envKey(platform, "CLIENT_ID"), platform, def.devPortal, oauthCallbackPort, def.callbackPath,
		)
	}

	// PKCE (required for Twitter public client)
	var codeVerifier, codeChallenge string
	if def.usePKCE {
		codeVerifier = pkceVerifier()
		codeChallenge = pkceChallenge(codeVerifier)
	}

	state := randomState()
	redirectURI := "http://localhost:" + oauthCallbackPort + def.callbackPath

	// Build the authorization URL.
	q := url.Values{
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {strings.Join(def.scopes, " ")},
		"state":         {state},
	}
	if def.usePKCE {
		q.Set("code_challenge", codeChallenge)
		q.Set("code_challenge_method", "S256")
	}
	// Reddit requires duration=permanent for refresh tokens
	if platform == "reddit" {
		q.Set("duration", "permanent")
	}
	authURL := def.authURL + "?" + q.Encode()

	// Start local callback server.
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(def.callbackPath, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errCh <- fmt.Errorf("OAuth state mismatch — possible CSRF attack")
			return
		}
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			desc := r.URL.Query().Get("error_description")
			http.Error(w, errParam+": "+desc, http.StatusBadRequest)
			errCh <- fmt.Errorf("OAuth denied: %s — %s", errParam, desc)
			return
		}
		code := r.URL.Query().Get("code")
		fmt.Fprintf(w,
			`<html><body style="font-family:sans-serif;padding:60px;text-align:center">
			<h2>✅ Priya is connected to %s!</h2>
			<p>You can close this tab and return to the app.</p>
			</body></html>`, platform)
		codeCh <- code
	})

	ln, err := net.Listen("tcp", ":"+oauthCallbackPort)
	if err != nil {
		return "", fmt.Errorf("cannot start callback server on :%s (%w) — ensure the port is free", oauthCallbackPort, err)
	}
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer srv.Close()

	// Open browser for the user.
	openBrowser(authURL)
	log.Printf("[OAuth] Browser opened for %s. Waiting for login callback on :%s...", platform, oauthCallbackPort)

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errCh:
		return "", err
	case code := <-codeCh:
		token, err := o.exchangeCode(platform, def, code, codeVerifier, redirectURI, clientID, clientSecret)
		if err != nil {
			return "", fmt.Errorf("token exchange failed: %w", err)
		}
		// For LinkedIn, fetch the user's person ID so we can author posts.
		if platform == "linkedin" {
			if personID, err := o.fetchLinkedInPersonID(token.AccessToken); err == nil {
				if token.Extra == nil {
					token.Extra = make(map[string]string)
				}
				token.Extra["person_id"] = personID
			}
		}
		o.persistToken(platform, token)
		return fmt.Sprintf("✅ Connected to %s! Priya will now post directly to your account.", platform), nil
	}
}

// Logout removes the stored token for a platform.
func (o *OAuthManager) Logout(platform string) {
	platform = strings.ToLower(platform)
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.tokens, platform)
	if o.mem.Data.OAuthTokens != nil {
		delete(o.mem.Data.OAuthTokens, platform)
	}
	_ = o.mem.Save()
}

// StatusText returns a human-readable connection status for all platforms.
func (o *OAuthManager) StatusText() string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	var sb strings.Builder
	sb.WriteString("Social platform connections:\n\n")
	for name := range oauthPlatforms {
		t, ok := o.tokens[name]
		if ok && t.Valid() {
			sb.WriteString(fmt.Sprintf("  ✅ %-12s connected (expires %s)\n", name, t.ExpiresAt.Format("2006-01-02")))
		} else {
			sb.WriteString(fmt.Sprintf("  ❌ %-12s not connected  →  /login %s\n", name, name))
		}
	}
	return sb.String()
}

// ── Token exchange ────────────────────────────────────────────────────────────

func (o *OAuthManager) exchangeCode(platform string, def platformDef, code, verifier, redirectURI, clientID, clientSecret string) (*OAuthToken, error) {
	params := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURI},
		"client_id":    {clientID},
	}
	if def.usePKCE && verifier != "" {
		params.Set("code_verifier", verifier)
	}
	if !def.usePKCE && clientSecret != "" {
		params.Set("client_secret", clientSecret)
	}

	req, _ := http.NewRequest("POST", def.tokenURL, strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Twitter requires HTTP Basic auth for confidential clients.
	if platform == "twitter" && clientSecret != "" {
		req.SetBasicAuth(clientID, clientSecret)
	}
	// Reddit requires HTTP Basic auth always.
	if platform == "reddit" {
		req.SetBasicAuth(clientID, clientSecret)
		req.Header.Set("User-Agent", "Priya-Bot/1.0")
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("invalid token response: %s", body)
	}
	if e, ok := raw["error"]; ok {
		return nil, fmt.Errorf("%v: %v", e, raw["error_description"])
	}

	tok := &OAuthToken{
		AccessToken:  strVal(raw, "access_token"),
		RefreshToken: strVal(raw, "refresh_token"),
		TokenType:    strVal(raw, "token_type"),
		Scope:        strVal(raw, "scope"),
		ExpiresAt:    time.Now().Add(tokenTTL(raw)),
	}
	return tok, nil
}

// fetchLinkedInPersonID gets the authenticated user's URN from the LinkedIn profile API.
func (o *OAuthManager) fetchLinkedInPersonID(accessToken string) (string, error) {
	req, _ := http.NewRequest("GET", "https://api.linkedin.com/v2/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", err
	}
	if id, ok := raw["id"].(string); ok {
		return id, nil
	}
	return "", fmt.Errorf("no id in LinkedIn profile response")
}

func (o *OAuthManager) persistToken(platform string, token *OAuthToken) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.tokens[platform] = token
	if o.mem.Data.OAuthTokens == nil {
		o.mem.Data.OAuthTokens = make(map[string]OAuthToken)
	}
	o.mem.Data.OAuthTokens[platform] = *token
	_ = o.mem.Save()
}

// ── PKCE helpers ──────────────────────────────────────────────────────────────

func pkceVerifier() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func pkceChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// ── Misc helpers ──────────────────────────────────────────────────────────────

func openBrowser(rawURL string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd, args = "cmd", []string{"/c", "start", rawURL}
	case "darwin":
		cmd, args = "open", []string{rawURL}
	default:
		cmd, args = "xdg-open", []string{rawURL}
	}
	_ = exec.Command(cmd, args...).Start()
}

func tokenTTL(raw map[string]interface{}) time.Duration {
	if v, ok := raw["expires_in"]; ok {
		if n, ok := v.(float64); ok && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 2 * time.Hour
}

func envKey(platform, suffix string) string {
	return strings.ToUpper(platform) + "_" + suffix
}

func strVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
