package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type ZipService struct{}

func NewZipService() *ZipService {
	return &ZipService{}
}

// FileData представляет файл для добавления в архив
type FileData struct {
	Name    string    // Имя файла в архиве
	Content io.Reader // Содержимое файла
	Size    int64     // Размер файла
}

// CreateSchemaArchive создает ZIP архив с файлами схемы алмазной мозаики
func (z *ZipService) CreateSchemaArchive(schemaID uuid.UUID, files []FileData) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Создаем структуру архива:
	// {schema_uuid}/
	// ├── original.jpg     - оригинальное изображение
	// ├── preview.jpg      - превью мозаики
	// └── schema.pdf       - схема алмазной мозаики

	baseDir := schemaID.String()

	for _, file := range files {
		// Создаем файл в архиве
		fileName := filepath.Join(baseDir, file.Name)
		fileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			zipWriter.Close()
			return nil, fmt.Errorf("failed to create file %s in archive: %w", fileName, err)
		}

		// Копируем содержимое файла
		_, err = io.Copy(fileWriter, file.Content)
		if err != nil {
			zipWriter.Close()
			return nil, fmt.Errorf("failed to write file %s to archive: %w", fileName, err)
		}
	}

	err := zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf, nil
}

// ValidateArchiveName проверяет, что имя архива соответствует UUID формату
func (z *ZipService) ValidateArchiveName(archiveName string) (uuid.UUID, error) {
	// Убираем расширение .zip если есть
	name := strings.TrimSuffix(archiveName, ".zip")

	// Парсим UUID
	schemaID, err := uuid.Parse(name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid archive name format, expected UUID: %w", err)
	}

	return schemaID, nil
}

// GetArchiveName возвращает имя архива для схемы
func (z *ZipService) GetArchiveName(schemaID uuid.UUID) string {
	return fmt.Sprintf("%s.zip", schemaID.String())
}

// ExtractArchiveFiles извлекает файлы из ZIP архива
func (z *ZipService) ExtractArchiveFiles(archiveData io.ReaderAt, size int64) (map[string][]byte, error) {
	reader, err := zip.NewReader(archiveData, size)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip reader: %w", err)
	}

	files := make(map[string][]byte)

	for _, file := range reader.File {
		// Пропускаем директории
		if file.FileInfo().IsDir() {
			continue
		}

		// Открываем файл
		fileReader, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", file.Name, err)
		}

		// Читаем содержимое
		content, err := io.ReadAll(fileReader)
		fileReader.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file.Name, err)
		}

		// Сохраняем содержимое (используем только имя файла без пути)
		fileName := filepath.Base(file.Name)
		files[fileName] = content
	}

	return files, nil
}
