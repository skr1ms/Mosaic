package mosaic

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type MosaicGenerator struct {
	ScriptPath    string
	OutputDir     string
	PythonCommand string
	logger        *middleware.Logger
}

type GenerationRequest struct {
	ImagePath   string
	StonesX     int
	StonesY     int
	StoneSizeMM float64
	DPI         int
	PreviewDPI  int
	SchemeDPI   int
	Mode        string
	Style       string
	WithLegend  bool
	Threads     int
	PalettePath string
}

type GenerationResult struct {
	PreviewPath string
	SchemePath  string
	LegendPath  string
	ZipPath     string
	SchemaUUID  string
}

func NewMosaicGenerator(scriptPath, outputDir, pythonCommand string, logger *middleware.Logger) *MosaicGenerator {
	return &MosaicGenerator{
		ScriptPath:    scriptPath,
		OutputDir:     outputDir,
		PythonCommand: pythonCommand,
		logger:        logger,
	}
}

func (mg *MosaicGenerator) Generate(ctx context.Context, req *GenerationRequest) (*GenerationResult, error) {
	mg.logger.GetZerologLogger().Info().
		Str("mode", req.Mode).
		Str("style", req.Style).
		Int("stones_x", req.StonesX).
		Int("stones_y", req.StonesY).
		Str("palette_path", req.PalettePath).
		Msg("Starting mosaic generation")

	if mg.OutputDir != "" {
		if err := os.MkdirAll(mg.OutputDir, 0755); err != nil {
			mg.logger.GetZerologLogger().Error().Err(err).Str("output_dir", mg.OutputDir).Msg("Failed to ensure output base directory")
			return nil, fmt.Errorf("failed to ensure output base directory: %w", err)
		}
	}

	outputDir, err := os.MkdirTemp(mg.OutputDir, "mosaic_*")
	if err != nil {
		mg.logger.GetZerologLogger().Error().Err(err).Str("output_dir", mg.OutputDir).Msg("Failed to create output directory")
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	schemaUUID := uuid.New().String()

	if _, err := os.Stat(mg.ScriptPath); os.IsNotExist(err) {
		mg.logger.GetZerologLogger().Error().Str("script_path", mg.ScriptPath).Msg("Python script not found")
		return nil, fmt.Errorf("python script not found at path: %s", mg.ScriptPath)
	}

	if req.PalettePath != "" {
		if _, err := os.Stat(req.PalettePath); os.IsNotExist(err) {
			mg.logger.GetZerologLogger().Error().Str("palette_path", req.PalettePath).Msg("Palette file not found")
			return nil, fmt.Errorf("palette file not found at path: %s", req.PalettePath)
		}
	}

	args := mg.buildPythonArgs(req)

	cmd := exec.CommandContext(ctx, mg.PythonCommand, args...)
	cmd.Dir = outputDir

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			mg.logger.GetZerologLogger().Error().Err(err).Str("stderr", stderrBuf.String()).Msg("Python script execution failed")
			return nil, fmt.Errorf("python script execution failed: %w: %s", err, stderrBuf.String())
		}
	case <-time.After(10 * time.Minute):
		cmd.Process.Kill()
		mg.logger.GetZerologLogger().Error().Msg("Mosaic generation timed out")
		return nil, fmt.Errorf("mosaic generation timed out")
	}

	result := &GenerationResult{
		SchemaUUID: schemaUUID,
	}

	previewPattern := filepath.Join(outputDir, "*preview*.png")
	if matches, err := filepath.Glob(previewPattern); err == nil && len(matches) > 0 {
		result.PreviewPath = matches[0]
	}

	schemePattern := filepath.Join(outputDir, "*scheme*.png")
	if matches, err := filepath.Glob(schemePattern); err == nil && len(matches) > 0 {
		result.SchemePath = matches[0]
	}

	legendPattern := filepath.Join(outputDir, "*legend*.csv")
	if matches, err := filepath.Glob(legendPattern); err == nil && len(matches) > 0 {
		result.LegendPath = matches[0]
	}

	zipPath, err := mg.createZipArchive(result, outputDir, schemaUUID)
	if err != nil {
		mg.logger.GetZerologLogger().Error().
			Err(err).
			Str("stderr", stderrBuf.String()).
			Msg("Failed to create ZIP archive")
		return nil, fmt.Errorf("failed to create zip archive: %w", err)
	} else {
		result.ZipPath = zipPath
	}

	return result, nil
}

func (mg *MosaicGenerator) buildPythonArgs(req *GenerationRequest) []string {
	args := []string{mg.ScriptPath}

	imgPath := req.ImagePath
	if imgPath != "" && !filepath.IsAbs(imgPath) {
		if abs, err := filepath.Abs(imgPath); err == nil {
			imgPath = abs
		}
	}
	args = append(args, imgPath)

	if req.PalettePath != "" {
		palettePath := req.PalettePath
		if !filepath.IsAbs(palettePath) {
			if abs, err := filepath.Abs(palettePath); err == nil {
				palettePath = abs
			}
		}
		args = append(args, palettePath)
	}

	args = append(args, "--stones-x", strconv.Itoa(req.StonesX))
	args = append(args, "--stones-y", strconv.Itoa(req.StonesY))

	args = append(args, "--stone-size-mm", fmt.Sprintf("%.2f", req.StoneSizeMM))

	if req.DPI > 0 {
		args = append(args, "--dpi", strconv.Itoa(req.DPI))
	}
	if req.PreviewDPI > 0 {
		args = append(args, "--preview-dpi", strconv.Itoa(req.PreviewDPI))
	}
	if req.SchemeDPI > 0 {
		args = append(args, "--scheme-dpi", strconv.Itoa(req.SchemeDPI))
	}

	if req.Mode != "" {
		args = append(args, "--mode", req.Mode)
	}

	if req.Style != "" {
		args = append(args, "--style", req.Style)
	}

	if req.WithLegend {
		args = append(args, "--legend")
	}

	if req.Threads > 0 {
		args = append(args, "--threads", strconv.Itoa(req.Threads))
	}

	return args
}

func (mg *MosaicGenerator) createZipArchive(result *GenerationResult, outputDir string, schemaUUID string) (string, error) {
	zipPath := filepath.Join(outputDir, schemaUUID+".zip")

	zipFile, err := os.Create(zipPath)
	if err != nil {
		mg.logger.GetZerologLogger().Error().Err(err).Str("zip_path", zipPath).Msg("Failed to create zip file")
		return "", fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	filesToZip := map[string]string{
		"preview.png": result.PreviewPath,
		"scheme.png":  result.SchemePath,
		"legend.csv":  result.LegendPath,
	}

	for archiveName, filePath := range filesToZip {
		if filePath == "" {
			continue
		}

		if err := mg.addFileToZip(zipWriter, filePath, archiveName); err != nil {
			continue
		}
	}

	return zipPath, nil
}

func (mg *MosaicGenerator) addFileToZip(zipWriter *zip.Writer, filePath, archiveName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		mg.logger.GetZerologLogger().Error().Err(err).Str("file_path", filePath).Msg("Failed to open file for zip")
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	zipFileWriter, err := zipWriter.Create(archiveName)
	if err != nil {
		mg.logger.GetZerologLogger().Error().Err(err).Str("archive_name", archiveName).Msg("Failed to create file in zip")
		return fmt.Errorf("failed to create file in zip: %w", err)
	}

	_, err = io.Copy(zipFileWriter, file)
	if err != nil {
		mg.logger.GetZerologLogger().Error().Err(err).Str("file_path", filePath).Str("archive_name", archiveName).Msg("Failed to copy file to zip")
		return fmt.Errorf("failed to copy file to zip: %w", err)
	}

	return nil
}

func (mg *MosaicGenerator) Cleanup(result *GenerationResult) error {
	var errors []error

	filesToClean := []string{
		result.PreviewPath,
		result.SchemePath,
		result.LegendPath,
		result.ZipPath,
	}

	for _, filePath := range filesToClean {
		if filePath != "" {
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				mg.logger.GetZerologLogger().Warn().
					Err(err).
					Str("filePath", filePath).
					Msg("Failed to cleanup file")
				errors = append(errors, err)
			} else {
				mg.logger.GetZerologLogger().Debug().
					Str("filePath", filePath).
					Msg("Successfully cleaned up file")
			}
		}
	}

	if len(errors) > 0 {
		mg.logger.GetZerologLogger().Error().Int("error_count", len(errors)).Msg("Cleanup completed with errors")
		return fmt.Errorf("cleanup completed with %d errors", len(errors))
	}

	return nil
}
