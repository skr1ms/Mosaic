package stablediffusion

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/config"
)

type StableDiffusionClient struct {
	baseURL    string
	httpClient *http.Client
}

// Img2ImgRequest структура запроса для img2img API
type Img2ImgRequest struct {
	InitImages                        []string               `json:"init_images"`                          // Базовые изображения в base64
	ResizeMode                        int                    `json:"resize_mode"`                          // Режим изменения размера (0-3)
	DenoisingStrength                 float64                `json:"denoising_strength"`                   // Сила шумоподавления (0.0-1.0)
	ImageCfgScale                     float64                `json:"image_cfg_scale"`                      // Масштаб конфигурации изображения
	Mask                              string                 `json:"mask,omitempty"`                       // Маска в base64 (опционально)
	MaskBlur                          int                    `json:"mask_blur"`                            // Размытие маски
	InpaintingFill                    int                    `json:"inpainting_fill"`                      // Заполнение при инпейнтинге
	InpaintFullRes                    bool                   `json:"inpaint_full_res"`                     // Инпейнтинг в полном разрешении
	InpaintFullResPadding             int                    `json:"inpaint_full_res_padding"`             // Отступы при полном разрешении
	InpaintingMaskInvert              int                    `json:"inpainting_mask_invert"`               // Инвертирование маски
	InitialNoiseMultiplier            float64                `json:"initial_noise_multiplier"`             // Множитель начального шума
	Prompt                            string                 `json:"prompt"`                               // Текстовый промпт
	Styles                            []string               `json:"styles"`                               // Стили
	Seed                              int64                  `json:"seed"`                                 // Сид для генерации
	Subseed                           int64                  `json:"subseed"`                              // Подсид
	SubseedStrength                   float64                `json:"subseed_strength"`                     // Сила подсида
	SeedResizeFromH                   int                    `json:"seed_resize_from_h"`                   // Высота для изменения размера сида
	SeedResizeFromW                   int                    `json:"seed_resize_from_w"`                   // Ширина для изменения размера сида
	SamplerName                       string                 `json:"sampler_name"`                         // Имя семплера
	BatchSize                         int                    `json:"batch_size"`                           // Размер батча
	NIter                             int                    `json:"n_iter"`                               // Количество итераций
	Steps                             int                    `json:"steps"`                                // Количество шагов
	CfgScale                          float64                `json:"cfg_scale"`                            // Масштаб CFG
	Width                             int                    `json:"width"`                                // Ширина изображения
	Height                            int                    `json:"height"`                               // Высота изображения
	RestoreFaces                      bool                   `json:"restore_faces"`                        // Восстановление лиц
	Tiling                            bool                   `json:"tiling"`                               // Тайлинг
	DoNotSaveSamples                  bool                   `json:"do_not_save_samples"`                  // Не сохранять образцы
	DoNotSaveGrid                     bool                   `json:"do_not_save_grid"`                     // Не сохранять сетку
	NegativePrompt                    string                 `json:"negative_prompt"`                      // Негативный промпт
	Eta                               float64                `json:"eta"`                                  // Параметр Eta
	SChurn                            float64                `json:"s_churn"`                              // Параметр S-churn
	STmax                             float64                `json:"s_tmax"`                               // Параметр S-tmax
	STmin                             float64                `json:"s_tmin"`                               // Параметр S-tmin
	SNoise                            float64                `json:"s_noise"`                              // Параметр S-noise
	OverrideSettings                  map[string]interface{} `json:"override_settings"`                    // Переопределение настроек
	OverrideSettingsRestoreAfterwards bool                   `json:"override_settings_restore_afterwards"` // Восстановить настройки после
	ScriptArgs                        []interface{}          `json:"script_args"`                          // Аргументы скрипта
	SamplerIndex                      string                 `json:"sampler_index"`                        // Индекс семплера
	IncludeInitImages                 bool                   `json:"include_init_images"`                  // Включить исходные изображения
	ScriptName                        string                 `json:"script_name,omitempty"`                // Имя скрипта
	SendImages                        bool                   `json:"send_images"`                          // Отправлять изображения
	SaveImages                        bool                   `json:"save_images"`                          // Сохранять изображения
	AlwaysonScripts                   map[string]interface{} `json:"alwayson_scripts"`                     // Постоянно активные скрипты
}

// Img2ImgResponse структура ответа от img2img API
type Img2ImgResponse struct {
	Images     []string               `json:"images"`     // Сгенерированные изображения в base64
	Parameters map[string]interface{} `json:"parameters"` // Параметры генерации
	Info       string                 `json:"info"`       // Информация о генерации
}

// ProcessingStyle стили обработки согласно ТЗ
type ProcessingStyle string

const (
	StyleGrayscale ProcessingStyle = "grayscale"  // Оттенки серого
	StyleSkinTones ProcessingStyle = "skin_tones" // Оттенки телесного
	StylePopArt    ProcessingStyle = "pop_art"    // Поп-арт
	StyleMaxColors ProcessingStyle = "max_colors" // Максимум цветов
)

// LightingType типы освещения
type LightingType string

const (
	LightingSun   LightingType = "sun"
	LightingMoon  LightingType = "moon"
	LightingVenus LightingType = "venus"
)

// ContrastLevel уровни контрастности
type ContrastLevel string

const (
	ContrastLow  ContrastLevel = "low"
	ContrastHigh ContrastLevel = "high"
)

// NewStableDiffusionClient создает новый клиент для Stable Diffusion API
func NewStableDiffusionClient(cfg config.StableDiffusionConfig) *StableDiffusionClient {
	return &StableDiffusionClient{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Увеличиваем таймаут для AI обработки
		},
	}
}

// ProcessImageRequest параметры обработки изображения
type ProcessImageRequest struct {
	ImageBase64 string          `json:"image_base64"` // Изображение в base64
	Style       ProcessingStyle `json:"style"`        // Стиль обработки
	UseAI       bool            `json:"use_ai"`       // Использовать AI обработку
	Lighting    LightingType    `json:"lighting"`     // Освещение
	Contrast    ContrastLevel   `json:"contrast"`     // Контрастность
	Brightness  float64         `json:"brightness"`   // Яркость (-100 до 100)
	Saturation  float64         `json:"saturation"`   // Насыщенность (-100 до 100)
	Width       int             `json:"width"`        // Ширина результата
	Height      int             `json:"height"`       // Высота результата
}

// ProcessImage обрабатывает изображение через Stable Diffusion
func (c *StableDiffusionClient) ProcessImage(ctx context.Context, req ProcessImageRequest) (string, error) {
	// Формируем промпт на основе стиля и параметров
	prompt := c.buildPrompt(req.Style, req.Lighting, req.UseAI)
	negativePrompt := c.buildNegativePrompt(req.Style)

	// Создаем запрос к API
	apiRequest := Img2ImgRequest{
		InitImages:        []string{req.ImageBase64},
		Prompt:            prompt,
		NegativePrompt:    negativePrompt,
		Steps:             20,
		CfgScale:          7.0,
		DenoisingStrength: c.getDenoisingStrength(req.UseAI),
		Width:             req.Width,
		Height:            req.Height,
		SamplerName:       "DPM++ 2M Karras",
		BatchSize:         1,
		NIter:             1,
		Seed:              -1, // Случайный сид
		RestoreFaces:      true,
		SendImages:        true,
		SaveImages:        false,
		ResizeMode:        1, // Crop and resize
	}

	// Применяем дополнительные настройки яркости и контрастности
	apiRequest.OverrideSettings = c.buildOverrideSettings(req.Brightness, req.Saturation, req.Contrast)
	apiRequest.OverrideSettingsRestoreAfterwards = true

	// Выполняем запрос
	return c.makeImg2ImgRequest(ctx, apiRequest)
}

// buildPrompt создает промпт на основе стиля и параметров
func (c *StableDiffusionClient) buildPrompt(style ProcessingStyle, lighting LightingType, useAI bool) string {
	basePrompt := "diamond painting mosaic, high quality, detailed"

	switch style {
	case StyleGrayscale:
		basePrompt += ", grayscale, monochrome, black and white"
	case StyleSkinTones:
		basePrompt += ", warm skin tones, natural colors, portrait"
	case StylePopArt:
		basePrompt += ", pop art style, vibrant colors, high contrast, bold"
	case StyleMaxColors:
		basePrompt += ", maximum colors, vibrant, colorful, rich palette"
	}

	switch lighting {
	case LightingSun:
		basePrompt += ", bright sunlight, golden hour, warm lighting"
	case LightingMoon:
		basePrompt += ", moonlight, cool lighting, night scene"
	case LightingVenus:
		basePrompt += ", soft diffused light, ethereal lighting"
	}

	if useAI {
		basePrompt += ", AI enhanced, professional quality, masterpiece"
	}

	return basePrompt
}

// buildNegativePrompt создает негативный промпт
func (c *StableDiffusionClient) buildNegativePrompt(style ProcessingStyle) string {
	baseNegative := "blurry, low quality, pixelated, artifacts, distorted"

	switch style {
	case StyleGrayscale:
		baseNegative += ", colorful, bright colors"
	case StyleSkinTones:
		baseNegative += ", unnatural skin, green skin, blue skin"
	case StylePopArt:
		baseNegative += ", muted colors, dull"
	case StyleMaxColors:
		baseNegative += ", monochrome, grayscale"
	}

	return baseNegative
}

// getDenoisingStrength возвращает силу шумоподавления в зависимости от использования AI
func (c *StableDiffusionClient) getDenoisingStrength(useAI bool) float64 {
	if useAI {
		return 0.7 // Более сильная обработка для AI
	}
	return 0.3 // Легкая обработка для сохранения оригинала
}

// buildOverrideSettings создает настройки переопределения
func (c *StableDiffusionClient) buildOverrideSettings(brightness, saturation float64, contrast ContrastLevel) map[string]interface{} {
	settings := make(map[string]interface{})

	// Настройки яркости и насыщенности (преобразуем из диапазона -100,100 в 0,2)
	if brightness != 0 {
		settings["brightness"] = 1.0 + (brightness / 100.0)
	}
	if saturation != 0 {
		settings["saturation"] = 1.0 + (saturation / 100.0)
	}

	// Настройки контрастности
	switch contrast {
	case ContrastLow:
		settings["contrast"] = 0.8
	case ContrastHigh:
		settings["contrast"] = 1.2
	}

	return settings
}

// makeImg2ImgRequest выполняет запрос к img2img API
func (c *StableDiffusionClient) makeImg2ImgRequest(ctx context.Context, req Img2ImgRequest) (string, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/sdapi/v1/img2img", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	log.Info().
		Str("url", url).
		Int("request_size", len(jsonData)).
		Msg("Making Stable Diffusion API request")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse Img2ImgResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResponse.Images) == 0 {
		return "", fmt.Errorf("no images returned from API")
	}

	log.Info().
		Int("images_count", len(apiResponse.Images)).
		Str("info", apiResponse.Info).
		Msg("Stable Diffusion API request completed")

	return apiResponse.Images[0], nil
}

// DecodeBase64Image декодирует base64 изображение
func (c *StableDiffusionClient) DecodeBase64Image(base64Data string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image: %w", err)
	}
	return data, nil
}

// EncodeImageToBase64 кодирует изображение в base64
func (c *StableDiffusionClient) EncodeImageToBase64(imageData []byte) string {
	return base64.StdEncoding.EncodeToString(imageData)
}
