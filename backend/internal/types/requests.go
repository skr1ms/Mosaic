package types

// Общие структуры для всех модулей, чтобы избежать циклических импортов

// EditImageRequest - запрос редактирования изображения
type EditImageRequest struct {
	CropX      int     `json:"crop_x"`      // X координата начала кадрирования
	CropY      int     `json:"crop_y"`      // Y координата начала кадрирования
	CropWidth  int     `json:"crop_width"`  // Ширина области кадрирования
	CropHeight int     `json:"crop_height"` // Высота области кадрирования
	Rotation   int     `json:"rotation"`    // Поворот в градусах (0, 90, 180, 270)
	Scale      float64 `json:"scale"`       // Масштаб (0.1 - 5.0)
}

// ProcessImageRequest - запрос обработки изображения
type ProcessImageRequest struct {
	Style       string                 `json:"style" validate:"required,oneof=grayscale skin_tones pop_art max_colors"`
	UseAI       bool                   `json:"use_ai"`       // Использовать AI обработку
	Lighting    string                 `json:"lighting"`     // sun, moon, venus
	Contrast    string                 `json:"contrast"`     // low, high
	Brightness  float64                `json:"brightness"`   // -100 до 100
	Saturation  float64                `json:"saturation"`   // -100 до 100
	Settings    map[string]interface{} `json:"settings"`     // Дополнительные настройки
}

// GenerateSchemaRequest - запрос создания схемы
type GenerateSchemaRequest struct {
	Confirmed bool `json:"confirmed" validate:"required"` // Подтверждение создания схемы
}
