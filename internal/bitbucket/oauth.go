package bitbucket

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	authURL  = "https://bitbucket.org/site/oauth2/authorize"
	tokenURL = "https://bitbucket.org/site/oauth2/access_token"
)

// TokenData holds OAuth tokens persisted to disk.
type TokenData struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	Scopes       string    `json:"scopes"`
	ObtainedAt   time.Time `json:"obtained_at"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
}

// IsExpired returns true if the access token is expired (with 5 min buffer).
func (t *TokenData) IsExpired() bool {
	expiry := t.ObtainedAt.Add(time.Duration(t.ExpiresIn) * time.Second)
	return time.Now().After(expiry.Add(-5 * time.Minute))
}

// TokenPath returns the path to the token storage file.
func TokenPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}
	return filepath.Join(home, ".config", "bitbucket-mcp", "token.json"), nil
}

// SaveToken persists token data to disk.
func SaveToken(token *TokenData) error {
	path, err := TokenPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing token file: %w", err)
	}

	return nil
}

// LoadToken reads persisted token data from disk.
func LoadToken() (*TokenData, error) {
	path, err := TokenPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading token file: %w", err)
	}

	var token TokenData
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parsing token file: %w", err)
	}

	return &token, nil
}

// RefreshAccessToken uses the refresh token to get a new access token.
func RefreshAccessToken(token *TokenData) error {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {token.RefreshToken},
	}

	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating refresh request: %w", err)
	}
	req.SetBasicAuth(token.ClientID, token.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("refreshing token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh failed (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scopes       string `json:"scopes"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing refresh response: %w", err)
	}

	token.AccessToken = result.AccessToken
	if result.RefreshToken != "" {
		token.RefreshToken = result.RefreshToken
	}
	token.ExpiresIn = result.ExpiresIn
	token.Scopes = result.Scopes
	token.ObtainedAt = time.Now()

	return SaveToken(token)
}

// OAuthLogin performs the Authorization Code Grant flow with a localhost callback.
// It opens the user's browser, waits for the callback, exchanges the code, and stores the token.
func OAuthLogin(clientID, clientSecret string) error {
	// Generate state for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return fmt.Errorf("generating state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("finding free port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://localhost:%d/callback", port)

	// Build authorize URL
	params := url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
		"state":         {state},
	}
	authorizeURL := authURL + "?" + params.Encode()

	// Channel for the auth code
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	// Setup callback server
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("state mismatch - possible CSRF attack")
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		if errParam := r.URL.Query().Get("error"); errParam != "" {
			errDesc := r.URL.Query().Get("error_description")
			errCh <- fmt.Errorf("OAuth error: %s - %s", errParam, errDesc)
			fmt.Fprintf(w, "<html><body><h2>Authentication failed</h2><p>%s: %s</p><p>You can close this window.</p></body></html>", errParam, errDesc)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback")
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, "<html><body><h2>Authenticated!</h2><p>You can close this window and return to your terminal.</p></body></html>")
		codeCh <- code
	})

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(listener); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Open browser
	fmt.Printf("\nOpening browser for Bitbucket authentication...\n")
	fmt.Printf("If the browser doesn't open, visit:\n  %s\n\n", authorizeURL)
	fmt.Printf("Callback URL: %s\n", callbackURL)
	fmt.Printf("Waiting for authentication...\n\n")
	openBrowser(authorizeURL)

	// Wait for code or error (timeout after 5 minutes)
	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		srv.Shutdown(context.Background())
		return err
	case <-time.After(5 * time.Minute):
		srv.Shutdown(context.Background())
		return fmt.Errorf("authentication timed out after 5 minutes")
	}

	srv.Shutdown(context.Background())

	// Exchange code for tokens
	fmt.Println("Exchanging code for tokens...")

	formData := url.Values{
		"grant_type": {"authorization_code"},
		"code":       {code},
	}

	tokenReq, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("creating token request: %w", err)
	}
	tokenReq.SetBasicAuth(clientID, clientSecret)
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		return fmt.Errorf("exchanging code: %w", err)
	}
	defer tokenResp.Body.Close()

	body, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return fmt.Errorf("reading token response: %w", err)
	}

	if tokenResp.StatusCode != http.StatusOK {
		return fmt.Errorf("token exchange failed (%d): %s", tokenResp.StatusCode, string(body))
	}

	var tokenResult struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scopes       string `json:"scopes"`
	}
	if err := json.Unmarshal(body, &tokenResult); err != nil {
		return fmt.Errorf("parsing token response: %w", err)
	}

	token := &TokenData{
		AccessToken:  tokenResult.AccessToken,
		RefreshToken: tokenResult.RefreshToken,
		TokenType:    tokenResult.TokenType,
		ExpiresIn:    tokenResult.ExpiresIn,
		Scopes:       tokenResult.Scopes,
		ObtainedAt:   time.Now(),
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	if err := SaveToken(token); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}

	path, _ := TokenPath()
	fmt.Printf("\nAuthentication successful!\n")
	fmt.Printf("Scopes: %s\n", tokenResult.Scopes)
	fmt.Printf("Token saved to: %s\n", path)
	return nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}
