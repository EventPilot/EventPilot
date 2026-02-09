package xapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	twitter "github.com/g8rswimmer/go-twitter/v2"
)

// Client handles communication with X (Twitter) API v2
type Client struct {
	client *twitter.Client
}

// Config holds X API configuration
type Config struct {
	APIKey            string
	APISecret         string
	AccessToken       string
	AccessTokenSecret string
}

// Tweet represents a posted tweet
type Tweet struct {
	ID   string
	Text string
}

// NewClient creates a new X API v2 client
func NewClient(config Config) (*Client, error) {
	if config.APIKey == "" || config.APISecret == "" ||
		config.AccessToken == "" || config.AccessTokenSecret == "" {
		return nil, fmt.Errorf("all X API credentials are required")
	}

	// Create Twitter client with proper OAuth 1.0a authorizer
	client := &twitter.Client{
		Authorizer: &oAuth1Authorizer{
			consumerKey:       config.APIKey,
			consumerSecret:    config.APISecret,
			accessToken:       config.AccessToken,
			accessTokenSecret: config.AccessTokenSecret,
		},
		Client: http.DefaultClient,
		Host:   "https://api.twitter.com",
	}

	return &Client{client: client}, nil
}

// oAuth1Authorizer implements proper OAuth 1.0a authorization
type oAuth1Authorizer struct {
	consumerKey       string
	consumerSecret    string
	accessToken       string
	accessTokenSecret string
}

// Add adds OAuth 1.0a authorization header to the request
func (o *oAuth1Authorizer) Add(req *http.Request) {
	// Generate OAuth 1.0a signature
	oauthParams := o.buildOAuthParams(req)
	signature := o.generateSignature(req, oauthParams)
	oauthParams["oauth_signature"] = signature

	// Build Authorization header
	authHeader := o.buildAuthHeader(oauthParams)
	req.Header.Set("Authorization", authHeader)
}

// buildOAuthParams creates the OAuth parameters
func (o *oAuth1Authorizer) buildOAuthParams(req *http.Request) map[string]string {
	return map[string]string{
		"oauth_consumer_key":     o.consumerKey,
		"oauth_token":            o.accessToken,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_nonce":            o.generateNonce(),
		"oauth_version":          "1.0",
	}
}

// generateNonce creates a random nonce
func (o *oAuth1Authorizer) generateNonce() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// generateSignature creates the OAuth 1.0a signature
func (o *oAuth1Authorizer) generateSignature(req *http.Request, oauthParams map[string]string) string {
	// Collect all parameters
	params := make(map[string]string)

	// Add OAuth params
	for k, v := range oauthParams {
		params[k] = v
	}

	// Add query parameters
	for k, v := range req.URL.Query() {
		params[k] = v[0]
	}

	// Create parameter string
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, percentEncode(k)+"="+percentEncode(params[k]))
	}
	paramString := strings.Join(pairs, "&")

	// Create signature base string
	method := req.Method
	baseURL := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path
	signatureBase := method + "&" + percentEncode(baseURL) + "&" + percentEncode(paramString)

	// Create signing key
	signingKey := percentEncode(o.consumerSecret) + "&" + percentEncode(o.accessTokenSecret)

	// Generate HMAC-SHA1 signature
	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(signatureBase))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature
}

// buildAuthHeader builds the OAuth Authorization header
func (o *oAuth1Authorizer) buildAuthHeader(params map[string]string) string {
	var pairs []string

	// Sort keys for consistent output
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		pairs = append(pairs, percentEncode(k)+"=\""+percentEncode(params[k])+"\"")
	}

	return "OAuth " + strings.Join(pairs, ", ")
}

// percentEncode encodes a string per OAuth spec
func percentEncode(s string) string {
	return url.QueryEscape(s)
}

// PostTweet posts a tweet to X
func (c *Client) PostTweet(text string) (*Tweet, error) {
	if len(text) > 280 {
		return nil, fmt.Errorf("tweet text exceeds 280 characters (length: %d)", len(text))
	}

	req := twitter.CreateTweetRequest{
		Text: text,
	}

	ctx := context.Background()
	tweetResponse, err := c.client.CreateTweet(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to post tweet: %w", err)
	}

	if tweetResponse == nil || tweetResponse.Tweet == nil {
		return nil, fmt.Errorf("empty response from X API")
	}

	tweet := &Tweet{
		ID:   tweetResponse.Tweet.ID,
		Text: tweetResponse.Tweet.Text,
	}

	return tweet, nil
}

// PostThread posts a thread of tweets
func (c *Client) PostThread(tweets []string) ([]*Tweet, error) {
	if len(tweets) == 0 {
		return nil, fmt.Errorf("no tweets to post")
	}

	postedTweets := make([]*Tweet, 0, len(tweets))
	var previousTweetID string

	for i, text := range tweets {
		if len(text) > 280 {
			return postedTweets, fmt.Errorf("tweet %d exceeds 280 characters (length: %d)", i+1, len(text))
		}

		req := twitter.CreateTweetRequest{
			Text: text,
		}

		// Reply to previous tweet to create a thread
		if previousTweetID != "" {
			req.Reply = &twitter.CreateTweetReply{
				InReplyToTweetID: previousTweetID,
			}
		}

		ctx := context.Background()
		tweetResponse, err := c.client.CreateTweet(ctx, req)
		if err != nil {
			return postedTweets, fmt.Errorf("failed to post tweet %d: %w", i+1, err)
		}

		if tweetResponse == nil || tweetResponse.Tweet == nil {
			return postedTweets, fmt.Errorf("empty response from X API for tweet %d", i+1)
		}

		tweet := &Tweet{
			ID:   tweetResponse.Tweet.ID,
			Text: tweetResponse.Tweet.Text,
		}

		postedTweets = append(postedTweets, tweet)
		previousTweetID = tweet.ID
	}

	return postedTweets, nil
}

// GetUserInfo gets information about the authenticated user
func (c *Client) GetUserInfo() (string, error) {
	ctx := context.Background()

	// Get the authenticated user's info
	opts := twitter.UserLookupOpts{
		UserFields: []twitter.UserField{twitter.UserFieldUserName},
	}

	userResponse, err := c.client.AuthUserLookup(ctx, opts)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	// AuthUserLookup returns Raw.Users as a slice with the authenticated user
	if userResponse == nil || userResponse.Raw == nil || len(userResponse.Raw.Users) == 0 {
		return "", fmt.Errorf("empty user response from X API")
	}

	// The first (and only) user in the array is the authenticated user
	return userResponse.Raw.Users[0].UserName, nil
}
