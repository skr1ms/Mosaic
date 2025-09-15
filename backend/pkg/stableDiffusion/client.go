package stableDiffusion

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type StableDiffusionClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *middleware.Logger
}

// Img2ImgRequest structure for img2img API request
type Img2ImgRequest struct {
	InitImages                        []string       `json:"init_images"`                          // Base images in base64
	ResizeMode                        int            `json:"resize_mode"`                          // Resize mode (0-3)
	DenoisingStrength                 float64        `json:"denoising_strength"`                   // Denoising strength (0.0-1.0)
	ImageCfgScale                     float64        `json:"image_cfg_scale"`                      // Image configuration scale
	Mask                              string         `json:"mask,omitempty"`                       // Mask in base64 (optional)
	MaskBlur                          int            `json:"mask_blur"`                            // Mask blur
	InpaintingFill                    int            `json:"inpainting_fill"`                      // Inpainting fill
	InpaintFullRes                    bool           `json:"inpaint_full_res"`                     // Inpainting in full resolution
	InpaintFullResPadding             int            `json:"inpaint_full_res_padding"`             // Padding for full resolution
	InpaintingMaskInvert              int            `json:"inpainting_mask_invert"`               // Mask inversion
	InitialNoiseMultiplier            float64        `json:"initial_noise_multiplier"`             // Initial noise multiplier
	Prompt                            string         `json:"prompt"`                               // Text prompt
	Styles                            []string       `json:"styles,omitempty"`                     // Styles
	Seed                              int64          `json:"seed"`                                 // Generation seed
	Subseed                           int64          `json:"subseed"`                              // Subseed
	SubseedStrength                   float64        `json:"subseed_strength"`                     // Subseed strength
	SeedResizeFromH                   int            `json:"seed_resize_from_h"`                   // Height for seed resize
	SeedResizeFromW                   int            `json:"seed_resize_from_w"`                   // Width for seed resize
	SamplerName                       string         `json:"sampler_name"`                         // Sampler name
	BatchSize                         int            `json:"batch_size"`                           // Batch size
	NIter                             int            `json:"n_iter"`                               // Number of iterations
	Steps                             int            `json:"steps"`                                // Number of steps
	CfgScale                          float64        `json:"cfg_scale"`                            // CFG scale
	Width                             int            `json:"width"`                                // Image width
	Height                            int            `json:"height"`                               // Image height
	RestoreFaces                      bool           `json:"restore_faces"`                        // Face restoration
	Tiling                            bool           `json:"tiling"`                               // Tiling
	DoNotSaveSamples                  bool           `json:"do_not_save_samples"`                  // Don't save samples
	DoNotSaveGrid                     bool           `json:"do_not_save_grid"`                     // Don't save grid
	NegativePrompt                    string         `json:"negative_prompt"`                      // Negative prompt
	Eta                               float64        `json:"eta"`                                  // Eta parameter
	SChurn                            float64        `json:"s_churn"`                              // S-churn parameter
	STmax                             float64        `json:"s_tmax"`                               // S-tmax parameter
	STmin                             float64        `json:"s_tmin"`                               // S-tmin parameter
	SNoise                            float64        `json:"s_noise"`                              // S-noise parameter
	OverrideSettings                  map[string]any `json:"override_settings"`                    // Settings override
	OverrideSettingsRestoreAfterwards bool           `json:"override_settings_restore_afterwards"` // Restore settings after
	ScriptArgs                        []any          `json:"script_args,omitempty"`                // Script arguments
	SamplerIndex                      string         `json:"sampler_index"`                        // Sampler index
	IncludeInitImages                 bool           `json:"include_init_images"`                  // Include source images
	ScriptName                        string         `json:"script_name,omitempty"`                // Script name
	SendImages                        bool           `json:"send_images"`                          // Send images
	SaveImages                        bool           `json:"save_images"`                          // Save images
	AlwaysonScripts                   map[string]any `json:"alwayson_scripts,omitempty"`           // Always active scripts
}

// Img2ImgResponse structure for img2img API response
type Img2ImgResponse struct {
	Images     []string       `json:"images"`     // Generated images in base64
	Parameters map[string]any `json:"parameters"` // Generation parameters
	Info       string         `json:"info"`       // Generation info
}

// ProcessingStyle processing styles according to requirements
type ProcessingStyle string

const (
	StyleGrayscale ProcessingStyle = "grayscale"  // Grayscale
	StyleSkinTones ProcessingStyle = "skin_tones" // Skin tones
	StylePopArt    ProcessingStyle = "pop_art"    // Pop art
	StyleMaxColors ProcessingStyle = "max_colors" // Maximum colors
)

// LightingType lighting types
type LightingType string

const (
	LightingSun   LightingType = "sun"
	LightingMoon  LightingType = "moon"
	LightingVenus LightingType = "venus"
)

// ContrastLevel contrast levels
type ContrastLevel string

const (
	ContrastLow  ContrastLevel = "low"
	ContrastHigh ContrastLevel = "high"
)

// NewStableDiffusionClient creates new client for Stable Diffusion API
func NewStableDiffusionClient(cfg config.StableDiffusionConfig, logger *middleware.Logger) *StableDiffusionClient {
	return &StableDiffusionClient{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
		logger: logger,
	}
}

// ProcessImageRequest image processing parameters
type ProcessImageRequest struct {
	ImageBase64 string          `json:"image_base64"` // Image in base64
	Style       ProcessingStyle `json:"style"`        // Processing style
	UseAI       bool            `json:"use_ai"`       // Use AI processing
	Lighting    LightingType    `json:"lighting"`     // Lighting
	Contrast    ContrastLevel   `json:"contrast"`     // Contrast
	Brightness  float64         `json:"brightness"`   // Brightness (-100 to 100)
	Saturation  float64         `json:"saturation"`   // Saturation (-100 to 100)
	Width       int             `json:"width"`        // Result width
	Height      int             `json:"height"`       // Result height
}

// ProcessImage processes image through Stable Diffusion
func (c *StableDiffusionClient) ProcessImage(ctx context.Context, req ProcessImageRequest) (string, error) {
	prompt := c.buildPrompt(req.Style, req.Lighting, req.UseAI)
	switch req.Contrast {
	case ContrastHigh:
		prompt += ", high contrast"
	case ContrastLow:
		prompt += ", soft contrast"
	}
	if req.Brightness > 0 {
		prompt += ", bright"
	} else if req.Brightness < 0 {
		prompt += ", dim lighting"
	}
	if req.Saturation > 0 {
		prompt += ", vibrant colors"
	} else if req.Saturation < 0 {
		prompt += ", muted colors"
	}

	negativePrompt := c.buildNegativePrompt(req.Style)

	computeDims := func(w, h, maxSide int) (int, int) {
		if w <= 0 || h <= 0 {
			return 512, 512
		}
		fw := float64(w)
		fh := float64(h)
		var nw, nh float64
		if fw >= fh {
			nw = float64(maxSide)
			nh = nw * fh / fw
		} else {
			nh = float64(maxSide)
			nw = nh * fw / fh
		}
		round64 := func(x float64) int {
			v := int(x + 0.5)
			if v < 64 {
				v = 64
			}
			return (v / 64) * 64
		}
		rw := round64(nw)
		rh := round64(nh)
		if rw < 256 {
			rw = 256
		}
		if rh < 256 {
			rh = 256
		}
		return rw, rh
	}

	maxSides := []int{768, 640, 512, 384}
	var lastErr error
	for i, maxSide := range maxSides {
		w, h := computeDims(req.Width, req.Height, maxSide)
		steps := 25
		if i > 0 {
			steps = 20 // reduce steps on retries
		}

		apiRequest := Img2ImgRequest{
			InitImages:        []string{req.ImageBase64},
			Prompt:            prompt,
			NegativePrompt:    negativePrompt,
			Steps:             steps,
			CfgScale:          7.5,
			DenoisingStrength: c.getDenoisingStrength(req.Style, req.UseAI),
			Width:             w,
			Height:            h,
			SamplerName:       "DPM++ 2M Karras",
			BatchSize:         1,
			NIter:             1,
			Seed:              -1,
			RestoreFaces:      true,
			SendImages:        true,
			SaveImages:        false,
			ResizeMode:        1,
			Tiling:            false,
		}

		apiRequest.OverrideSettings = map[string]any{}
		apiRequest.OverrideSettingsRestoreAfterwards = true

		c.logger.GetZerologLogger().Info().Int("attempt", i+1).Int("width", w).Int("height", h).Msg("Stable Diffusion request with safe dimensions")
		img, err := c.makeImg2ImgRequest(ctx, apiRequest)
		if err == nil {
			return img, nil
		}
		lastErr = err
		if !strings.Contains(strings.ToLower(err.Error()), "out of memory") && !strings.Contains(strings.ToLower(err.Error()), "cuda out of memory") {
			break
		}
	}

	return "", lastErr
}

// buildPrompt creates prompt based on style and parameters
func (c *StableDiffusionClient) buildPrompt(style ProcessingStyle, lighting LightingType, useAI bool) string {
	basePrompt := "high quality, detailed, professional"

	switch style {
	case StyleGrayscale:
		basePrompt += ", grayscale, monochrome, black and white, elegant"
	case StyleSkinTones:
		basePrompt += ", natural skin tones, warm colors, portrait, realistic"
	case StylePopArt:
		basePrompt += ", pop art style, vibrant colors, high contrast, bold, artistic"
	case StyleMaxColors:
		basePrompt += ", maximum colors, vibrant, colorful, rich palette, dynamic"
	}

	switch lighting {
	case LightingSun:
		basePrompt += ", bright sunlight, golden hour, warm lighting, natural"
	case LightingMoon:
		basePrompt += ", moonlight, cool lighting, night scene, atmospheric"
	case LightingVenus:
		basePrompt += ", soft diffused light, ethereal lighting, dreamy"
	}

	if useAI {
		basePrompt += ", enhanced, professional quality, masterpiece, refined"
	}

	return basePrompt
}

// buildNegativePrompt creates negative prompt
func (c *StableDiffusionClient) buildNegativePrompt(style ProcessingStyle) string {
	baseNegative := "blurry, low quality, pixelated, artifacts, distorted, watermark, signature, text, ugly, deformed, bad anatomy, disfigured, poorly drawn face, mutated, extra limb, ugly, poorly drawn hands, missing limb, floating limbs, disconnected limbs, malformed hands, blur, out of focus, long neck, long body, mutated hands and fingers, out of frame, blender, doll, cropped, low-res, close-up, poorly-drawn face, out of frame double, two heads, blurred, ugly, disfigured, too many limbs, deformed, repetitive, black and white, grainy, extra limbs, bad anatomy, high pass filter, airbrush, portrait, zoomed, soft light, smooth skin, closeup, deformed, extra limbs, extra fingers, mutated hands, bad anatomy, bad proportions, blind, extra eyes, ugly eyes, dead eyes, blur, vignette, out of shot, out of focus, gaussian, closeup, monochrome, grainy, noisy, text, watermarked, logo, oversaturation, over contrast, over shadow"

	switch style {
	case StyleGrayscale:
		baseNegative += ", colorful, bright colors, saturation"
	case StyleSkinTones:
		baseNegative += ", unnatural skin, green skin, blue skin, purple skin, red skin"
	case StylePopArt:
		baseNegative += ", muted colors, dull, grayscale, monochrome"
	case StyleMaxColors:
		baseNegative += ", monochrome, grayscale, black and white"
	}

	return baseNegative
}

// getDenoisingStrength returns denoising strength depending on style and AI usage
func (c *StableDiffusionClient) getDenoisingStrength(style ProcessingStyle, useAI bool) float64 {
	if !useAI {
		return 0.3 // Light processing to preserve original
	}

	switch style {
	case StyleGrayscale:
		return 0.6 // Medium processing for grayscale
	case StyleSkinTones:
		return 0.6 // Light processing to preserve skin naturalness
	case StylePopArt:
		return 0.6 // Strong processing for pop art
	case StyleMaxColors:
		return 0.6 // Medium-strong processing for maximum colors
	default:
		return 0.6 // Medium processing by default
	}
}

// makeImg2ImgRequest executes request to img2img API
func (c *StableDiffusionClient) makeImg2ImgRequest(ctx context.Context, req Img2ImgRequest) (string, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		c.logger.GetZerologLogger().Error().
			Err(err).
			Msg("Failed to marshal Stable Diffusion request")
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.GetZerologLogger().Error().
			Err(err).
			Str("url", url).
			Msg("Failed to create Stable Diffusion request")
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	c.logger.GetZerologLogger().Info().
		Str("url", url).
		Int("request_size", len(jsonData)).
		Str("prompt", req.Prompt).
		Str("negative_prompt", req.NegativePrompt).
		Float64("denoising_strength", req.DenoisingStrength).
		Int("steps", req.Steps).
		Float64("cfg_scale", req.CfgScale).
		Msg("Making Stable Diffusion API request")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.GetZerologLogger().Error().
			Err(err).
			Str("url", url).
			Msg("Failed to make Stable Diffusion request")
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.GetZerologLogger().Error().
			Int("status_code", resp.StatusCode).
			Str("response_body", string(body)).
			Str("url", url).
			Msg("Stable Diffusion API request failed")
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse Img2ImgResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		c.logger.GetZerologLogger().Error().
			Err(err).
			Msg("Failed to decode Stable Diffusion response")
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResponse.Images) == 0 {
		c.logger.GetZerologLogger().Error().
			Msg("No images returned from Stable Diffusion API")
		return "", fmt.Errorf("no images returned from API")
	}

	c.logger.GetZerologLogger().Info().
		Int("images_count", len(apiResponse.Images)).
		Str("info", apiResponse.Info).
		Int("image_size", len(apiResponse.Images[0])).
		Msg("Stable Diffusion API request completed successfully")

	return apiResponse.Images[0], nil
}

// CheckHealth checks Stable Diffusion API health
func (c *StableDiffusionClient) CheckHealth(ctx context.Context) error {
	healthURL := strings.Replace(c.baseURL, "/img2img", "/samplers", 1)
	if !strings.Contains(healthURL, "/samplers") {
		healthURL = strings.TrimRight(c.baseURL, "/") + "/sdapi/v1/samplers"
	}
	url := healthURL

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.GetZerologLogger().Error().
			Err(err).
			Str("url", url).
			Msg("Failed to create Stable Diffusion health check request")
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.GetZerologLogger().Error().
			Err(err).
			Str("url", url).
			Msg("Stable Diffusion health check failed")
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.GetZerologLogger().Error().
			Int("status_code", resp.StatusCode).
			Str("url", url).
			Msg("Stable Diffusion health check failed with non-OK status")
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	c.logger.GetZerologLogger().Info().Msg("Stable Diffusion API health check passed")
	return nil
}

// DecodeBase64Image decodes base64 image
func (c *StableDiffusionClient) DecodeBase64Image(base64Data string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		c.logger.GetZerologLogger().Error().
			Err(err).
			Msg("Failed to decode base64 image")
		return nil, fmt.Errorf("failed to decode base64 image: %w", err)
	}
	return data, nil
}

// EncodeImageToBase64 encodes image to base64
func (c *StableDiffusionClient) EncodeImageToBase64(imageData []byte) string {
	return base64.StdEncoding.EncodeToString(imageData)
}
