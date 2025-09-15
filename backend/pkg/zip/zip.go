package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type ZipService struct {
	logger *middleware.Logger
}

func NewZipService(logger *middleware.Logger) *ZipService {
	return &ZipService{
		logger: logger,
	}
}

type FileData struct {
	Name    string
	Content io.Reader
	Size    int64
}

func (z *ZipService) CreateSchemaArchive(schemaID uuid.UUID, files []FileData) (*bytes.Buffer, error) {
	z.logger.GetZerologLogger().Info().
		Str("schema_id", schemaID.String()).
		Int("files_count", len(files)).
		Msg("Creating schema archive")

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	baseDir := schemaID.String()

	for _, file := range files {
		fileName := filepath.Join(baseDir, file.Name)
		fileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			z.logger.GetZerologLogger().Error().Err(err).Str("file_name", fileName).Msg("Failed to create file in archive")
			zipWriter.Close()
			return nil, fmt.Errorf("failed to create file %s in archive: %w", fileName, err)
		}

		_, err = io.Copy(fileWriter, file.Content)
		if err != nil {
			z.logger.GetZerologLogger().Error().Err(err).Str("file_name", fileName).Msg("Failed to write file to archive")
			zipWriter.Close()
			return nil, fmt.Errorf("failed to write file %s to archive: %w", fileName, err)
		}
	}

	err := zipWriter.Close()
	if err != nil {
		z.logger.GetZerologLogger().Error().
			Err(err).
			Msg("Failed to close zip writer")
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf, nil
}

func (z *ZipService) ValidateArchiveName(archiveName string) (uuid.UUID, error) {
	name := strings.TrimSuffix(archiveName, ".zip")

	schemaID, err := uuid.Parse(name)
	if err != nil {
		z.logger.GetZerologLogger().Error().Err(err).Str("archive_name", archiveName).Msg("Invalid archive name format, expected UUID")
		return uuid.Nil, fmt.Errorf("invalid archive name format, expected UUID: %w", err)
	}

	return schemaID, nil
}

func (z *ZipService) GetArchiveName(schemaID uuid.UUID) string {
	return fmt.Sprintf("%s.zip", schemaID.String())
}

func (z *ZipService) ExtractArchiveFiles(archiveData io.ReaderAt, size int64) (map[string][]byte, error) {
	reader, err := zip.NewReader(archiveData, size)
	if err != nil {
		z.logger.GetZerologLogger().Error().Err(err).Int64("size", size).Msg("Failed to open zip reader")
		return nil, fmt.Errorf("failed to open zip reader: %w", err)
	}

	files := make(map[string][]byte)

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			z.logger.GetZerologLogger().Error().Err(err).Str("file_name", file.Name).Msg("Failed to open file in archive")
			return nil, fmt.Errorf("failed to open file %s: %w", file.Name, err)
		}

		content, err := io.ReadAll(fileReader)
		fileReader.Close()
		if err != nil {
			z.logger.GetZerologLogger().Error().Err(err).Str("file_name", file.Name).Msg("Failed to read file from archive")
			return nil, fmt.Errorf("failed to read file %s: %w", file.Name, err)
		}

		// Save content (use only filename without path)
		fileName := filepath.Base(file.Name)
		files[fileName] = content
	}

	return files, nil
}
