package palette

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/skr1ms/mosaic/pkg/middleware"
)

// PaletteService processes palettes for mosaic generation
type PaletteService struct {
	paletteDir string
	logger     *middleware.Logger
}

// NewPaletteService creates new service for working with palettes
func NewPaletteService(paletteDir string, logger *middleware.Logger) *PaletteService {
	return &PaletteService{
		paletteDir: paletteDir,
		logger:     logger,
	}
}

// Style represents palette style
type Style string

const (
	StyleGrayscale Style = "grayscale"  // Grayscale - pallete_bw.xlsx
	StyleSkinTones Style = "skin_tones" // Skin tones - pallete_fl.xlsx
	StylePopArt    Style = "pop_art"    // Pop art - pallete_tl.xlsx
	StyleMaxColors Style = "max_colors" // Maximum colors - pallete_max.xlsx
)

// GetPalettePath returns path to palette file for specified style
func (ps *PaletteService) GetPalettePath(style Style) (string, error) {

	var filename string

	switch style {
	case StyleGrayscale:
		filename = "pallete_bw.xlsx"
	case StyleSkinTones:
		filename = "pallete_fl.xlsx"
	case StylePopArt:
		filename = "pallete_tl.xlsx"
	case StyleMaxColors:
		filename = "pallete_max.xlsx"
	default:
		ps.logger.GetZerologLogger().Error().Str("style", string(style)).Msg("Unknown palette style requested")
		return "", fmt.Errorf("unknown palette style: %s", style)
	}

	palettePath := filepath.Join(ps.paletteDir, filename)

	if _, err := os.Stat(palettePath); os.IsNotExist(err) {
		ps.logger.GetZerologLogger().Error().Str("path", palettePath).Str("style", string(style)).Msg("Palette file not found")
		return "", fmt.Errorf("palette file not found: %s", filename)
	}

	ps.logger.GetZerologLogger().Info().Str("path", palettePath).Str("style", string(style)).Msg("Palette path resolved successfully")
	return palettePath, nil
}

// ValidateStyle validates palette style correctness
func (ps *PaletteService) ValidateStyle(style string) error {
	switch Style(style) {
	case StyleGrayscale, StyleSkinTones, StylePopArt, StyleMaxColors:
		ps.logger.GetZerologLogger().Info().Str("style", style).Msg("Palette style validated successfully")
		return nil
	default:
		ps.logger.GetZerologLogger().Error().Str("style", style).Msg("Invalid palette style")
		return fmt.Errorf("invalid palette style: %s. Available styles: grayscale, skin_tones, pop_art, max_colors", style)
	}
}

// GetAvailableStyles returns list of available styles
func (ps *PaletteService) GetAvailableStyles() []Style {
	return []Style{
		StyleGrayscale,
		StyleSkinTones,
		StylePopArt,
		StyleMaxColors,
	}
}

// GetStyleDescription returns style description
func (ps *PaletteService) GetStyleDescription(style Style) string {
	switch style {
	case StyleGrayscale:
		return "Classic grayscale processing"
	case StyleSkinTones:
		return "Suitable for portraits, uses skin tone shades"
	case StylePopArt:
		return "Bright saturated colors in pop art style"
	case StyleMaxColors:
		return "Maximum number of shades for detail"
	default:
		return "Unknown style"
	}
}

// GetStyleTitle returns human-readable style title
func (ps *PaletteService) GetStyleTitle(style Style) string {
	switch style {
	case StyleGrayscale:
		return "Grayscale"
	case StyleSkinTones:
		return "Skin Tones"
	case StylePopArt:
		return "Pop Art"
	case StyleMaxColors:
		return "Maximum Colors"
	default:
		return "Unknown Style"
	}
}

// InitializePalettes checks presence of all palette files
func (ps *PaletteService) InitializePalettes() error {
	ps.logger.GetZerologLogger().Info().Str("palette_dir", ps.paletteDir).Msg("Initializing palettes")

	if err := os.MkdirAll(ps.paletteDir, 0755); err != nil {
		ps.logger.GetZerologLogger().Error().Err(err).Str("palette_dir", ps.paletteDir).Msg("Failed to create palette directory")
		return fmt.Errorf("failed to create palette directory: %w", err)
	}

	requiredFiles := map[Style]string{
		StyleGrayscale: "pallete_bw.xlsx",
		StyleSkinTones: "pallete_fl.xlsx",
		StylePopArt:    "pallete_tl.xlsx",
		StyleMaxColors: "pallete_max.xlsx",
	}

	missingFiles := make([]string, 0)

	for style, filename := range requiredFiles {
		palettePath := filepath.Join(ps.paletteDir, filename)
		if _, err := os.Stat(palettePath); os.IsNotExist(err) {
			missingFiles = append(missingFiles, filename)
			ps.logger.GetZerologLogger().Warn().
				Str("style", string(style)).
				Str("file", filename).
				Str("path", palettePath).
				Msg("Palette file not found")
		} else {
			ps.logger.GetZerologLogger().Info().
				Str("style", string(style)).
				Str("file", filename).
				Str("path", palettePath).
				Msg("Palette file found")
		}
	}

	if len(missingFiles) > 0 {
		return fmt.Errorf("missing palette files: %s", strings.Join(missingFiles, ", "))
	}

	ps.logger.GetZerologLogger().Info().Msg("All palette files initialized successfully")
	return nil
}

// CopyPaletteFiles copies palette files from source directory to working directory
func (ps *PaletteService) CopyPaletteFiles(sourceDir string) error {
	if sourceDir == "" {
		sourceDir = "." // Current directory
	}

	requiredFiles := []string{
		"pallete_bw.xlsx",
		"pallete_fl.xlsx",
		"pallete_tl.xlsx",
		"pallete_max.xlsx",
	}

	if err := os.MkdirAll(ps.paletteDir, 0755); err != nil {
		ps.logger.GetZerologLogger().Error().Err(err).Str("palette_dir", ps.paletteDir).Msg("Failed to create palette directory")
		return fmt.Errorf("failed to create palette directory: %w", err)
	}

	for _, filename := range requiredFiles {
		srcPath := filepath.Join(sourceDir, filename)
		dstPath := filepath.Join(ps.paletteDir, filename)

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			ps.logger.GetZerologLogger().Warn().Str("file", srcPath).Msg("Source palette file not found, skipping")
			continue
		}

		if err := ps.copyFile(srcPath, dstPath); err != nil {
			ps.logger.GetZerologLogger().Error().Err(err).Str("src", srcPath).Str("dst", dstPath).Msg("Failed to copy palette file")
			return fmt.Errorf("failed to copy palette file %s: %w", filename, err)
		}

		ps.logger.GetZerologLogger().Info().Str("src", srcPath).Str("dst", dstPath).Msg("Palette file copied successfully")
	}

	return nil
}

// copyFile copies file from src to dst
func (ps *PaletteService) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		ps.logger.GetZerologLogger().Error().Err(err).Str("src", src).Msg("Failed to open source file")
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		ps.logger.GetZerologLogger().Error().Err(err).Str("dst", dst).Msg("Failed to create destination file")
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	if err != nil {
		ps.logger.GetZerologLogger().Error().Err(err).Str("src", src).Str("dst", dst).Msg("Failed to copy file content")
		return err
	}

	sourceInfo, err := os.Stat(src)
	if err != nil {
		ps.logger.GetZerologLogger().Error().Err(err).Str("src", src).Msg("Failed to get source file info")
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// ListPaletteFiles returns list of palette files in directory
func (ps *PaletteService) ListPaletteFiles() (map[Style]string, error) {

	paletteFiles := make(map[Style]string)

	err := filepath.WalkDir(ps.paletteDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		filename := d.Name()
		if !strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
			return nil
		}

		switch filename {
		case "pallete_bw.xlsx":
			paletteFiles[StyleGrayscale] = path
		case "pallete_fl.xlsx":
			paletteFiles[StyleSkinTones] = path
		case "pallete_tl.xlsx":
			paletteFiles[StylePopArt] = path
		case "pallete_max.xlsx":
			paletteFiles[StyleMaxColors] = path
		}

		return nil
	})

	if err != nil {
		ps.logger.GetZerologLogger().Error().Err(err).Str("dir", ps.paletteDir).Msg("Failed to list palette files")
		return nil, fmt.Errorf("failed to list palette files: %w", err)
	}

	ps.logger.GetZerologLogger().Info().Int("count", len(paletteFiles)).Msg("Palette files listed successfully")
	return paletteFiles, nil
}
