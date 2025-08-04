package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

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
	// ├── preview.jpg      - превью
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

	// Добавляем файл README с информацией о схеме
	readmeContent := z.generateReadmeContent(schemaID)
	readmeWriter, err := zipWriter.Create(filepath.Join(baseDir, "README.txt"))
	if err != nil {
		zipWriter.Close()
		return nil, fmt.Errorf("failed to create README.txt: %w", err)
	}
	_, err = readmeWriter.Write([]byte(readmeContent))
	if err != nil {
		zipWriter.Close()
		return nil, fmt.Errorf("failed to write README.txt: %w", err)
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf, nil
}

// generateReadmeContent генерирует содержимое README файла
func (z *ZipService) generateReadmeContent(schemaID uuid.UUID) string {
	return fmt.Sprintf(`СХЕМА АЛМАЗНОЙ МОЗАИКИ
======================

ID схемы: %s

Содержимое архива:
- original.jpg  - Ваше оригинальное изображение
- preview.jpg   - Превью готовой мозаики
- schema.pdf    - Схема для выкладки алмазных камней

Инструкция по использованию:
1. Откройте файл schema.pdf
2. Следуйте цветовой схеме для размещения камней
3. Используйте preview.jpg как референс готового результата

Поддержка: support@mosaic.com

Дата создания: %s
`, schemaID.String(), time.Now().Format("2006-01-02 15:04:05"))
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
