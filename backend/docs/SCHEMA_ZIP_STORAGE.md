# 📦 Логика сохранения схем алмазной мозаики в ZIP-архивах

## 🎯 Обзор системы

Система создает ZIP-архивы со схемами алмазной мозаики, используя основной `ID` записи изображения для именования архивов. Это обеспечивает быстрый поиск и уникальность имен файлов в S3 MinIO.

## 📁 Структура хранения в S3 MinIO

```
mosaic-bucket/
├── images/
│   └── {image_id}/
│       ├── original.jpg         # Оригинальное изображение пользователя
│       ├── edited.jpg           # Отредактированное изображение
│       ├── processed.jpg        # Обработанное через Stable Diffusion
│       └── preview.jpg          # Превью для веб-интерфейса
└── schemas/
    └── {image_id}.zip           # ZIP-архив схемы (именован по ID изображения)
```

## 📋 Содержимое ZIP-архива

Каждый ZIP-архив содержит:

### 📄 Структура архива
```
{image_id}.zip
├── {image_id}/
│   ├── original.jpg             # Оригинальное изображение пользователя
│   ├── preview.jpg              # Превью готовой мозаики
│   ├── schema.pdf               # PDF-схема алмазной мозаики
│   └── README.txt               # Инструкция по использованию
```

### 🗂️ Описание файлов

| Файл | Описание | Источник |
|------|----------|----------|
| `original.jpg` | Оригинальное изображение, загруженное пользователем | `Image.OriginalImageS3Key` |
| `preview.jpg` | Превью готовой мозаики для визуального контроля | `Image.PreviewS3Key` |
| `schema.pdf` | PDF с подробной схемой алмазной мозаики | Генерируется из обработанного изображения |
| `README.txt` | Инструкция по использованию файлов архива | Автогенерируемый файл |

## 🔄 Процесс создания ZIP-архива

### 1. Инициализация создания схемы
```go
// Вызов метода GenerateSchema
err := imageService.GenerateSchema(ctx, imageID, confirmed)
```

### 2. Сбор файлов для архива
```go
// Скачивание файлов из S3
originalData := s3Client.DownloadFile(ctx, originalImageS3Key)
previewData := s3Client.DownloadFile(ctx, previewS3Key)
schemaData := createMosaicSchemaPDF(ctx, sourceS3Key, imageRecord)
```

### 3. Создание ZIP-архива
```go
// Использование ZipService для создания архива
zipBuffer := zipService.CreateSchemaArchive(imageRecord.ID, files)
```

### 4. Загрузка в S3
```go
// Сохранение с именем по ID изображения
zipS3Key := fmt.Sprintf("schemas/%s.zip", imageRecord.ID.String())
s3Client.UploadFile(ctx, zipBuffer, size, "application/zip", zipS3Key, imageID)
```

### 5. Обновление записи в БД
```go
// Сохранение пути к ZIP-архиву
imageRecord.SchemaS3Key = &zipS3Key
imageRecord.Status = "completed"
```

### 6. Отправка email пользователю
```go
// Асинхронная отправка письма с presigned URL
go func() {
    schemaURL := s3Client.GetFileURL(ctx, schemaS3Key, 30*24*time.Hour)
    emailService.SendSchemaEmail(userEmail, schemaURL, couponCode)
}()
```

## 🗄️ Модель данных

### Поля в таблице `images`
```sql
CREATE TABLE images (
    id UUID PRIMARY KEY,                    -- Основной ID (используется для именования ZIP)
    coupon_id UUID NOT NULL,               -- Связь с купоном
    original_image_s3_key VARCHAR NOT NULL, -- Ключ оригинального изображения
    edited_image_s3_key VARCHAR,           -- Ключ отредактированного изображения
    processed_image_s3_key VARCHAR,        -- Ключ обработанного изображения
    preview_s3_key VARCHAR,                -- Ключ превью изображения
    schema_s3_key VARCHAR,                 -- Ключ ZIP-архива схемы
    processing_params JSON,                -- Параметры обработки
    user_email VARCHAR NOT NULL,           -- Email пользователя
    status processing_status DEFAULT 'queued', -- Статус обработки
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## 🔍 Поиск и доступ к архивам

### Быстрый поиск по ID
```go
// Поиск ZIP-архива по ID изображения
zipS3Key := fmt.Sprintf("schemas/%s.zip", imageID.String())
archiveURL := s3Client.GetFileURL(ctx, zipS3Key, 24*time.Hour)
```

### Валидация имени архива
```go
// Проверка корректности UUID в имени файла
schemaID, err := zipService.ValidateArchiveName("550e8400-e29b-41d4-a716-446655440000.zip")
```

## 📧 Email уведомления

### Содержимое письма
- 🎨 Заголовок: "Ваша схема алмазной мозаики готова!"
- 📥 Кнопка скачивания с presigned URL (действует 30 дней)
- 📋 Описание содержимого ZIP-архива:
  - `schema.pdf` - Подробная схема с цветовой картой
  - `original.jpg` - Оригинальное изображение
  - `preview.jpg` - Превью готовой мозаики
  - `README.txt` - Инструкция по использованию

## 🔐 Безопасность

### Presigned URLs
- ⏰ Срок действия: 30 дней для архивов схем
- 🔒 Приватный доступ через временные ссылки
- 🚫 Нет прямого публичного доступа к файлам

### Именование файлов
- ✅ UUID обеспечивает уникальность
- 🔍 Невозможно угадать имена других архивов
- 📊 Легкое сопоставление с записями в БД

## 🛠️ Техническая реализация

### Основные компоненты
- `ZipService` - создание и управление ZIP-архивами
- `ImageService` - основная бизнес-логика
- `S3Client` - взаимодействие с MinIO
- `EmailService` - отправка уведомлений

### Обработка ошибок
- 🔄 Retry механизм для неудачных операций
- 📝 Подробное логирование всех этапов
- ⚠️ Graceful degradation при недоступности файлов

## 🚀 Преимущества подхода

1. **Производительность**: Быстрый поиск по UUID
2. **Консистентность**: Один архив = один ID изображения
3. **Масштабируемость**: Эффективное использование S3
4. **Безопасность**: Контролируемый доступ через presigned URLs
5. **Удобство**: Все материалы в одном архиве
6. **Трассируемость**: Полная история обработки в БД

## 📈 Мониторинг

### Метрики для отслеживания
- Размер создаваемых ZIP-архивов
- Время создания архивов
- Успешность отправки email уведомлений
- Количество скачиваний по presigned URLs
- Ошибки при создании/загрузке архивов

### Логирование
```go
log.Info().
    Str("image_id", imageID.String()).
    Str("zip_s3_key", zipS3Key).
    Int("files_count", len(files)).
    Msg("Schema ZIP archive created successfully")
```

Эта архитектура обеспечивает надежное, масштабируемое и удобное хранение схем алмазной мозаики с возможностью быстрого поиска и безопасного доступа.