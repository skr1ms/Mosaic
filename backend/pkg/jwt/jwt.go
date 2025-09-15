package jwt

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type JWT struct {
	SecretKey        string
	RefreshSecretKey string
}

func NewJWT(secretKey, refreshSecretKey string) *JWT {
	return &JWT{
		SecretKey:        secretKey,
		RefreshSecretKey: refreshSecretKey,
	}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type SignedOpenClaims struct {
	MessageID uint      `json:"message_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CreateSignedOpenToken creates short token for public access to attachment
func (j *JWT) CreateSignedOpenToken(messageID uint, ttl time.Duration) (string, error) {
	now := time.Now()
	payload := jwt.MapClaims{
		"mid": messageID,
		"exp": now.Add(ttl).Unix(),
		"nbf": now.Unix(),
		"typ": "open",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString([]byte(j.SecretKey))
}

// ParseSignedOpenToken validates and returns messageID
func (j *JWT) ParseSignedOpenToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error().Msg("Unexpected signing method in open token")
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(j.SecretKey), nil
	})
	if err != nil || !token.Valid {
		log.Error().Err(err).Msg("Invalid open token")
		return 0, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Error().Msg("Invalid claims in open token")
		return 0, fmt.Errorf("invalid claims")
	}
	if t, ok := claims["typ"].(string); !ok || t != "open" {
		log.Error().Str("token_type", t).Msg("Invalid token type in open token")
		return 0, fmt.Errorf("invalid token type")
	}
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			log.Error().Msg("Open token expired")
			return 0, fmt.Errorf("token expired")
		}
	}
	var mid uint
	switch v := claims["mid"].(type) {
	case float64:
		mid = uint(v)
	case int64:
		mid = uint(v)
	default:
		log.Error().Interface("mid", v).Msg("Invalid message ID in open token")
		return 0, fmt.Errorf("invalid message id")
	}
	return mid, nil
}

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Login     string    `json:"login"`
	Role      string    `json:"role"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

const (
	AdminAccessTokenDuration  = time.Hour * 2
	AdminRefreshTokenDuration = time.Hour * 24 * 3

	PartnerAccessTokenDuration  = time.Hour * 1
	PartnerRefreshTokenDuration = time.Hour * 24

	PasswordResetTokenDuration = time.Minute * 15
)

func (j *JWT) CreateAccessToken(userID uuid.UUID, login, role string) (string, error) {
	var duration time.Duration
	if role == "admin" || role == "main_admin" {
		duration = AdminAccessTokenDuration
	} else {
		duration = PartnerAccessTokenDuration
	}

	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Login:     login,
		Role:      role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.SecretKey))
}

func (j *JWT) CreateRefreshToken(userID uuid.UUID, login, role string) (string, error) {
	var duration time.Duration
	if role == "admin" || role == "main_admin" {
		duration = AdminRefreshTokenDuration
	} else {
		duration = PartnerRefreshTokenDuration
	}

	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Login:     login,
		Role:      role,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.RefreshSecretKey))
}

// CreateTokenPair creates access and refresh token pair
func (j *JWT) CreateTokenPair(userID uuid.UUID, login, role string) (*TokenPair, error) {
	log.Info().
		Str("user_id", userID.String()).
		Str("login", login).
		Str("role", role).
		Msg("Creating token pair")

	accessToken, err := j.CreateAccessToken(userID, login, role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create access token")
		return nil, fmt.Errorf("error creating access token: %w", err)
	}

	refreshToken, err := j.CreateRefreshToken(userID, login, role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create refresh token")
		return nil, fmt.Errorf("error creating refresh token: %w", err)
	}

	var accessDuration time.Duration
	if role == "admin" || role == "main_admin" {
		accessDuration = AdminAccessTokenDuration
	} else {
		accessDuration = PartnerAccessTokenDuration
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessDuration.Seconds()),
	}, nil
}

// ValidateAccessToken validates access token
func (j *JWT) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error().Msg("Unexpected signing method in access token validation")
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(j.SecretKey), nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to validate access token")
		return nil, fmt.Errorf("error validating access token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.TokenType != "access" {
			log.Error().Str("token_type", claims.TokenType).Msg("Invalid token type in access token validation")
			return nil, fmt.Errorf("invalid token type")
		}
		return claims, nil
	}

	log.Error().Msg("Invalid access token claims or token not valid")
	return nil, fmt.Errorf("invalid token")
}

// ValidateRefreshToken validates refresh token
func (j *JWT) ValidateRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error().Msg("Unexpected signing method in refresh token validation")
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(j.RefreshSecretKey), nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to validate refresh token")
		return nil, fmt.Errorf("error validating refresh token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.TokenType != "refresh" {
			log.Error().Str("token_type", claims.TokenType).Msg("Invalid token type in refresh token validation")
			return nil, fmt.Errorf("invalid token type")
		}
		return claims, nil
	}

	log.Error().Msg("Invalid refresh token claims or token not valid")
	return nil, fmt.Errorf("invalid token")
}

func (j *JWT) RefreshTokens(refreshTokenString string) (*TokenPair, error) {
	claims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to refresh tokens")
		return nil, fmt.Errorf("error refreshing tokens: %w", err)
	}

	return j.CreateTokenPair(claims.UserID, claims.Login, claims.Role)
}

func (j *JWT) GetSecretKey() []byte {
	return []byte(j.SecretKey)
}

func GetClaimsFromFiberContext(c *fiber.Ctx) (*Claims, error) {
	user := c.Locals("user")
	if user == nil {
		log.Error().Msg("Token not found in Fiber context")
		return nil, fmt.Errorf("token not found in context")
	}

	token, ok := user.(*jwt.Token)
	if !ok {
		log.Error().Interface("user_type", fmt.Sprintf("%T", user)).Msg("Invalid token type in Fiber context")
		return nil, fmt.Errorf("invalid token type: expected *jwt.Token, got %T", user)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Error().Msg("Invalid token claims in Fiber context")
		return nil, fmt.Errorf("invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		log.Error().Msg("User ID not found in token claims")
		return nil, fmt.Errorf("user_id not found in claims")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Msg("Invalid user ID format in token claims")
		return nil, fmt.Errorf("invalid user_id format")
	}

	login, ok := claims["login"].(string)
	if !ok {
		log.Error().Msg("Login not found in token claims")
		return nil, fmt.Errorf("login not found in claims")
	}

	role, ok := claims["role"].(string)
	if !ok {
		log.Error().Msg("Role not found in token claims")
		return nil, fmt.Errorf("role not found in claims")
	}

	tokenType, ok := claims["token_type"].(string)
	if !ok {
		log.Error().Msg("Token type not found in token claims")
		return nil, fmt.Errorf("token_type not found in claims")
	}

	return &Claims{
		UserID:    userID,
		Login:     login,
		Role:      role,
		TokenType: tokenType,
	}, nil
}

func (j *JWT) CreatePasswordResetToken(userID uuid.UUID, email string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Login:     email,
		TokenType: "password_reset",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.SecretKey))
}

func (j *JWT) ValidatePasswordResetToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error().Msg("Unexpected signing method in password reset token validation")
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(j.SecretKey), nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to validate password reset token")
		return nil, fmt.Errorf("error validating password reset token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.TokenType != "password_reset" {
			log.Error().Str("token_type", claims.TokenType).Msg("Invalid token type in password reset token validation")
			return nil, fmt.Errorf("invalid token type")
		}
		return claims, nil
	}

	log.Error().Msg("Invalid password reset token claims or token not valid")
	return nil, fmt.Errorf("invalid token")
}
