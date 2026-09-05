package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/config"
	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/dhruv15803/url-shortener-service/internal/repositories"
	"github.com/dhruv15803/url-shortener-service/internal/token"
	"github.com/lib/pq"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleUserInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"

const maxUsernameAttempts = 5

var usernameSanitizer = regexp.MustCompile(`[^a-z0-9._-]`)

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type AuthService struct {
	repository  *repositories.Repository
	oauthConfig *oauth2.Config
	jwtSecret   string
	jwtExpiry   time.Duration
}

func NewAuthService(repository *repositories.Repository, cfg *config.Config) *AuthService {
	return &AuthService{
		repository: repository,
		oauthConfig: &oauth2.Config{
			ClientID:     cfg.OAuthConfig.GoogleClientID,
			ClientSecret: cfg.OAuthConfig.GoogleClientSecret,
			RedirectURL:  cfg.OAuthConfig.GoogleRedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		jwtSecret: cfg.JWTConfig.Secret,
		jwtExpiry: cfg.JWTConfig.Expiry,
	}
}

func (s *AuthService) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *AuthService) GoogleAuthCodeURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (s *AuthService) HandleGoogleCallback(ctx context.Context, code string) (*models.User, string, error) {
	oauthToken, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange google auth code: %w", err)
	}

	info, err := s.fetchGoogleUserInfo(ctx, oauthToken)
	if err != nil {
		return nil, "", err
	}

	user, err := s.upsertGoogleUser(info)
	if err != nil {
		return nil, "", err
	}

	jwtToken, err := token.Generate(s.jwtSecret, s.jwtExpiry, user.ID, user.Email)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate session token: %w", err)
	}

	return user, jwtToken, nil
}

func (s *AuthService) fetchGoogleUserInfo(ctx context.Context, oauthToken *oauth2.Token) (*googleUserInfo, error) {
	client := s.oauthConfig.Client(ctx, oauthToken)

	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch google user info: %w", err)
	}
	defer resp.Body.Close()

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode google user info: %w", err)
	}

	return &info, nil
}

func (s *AuthService) upsertGoogleUser(info *googleUserInfo) (*models.User, error) {
	now := time.Now()
	baseUsername := sanitizeUsername(strings.Split(info.Email, "@")[0])

	var (
		user *models.User
		err  error
	)

	for attempt := 0; attempt < maxUsernameAttempts; attempt++ {
		username := baseUsername
		if attempt > 0 {
			suffix, suffixErr := randDigits(4)
			if suffixErr != nil {
				return nil, suffixErr
			}
			username = baseUsername + suffix
		}

		candidate := &models.User{
			Email:      info.Email,
			GoogleID:   &info.Sub,
			Username:   username,
			ImageURL:   &info.Picture,
			IsVerified: true,
			VerifiedAt: &now,
		}

		user, err = s.repository.Users.UpsertGoogleUser(candidate)
		if err == nil {
			return user, nil
		}

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "users_username_key" {
			continue
		}

		return nil, fmt.Errorf("failed to upsert google user: %w", err)
	}

	return nil, fmt.Errorf("failed to generate a unique username after %d attempts: %w", maxUsernameAttempts, err)
}

func sanitizeUsername(raw string) string {
	sanitized := usernameSanitizer.ReplaceAllString(strings.ToLower(raw), "")
	if sanitized == "" {
		return "user"
	}
	return sanitized
}

func randDigits(n int) (string, error) {
	max := big.NewInt(1)
	for i := 0; i < n; i++ {
		max.Mul(max, big.NewInt(10))
	}

	v, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0*d", n, v.Int64()), nil
}
