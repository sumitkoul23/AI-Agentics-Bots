package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// SocialPoster performs real posts to social platforms using stored OAuth tokens.
type SocialPoster struct {
	oauth  *OAuthManager
	client *http.Client
}

func NewSocialPoster(oauth *OAuthManager) *SocialPoster {
	return &SocialPoster{
		oauth:  oauth,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// CanPost returns true when the platform has a valid OAuth token.
func (s *SocialPoster) CanPost(platform string) bool {
	return s.oauth.IsConnected(strings.ToLower(platform))
}

// ── Twitter ───────────────────────────────────────────────────────────────────

// PostTweet publishes a tweet using the authenticated user's token.
func (s *SocialPoster) PostTweet(text string) (string, error) {
	tok := s.oauth.Token("twitter")
	if tok == nil {
		return "", fmt.Errorf("not logged in to Twitter — use /login twitter")
	}

	body, _ := json.Marshal(map[string]interface{}{"text": text})
	req, _ := http.NewRequest("POST", "https://api.twitter.com/2/tweets", bytes.NewReader(body))
	req.Header.Set("Authorization", tok.BearerHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Twitter API error: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	json.Unmarshal(raw, &result)
	if len(result.Errors) > 0 {
		return "", fmt.Errorf("Twitter error: %s", result.Errors[0].Message)
	}
	tweetURL := "https://twitter.com/i/web/status/" + result.Data.ID
	return fmt.Sprintf("✅ Tweet posted!\n%s", tweetURL), nil
}

// GetTwitterTrends returns the top trending topics from Twitter's v2 API.
func (s *SocialPoster) GetTwitterTrends() (string, error) {
	tok := s.oauth.Token("twitter")
	if tok == nil {
		return "", fmt.Errorf("not connected to Twitter")
	}
	// Use recent search to surface high-volume hashtags as a trends proxy
	q := url.Values{
		"query":       {"#trending lang:en -is:retweet"},
		"max_results": {"10"},
		"tweet.fields": {"public_metrics,created_at"},
	}
	req, _ := http.NewRequest("GET", "https://api.twitter.com/2/tweets/search/recent?"+q.Encode(), nil)
	req.Header.Set("Authorization", tok.BearerHeader())

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var result struct {
		Data []struct {
			Text            string `json:"text"`
			PublicMetrics   struct {
				LikeCount    int `json:"like_count"`
				RetweetCount int `json:"retweet_count"`
			} `json:"public_metrics"`
		} `json:"data"`
	}
	json.Unmarshal(raw, &result)

	var sb strings.Builder
	sb.WriteString("Twitter trending right now:\n\n")
	for i, t := range result.Data {
		sb.WriteString(fmt.Sprintf("%d. %s\n   ❤️ %d  🔁 %d\n\n",
			i+1,
			truncate(t.Text, 120),
			t.PublicMetrics.LikeCount,
			t.PublicMetrics.RetweetCount,
		))
	}
	return sb.String(), nil
}

// ── LinkedIn ──────────────────────────────────────────────────────────────────

// PostLinkedIn publishes a text update on LinkedIn.
func (s *SocialPoster) PostLinkedIn(text string) (string, error) {
	tok := s.oauth.Token("linkedin")
	if tok == nil {
		return "", fmt.Errorf("not logged in to LinkedIn — use /login linkedin")
	}

	personID := ""
	if tok.Extra != nil {
		personID = tok.Extra["person_id"]
	}
	if personID == "" {
		return "", fmt.Errorf("LinkedIn person ID not found — please /logout linkedin then /login linkedin again")
	}

	body, _ := json.Marshal(map[string]interface{}{
		"author":         "urn:li:person:" + personID,
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary":    map[string]string{"text": text},
				"shareMediaCategory": "NONE",
			},
		},
		"visibility": map[string]string{
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	})

	req, _ := http.NewRequest("POST", "https://api.linkedin.com/v2/ugcPosts", bytes.NewReader(body))
	req.Header.Set("Authorization", tok.BearerHeader())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("LinkedIn API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		postID := resp.Header.Get("X-RestLi-Id")
		return fmt.Sprintf("✅ LinkedIn post published!\nPost ID: %s", postID), nil
	}
	raw, _ := io.ReadAll(resp.Body)
	return "", fmt.Errorf("LinkedIn error %d: %s", resp.StatusCode, raw)
}

// ── Instagram ─────────────────────────────────────────────────────────────────

// PostInstagram publishes a caption to Instagram (image URL required by the Graph API).
// For caption-only posts, we surface the content as a ready-to-copy block.
func (s *SocialPoster) PostInstagram(imageURL, caption string) (string, error) {
	tok := s.oauth.Token("instagram")
	if tok == nil {
		return "", fmt.Errorf("not logged in to Instagram — use /login instagram")
	}

	accountID := os.Getenv("INSTAGRAM_BUSINESS_ACCOUNT_ID")
	if accountID == "" {
		return "", fmt.Errorf("set INSTAGRAM_BUSINESS_ACCOUNT_ID in .env (your Instagram Business account ID)")
	}

	if imageURL == "" {
		// Instagram Graph API requires a media URL — return draft instead.
		return fmt.Sprintf(
			"📋 Instagram caption ready to post (add an image URL with /post instagram <image_url>):\n\n%s", caption,
		), nil
	}

	// Step 1: Create media container.
	createURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/media", accountID)
	q := url.Values{
		"image_url":    {imageURL},
		"caption":      {caption},
		"access_token": {tok.AccessToken},
	}
	resp, err := s.client.Post(createURL+"?"+q.Encode(), "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("Instagram create error: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var container struct {
		ID    string `json:"id"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	json.Unmarshal(raw, &container)
	if container.Error.Message != "" {
		return "", fmt.Errorf("Instagram error: %s", container.Error.Message)
	}

	// Step 2: Publish the container.
	pubURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/media_publish", accountID)
	pubQ := url.Values{
		"creation_id":  {container.ID},
		"access_token": {tok.AccessToken},
	}
	pubResp, err := s.client.Post(pubURL+"?"+pubQ.Encode(), "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("Instagram publish error: %w", err)
	}
	defer pubResp.Body.Close()
	pubRaw, _ := io.ReadAll(pubResp.Body)

	var published struct {
		ID string `json:"id"`
	}
	json.Unmarshal(pubRaw, &published)
	return fmt.Sprintf("✅ Instagram post published!\nMedia ID: %s", published.ID), nil
}

// ── Reddit ────────────────────────────────────────────────────────────────────

// PostReddit submits a text post to a subreddit.
func (s *SocialPoster) PostReddit(subreddit, title, text string) (string, error) {
	tok := s.oauth.Token("reddit")
	if tok == nil {
		return "", fmt.Errorf("not logged in to Reddit — use /login reddit")
	}

	params := url.Values{
		"api_type": {"json"},
		"kind":     {"self"},
		"sr":       {subreddit},
		"title":    {title},
		"text":     {text},
	}
	req, _ := http.NewRequest("POST", "https://oauth.reddit.com/api/submit",
		strings.NewReader(params.Encode()))
	req.Header.Set("Authorization", tok.BearerHeader())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Priya-Bot/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Reddit API error: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var result struct {
		JSON struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
			Errors [][]string `json:"errors"`
		} `json:"json"`
	}
	json.Unmarshal(raw, &result)
	if len(result.JSON.Errors) > 0 && len(result.JSON.Errors[0]) > 1 {
		return "", fmt.Errorf("Reddit error: %s", result.JSON.Errors[0][1])
	}
	return fmt.Sprintf("✅ Reddit post submitted!\n%s", result.JSON.Data.URL), nil
}
