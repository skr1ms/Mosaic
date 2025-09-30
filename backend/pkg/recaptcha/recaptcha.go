package recaptcha

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/skr1ms/mosaic/pkg/middleware"
)

type ReCaptchaResponse struct {
	Success     bool     `json:"success"`
	Score       float64  `json:"score"`
	Action      string   `json:"action"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

type Verifier struct {
	secret     string
	minScore   float64
	httpClient *http.Client
	logger     *middleware.Logger
}

func NewVerifier(secret string, minScore float64, logger *middleware.Logger) *Verifier {
	return &Verifier{
		secret:     secret,
		minScore:   minScore,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		logger:     logger,
	}
}

func (v *Verifier) Verify(token, expectedAction string) (bool, error) {
	if token == "" {
		v.logger.GetZerologLogger().Error().Msg("Empty reCAPTCHA token provided")
		return false, fmt.Errorf("empty reCAPTCHA token")
	}

	form := url.Values{}
	form.Add("secret", v.secret)
	form.Add("response", token)

	resp, err := v.httpClient.PostForm(
		"https://www.google.com/recaptcha/api/siteverify",
		form,
	)
	if err != nil {
		v.logger.GetZerologLogger().Error().Err(err).Msg("reCAPTCHA request error")
		return false, fmt.Errorf("reCAPTCHA request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		v.logger.GetZerologLogger().Error().Int("status_code", resp.StatusCode).Msg("Invalid reCAPTCHA status")
		return false, fmt.Errorf("invalid reCAPTCHA status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		v.logger.GetZerologLogger().Error().Err(err).Msg("Response reading error")
		return false, fmt.Errorf("response reading error: %w", err)
	}

	var result ReCaptchaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		v.logger.GetZerologLogger().Error().Err(err).Str("body", string(body)).Msg("JSON parsing error")
		return false, fmt.Errorf("JSON parsing error: %w", err)
	}

	if !result.Success {
		v.logger.GetZerologLogger().Error().Interface("error_codes", result.ErrorCodes).Msg("reCAPTCHA failed")
		return false, fmt.Errorf("reCAPTCHA failed, errors: %v", result.ErrorCodes)
	}

	if result.Action == "" {
		return true, nil
	}

	if result.Score < v.minScore {
		v.logger.GetZerologLogger().Error().Float64("score", result.Score).Float64("min_score", v.minScore).Msg("Low reCAPTCHA score")
		return false, fmt.Errorf("low reCAPTCHA score: %.2f", result.Score)
	}

	if expectedAction != "" && result.Action != expectedAction {
		v.logger.GetZerologLogger().Error().Str("expected_action", expectedAction).Str("got_action", result.Action).Msg("Action mismatch")
		return false, fmt.Errorf("action mismatch: expected %s, got %s",
			expectedAction, result.Action)
	}

	v.logger.GetZerologLogger().Info().
		Str("action", result.Action).
		Float64("score", result.Score).
		Msg("reCAPTCHA verification successful")
	return true, nil
}
