package handlers

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/supabase-community/gotrue-go/types"
	"github.com/supabase-community/supabase-go"
)

func authenticatedUserFromRequest(r *http.Request) (*types.UserResponse, error) {
	token := strings.TrimSpace(r.URL.Query().Get("access_token"))
	if token == "" {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			return nil, errors.New("missing bearer token")
		}
		token = strings.TrimSpace(authHeader[len("Bearer "):])
	}
	if token == "" {
		return nil, errors.New("missing bearer token")
	}

	sb, err := supabase.NewClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_API_KEY"), nil)
	if err != nil {
		return nil, err
	}
	sb.Auth = sb.Auth.WithToken(token)
	return sb.Auth.GetUser()
}
