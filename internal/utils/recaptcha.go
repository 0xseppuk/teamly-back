package utils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type RecaptchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ErrorCodes  []string  `json:"error-codes"`
}

func VerifyRecaptcha(token string) (bool, error) {
	secret := GetEnv("RECAPTCHA_SECRET", "")
	if secret == "" {
		log.Println("[reCAPTCHA] WARNING: RECAPTCHA_SECRET is empty!")
		return false, nil
	}

	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify",
		url.Values{
			"secret":   {secret},
			"response": {token},
		})
	if err != nil {
		log.Printf("[reCAPTCHA] HTTP error: %v", err)
		return false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result RecaptchaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[reCAPTCHA] JSON parse error: %v", err)
		return false, err
	}

	// Log full response for debugging
	log.Printf("[reCAPTCHA] Response: success=%v, score=%.2f, action=%s, errors=%v",
		result.Success, result.Score, result.Action, result.ErrorCodes)

	// For reCAPTCHA v3: check both success and score (0.5+ is typically human)
	if !result.Success {
		return false, nil
	}

	// Score threshold: 0.5 is default, adjust if needed
	if result.Score < 0.5 {
		log.Printf("[reCAPTCHA] Low score: %.2f (threshold: 0.5)", result.Score)
		return false, nil
	}

	return true, nil
}
