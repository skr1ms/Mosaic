package recaptcha

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Verifier struct {
	secret     string
	minScore   float64 
	httpClient *http.Client
}

func NewVerifier(secret string, minScore float64) *Verifier {
	return &Verifier{
		secret:     secret,
		minScore:   minScore,
		httpClient: &http.Client{Timeout: 5 * time.Second}, 
	}
}

// Verify проверяет токен reCAPTCHA с улучшенной обработкой ошибок
func (v *Verifier) Verify(token, expectedAction string) (bool, error) {
	if token == "" {
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
		return false, fmt.Errorf("reCAPTCHA request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("invalid reCAPTCHA status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("response reading error: %w", err)
	}

	var result ReCaptchaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("JSON parsing error: %w", err)
	}

	if !result.Success {
		return false, fmt.Errorf("reCAPTCHA failed, errors: %v", result.ErrorCodes)
	}

	if result.Score < v.minScore {
		return false, fmt.Errorf("low reCAPTCHA score: %.2f", result.Score)
	}

	if expectedAction != "" && result.Action != expectedAction {
		return false, fmt.Errorf("action mismatch: expected %s, got %s",
			expectedAction, result.Action)
	}

	return true, nil
}
