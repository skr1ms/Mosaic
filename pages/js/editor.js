// Глобальные переменные редактора
let currentStep = 1;
let totalSteps = 4;
let currentImage = null;
let canvas = null;
let ctx = null;
let previewCanvas = null;
let previewCtx = null;
let finalCanvas = null;
let finalCtx = null;
let zoomLevel = 1;
let rotation = 0;
let selectedStyle = 'original';
let couponData = null;

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    initializeEditor();
    setupEditorEventListeners();
    loadCouponData();
});

// Инициализация редактора
function initializeEditor() {
    // Получаем купон из URL
    const urlParams = new URLSearchParams(window.location.search);
    const couponCode = urlParams.get('coupon');
    const devMode = urlParams.get('dev') === 'true';
    
    if (!couponCode && !devMode) {
        showNotification('Купон не найден', 'error');
        setTimeout(() => {
            window.location.href = 'index.html';
        }, 2000);
        return;
    }
    
    // Инициализируем canvas
    canvas = document.getElementById('image-canvas');
    ctx = canvas.getContext('2d');
    
    previewCanvas = document.getElementById('preview-canvas');
    previewCtx = previewCanvas.getContext('2d');
    
    finalCanvas = document.getElementById('final-canvas');
    finalCtx = finalCanvas.getContext('2d');
    
    // Устанавливаем размеры canvas
    setupCanvasSizes();
    
    // Инициализируем отображение этапов
    updateStepDisplay();
}

// Настройка обработчиков событий
function setupEditorEventListeners() {
    // Загрузка файла
    const fileInput = document.getElementById('image-upload');
    const uploadArea = document.querySelector('.upload-area');
    
    fileInput.addEventListener('change', handleFileSelect);
    
    // Drag & Drop
    uploadArea.addEventListener('dragover', handleDragOver);
    uploadArea.addEventListener('dragleave', handleDragLeave);
    uploadArea.addEventListener('drop', handleDrop);
    uploadArea.addEventListener('click', () => fileInput.click());
    
    // Инструменты редактирования
    document.querySelectorAll('.tool-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            const action = this.dataset.action;
            handleToolAction(action);
        });
    });
    
    // Выбор стилей
    document.querySelectorAll('.style-option').forEach(option => {
        option.addEventListener('click', function() {
            selectStyle(this.dataset.style);
        });
    });
    
    // Настройки мозаики
    document.getElementById('lighting-setting').addEventListener('change', updatePreview);
    document.getElementById('contrast-setting').addEventListener('change', updatePreview);
    
    // Кнопки навигации
    const prevBtn = document.getElementById('prev-step');
    const nextBtn = document.getElementById('next-step');
    
    if (prevBtn) {
        prevBtn.addEventListener('click', previousStep);
    }
    
    if (nextBtn) {
        nextBtn.addEventListener('click', nextStep);
    }
    
    // Обработчик изменения размера окна
    window.addEventListener('resize', resizeCanvas);
}

// Загрузка данных купона
function loadCouponData() {
    const urlParams = new URLSearchParams(window.location.search);
    const couponCode = urlParams.get('coupon');
    const devMode = urlParams.get('dev') === 'true';
    
    // Здесь будет API запрос для получения данных купона
    // Пока используем заглушку
    if (devMode) {
        couponData = {
            number: 'DEMO-123456',
            size: '40×50 см',
            style: 'Оттенки серого'
        };
    } else {
        couponData = {
            number: couponCode,
            size: '40×50 см',
            style: 'Оттенки серого'
        };
    }
    
    displayCouponInfo();
}

// Отображение информации о купоне
function displayCouponInfo() {
    if (!couponData) return;
    
    document.getElementById('coupon-number').textContent = couponData.number;
    document.getElementById('coupon-size').textContent = couponData.size;
    document.getElementById('coupon-style').textContent = couponData.style;
    
    // Показываем индикатор режима разработки, если он активен
    const urlParams = new URLSearchParams(window.location.search);
    const devMode = urlParams.get('dev') === 'true';
    
    const devIndicator = document.getElementById('dev-mode-indicator');
    if (devMode && devIndicator) {
        devIndicator.style.display = 'block';
    }
}

// Настройка размеров canvas
function setupCanvasSizes() {
    const container = document.querySelector('.canvas-container');
    const previewContainer = document.querySelector('.preview-container');
    const finalContainer = document.querySelector('.final-image');
    
    if (container) {
        canvas.width = container.clientWidth - 40;
        canvas.height = container.clientHeight - 40;
    }
    
    if (previewContainer) {
        previewCanvas.width = previewContainer.clientWidth - 40;
        previewCanvas.height = previewContainer.clientHeight - 40;
    }
    
    if (finalContainer) {
        finalCanvas.width = finalContainer.clientWidth - 40;
        finalCanvas.height = finalContainer.clientHeight - 40;
    }
}

// Обновление размеров canvas при изменении размера окна
function resizeCanvas() {
    setupCanvasSizes();
    if (currentImage) {
        drawImage();
    }
}

// Обработка выбора файла
function handleFileSelect(event) {
    const file = event.target.files[0];
    if (file) {
        // Проверяем тип файла
        const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
        if (!allowedTypes.includes(file.type)) {
            showNotification('Поддерживаются только изображения: JPG, PNG, GIF, WebP', 'error');
            return;
        }
        
        loadImage(file);
    }
    
    // Очищаем input для возможности повторной загрузки того же файла
    event.target.value = '';
}

// Обработка Drag & Drop
function handleDragOver(event) {
    event.preventDefault();
    event.currentTarget.classList.add('dragover');
}

function handleDragLeave(event) {
    event.currentTarget.classList.remove('dragover');
}

function handleDrop(event) {
    event.preventDefault();
    event.currentTarget.classList.remove('dragover');
    
    const files = event.dataTransfer.files;
    if (files.length > 0) {
        const file = files[0];
        
        // Проверяем тип файла
        const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
        if (!allowedTypes.includes(file.type)) {
            showNotification('Поддерживаются только изображения: JPG, PNG, GIF, WebP', 'error');
            return;
        }
        
        loadImage(file);
    }
}

// Загрузка изображения
function loadImage(file) {
    if (!file.type.startsWith('image/')) {
        showNotification('Пожалуйста, выберите изображение', 'error');
        return;
    }
    
    // Проверяем размер файла (максимум 10MB)
    if (file.size > 10 * 1024 * 1024) {
        showNotification('Размер файла слишком большой. Максимум 10MB', 'error');
        return;
    }
    
    // Показываем индикатор загрузки
    showNotification('Загружаем изображение...', 'info');
    
    const reader = new FileReader();
    
    reader.onload = function(e) {
        const img = new Image();
        
        img.onload = function() {
            // Сбрасываем масштаб и поворот для нового изображения
            zoomLevel = 1;
            rotation = 0;
            
            currentImage = img;
            
            // Обновляем размеры canvas перед отрисовкой
            setupCanvasSizes();
            
            // Отрисовываем изображение
            drawImage();
            
            // Обновляем отображение масштаба
            updateZoomDisplay();
            
            // Переходим к следующему этапу
            nextStep();
            
            showNotification('Изображение успешно загружено!', 'success');
        };
        
        img.onerror = function() {
            showNotification('Ошибка при загрузке изображения', 'error');
        };
        
        img.src = e.target.result;
    };
    
    reader.onerror = function() {
        showNotification('Ошибка при чтении файла', 'error');
    };
    
    reader.readAsDataURL(file);
}

// Отрисовка изображения
function drawImage() {
    if (!currentImage || !canvas) return;
    
    const canvasWidth = canvas.width;
    const canvasHeight = canvas.height;
    
    // Очищаем canvas
    ctx.clearRect(0, 0, canvasWidth, canvasHeight);
    
    // Устанавливаем качество рендеринга
    ctx.imageSmoothingEnabled = true;
    ctx.imageSmoothingQuality = 'high';
    
    // Вычисляем размеры для отображения с сохранением пропорций
    const imgAspect = currentImage.width / currentImage.height;
    const canvasAspect = canvasWidth / canvasHeight;
    
    let drawWidth, drawHeight, offsetX, offsetY;
    
    // Вычисляем базовые размеры (без масштабирования)
    let baseWidth, baseHeight;
    
    if (imgAspect > canvasAspect) {
        // Изображение шире, чем canvas
        baseWidth = canvasWidth;
        baseHeight = canvasWidth / imgAspect;
        offsetX = 0;
        offsetY = (canvasHeight - baseHeight) / 2;
    } else {
        // Изображение выше, чем canvas
        baseHeight = canvasHeight;
        baseWidth = canvasHeight * imgAspect;
        offsetX = (canvasWidth - baseWidth) / 2;
        offsetY = 0;
    }
    
    // Применяем масштабирование
    drawWidth = baseWidth * zoomLevel;
    drawHeight = baseHeight * zoomLevel;
    
    // Корректируем смещение при масштабировании
    offsetX += (baseWidth - drawWidth) / 2;
    offsetY += (baseHeight - drawHeight) / 2;
    
    // Сохраняем контекст
    ctx.save();
    
    // Перемещаем в центр и поворачиваем
    ctx.translate(canvasWidth / 2, canvasHeight / 2);
    ctx.rotate(rotation * Math.PI / 180);
    ctx.translate(-canvasWidth / 2, -canvasHeight / 2);
    
    // Рисуем изображение
    ctx.drawImage(currentImage, offsetX, offsetY, drawWidth, drawHeight);
    
    // Восстанавливаем контекст
    ctx.restore();
}

// Обработка действий инструментов
function handleToolAction(action) {
    switch (action) {
        case 'crop':
            toggleCropMode();
            break;
        case 'reset-crop':
            resetCrop();
            break;
        case 'zoom-in':
            zoomIn();
            break;
        case 'zoom-out':
            zoomOut();
            break;
        case 'rotate-left':
            rotate(-90);
            break;
        case 'rotate-right':
            rotate(90);
            break;
        case 'rotate-180':
            rotate(180);
            break;
    }
}

// Переключение режима кадрирования
function toggleCropMode() {
    const cropOverlay = document.getElementById('crop-overlay');
    const isVisible = cropOverlay.style.display === 'block';
    
    if (isVisible) {
        cropOverlay.style.display = 'none';
        document.querySelector('[data-action="crop"]').classList.remove('active');
    } else {
        cropOverlay.style.display = 'block';
        document.querySelector('[data-action="crop"]').classList.add('active');
        setupCropOverlay();
    }
}

// Настройка области кадрирования
function setupCropOverlay() {
    const cropOverlay = document.getElementById('crop-overlay');
    const container = document.querySelector('.canvas-container');
    
    const size = Math.min(container.clientWidth, container.clientHeight) * 0.8;
    const left = (container.clientWidth - size) / 2;
    const top = (container.clientHeight - size) / 2;
    
    cropOverlay.style.left = left + 'px';
    cropOverlay.style.top = top + 'px';
    cropOverlay.style.width = size + 'px';
    cropOverlay.style.height = size + 'px';
}

// Сброс кадрирования
function resetCrop() {
    const cropOverlay = document.getElementById('crop-overlay');
    cropOverlay.style.display = 'none';
    document.querySelector('[data-action="crop"]').classList.remove('active');
}

// Масштабирование
function zoomIn() {
    if (zoomLevel < 3) {
        zoomLevel += 0.2;
        updateZoomDisplay();
        drawImage();
    }
}

function zoomOut() {
    if (zoomLevel > 0.3) {
        zoomLevel -= 0.2;
        updateZoomDisplay();
        drawImage();
    }
}

function updateZoomDisplay() {
    const zoomLevelElement = document.querySelector('.zoom-level');
    zoomLevelElement.textContent = Math.round(zoomLevel * 100) + '%';
}

// Поворот
function rotate(angle) {
    rotation += angle;
    if (rotation >= 360) rotation -= 360;
    if (rotation < 0) rotation += 360;
    drawImage();
}

// Замена изображения
function replaceImage() {
    // Сбрасываем масштаб и поворот
    zoomLevel = 1;
    rotation = 0;
    
    // Обновляем отображение масштаба
    updateZoomDisplay();
    
    // Скрываем область кадрирования, если она активна
    const cropOverlay = document.getElementById('crop-overlay');
    if (cropOverlay && cropOverlay.style.display === 'block') {
        resetCrop();
    }
    
    // Открываем диалог выбора файла
    document.getElementById('image-upload').click();
}

// Выбор стиля
function selectStyle(style) {
    selectedStyle = style;
    
    // Обновляем визуальное выделение
    document.querySelectorAll('.style-option').forEach(option => {
        option.classList.remove('selected');
    });
    
    document.querySelector(`[data-style="${style}"]`).classList.add('selected');
    
    updatePreview();
}

// Обновление предварительного просмотра
function updatePreview() {
    if (!currentImage || !previewCanvas) return;
    
    const canvasWidth = previewCanvas.width;
    const canvasHeight = previewCanvas.height;
    
    // Очищаем canvas
    previewCtx.clearRect(0, 0, canvasWidth, canvasHeight);
    
    // Устанавливаем качество рендеринга
    previewCtx.imageSmoothingEnabled = true;
    previewCtx.imageSmoothingQuality = 'high';
    
    // Применяем выбранный стиль
    applyStyle(previewCtx, canvasWidth, canvasHeight);
    
    // Рисуем изображение с сохранением пропорций
    const imgAspect = currentImage.width / currentImage.height;
    const canvasAspect = canvasWidth / canvasHeight;
    
    let drawWidth, drawHeight, offsetX, offsetY;
    
    if (imgAspect > canvasAspect) {
        // Изображение шире, чем canvas
        drawWidth = canvasWidth;
        drawHeight = canvasWidth / imgAspect;
        offsetX = 0;
        offsetY = (canvasHeight - drawHeight) / 2;
    } else {
        // Изображение выше, чем canvas
        drawHeight = canvasHeight;
        drawWidth = canvasHeight * imgAspect;
        offsetX = (canvasWidth - drawWidth) / 2;
        offsetY = 0;
    }
    
    previewCtx.drawImage(currentImage, offsetX, offsetY, drawWidth, drawHeight);
}

// Применение стиля
function applyStyle(context, width, height) {
    const lighting = document.getElementById('lighting-setting').value;
    const contrast = document.getElementById('contrast-setting').value;
    
    switch (selectedStyle) {
        case 'enhanced':
            // AI улучшение
            context.filter = 'contrast(1.2) brightness(1.1) saturate(1.1)';
            break;
        case 'pop-art':
            // Поп-арт
            context.filter = 'contrast(1.5) saturate(1.8) brightness(1.2)';
            break;
        case 'vintage':
            // Винтаж
            context.filter = 'sepia(0.8) contrast(1.1) brightness(0.9)';
            break;
        default:
            // Оригинал
            context.filter = 'none';
    }
    
    // Применяем настройки освещения
    switch (lighting) {
        case 'sun':
            context.filter += ' brightness(1.2)';
            break;
        case 'moon':
            context.filter += ' brightness(0.8) saturate(0.7)';
            break;
        case 'venus':
            context.filter += ' brightness(1.1) saturate(1.3)';
            break;
    }
    
    // Применяем настройки контрастности
    if (contrast === 'high') {
        context.filter += ' contrast(1.3)';
    }
}

// Навигация по этапам
function nextStep() {
    if (currentStep < totalSteps) {
        currentStep++;
        updateStepDisplay();
    }
}

function previousStep() {
    if (currentStep > 1) {
        currentStep--;
        updateStepDisplay();
    }
}

function updateStepDisplay() {
    // Обновляем активный этап
    document.querySelectorAll('.step').forEach((step, index) => {
        step.classList.remove('active', 'completed');
        if (index + 1 < currentStep) {
            step.classList.add('completed');
        } else if (index + 1 === currentStep) {
            step.classList.add('active');
        }
    });
    
    // Показываем соответствующий контент
    document.querySelectorAll('.editor-step').forEach((step, index) => {
        step.classList.remove('active');
        if (index + 1 === currentStep) {
            step.classList.add('active');
        }
    });
    
    // Обновляем кнопки навигации
    const prevBtn = document.getElementById('prev-step');
    const nextBtn = document.getElementById('next-step');
    
    // Скрываем кнопку "Назад" на первом этапе
    if (currentStep === 1) {
        prevBtn.style.display = 'none';
    } else {
        prevBtn.style.display = 'inline-flex';
    }
    
    if (currentStep === totalSteps) {
        nextBtn.textContent = 'Завершить';
        // Удаляем старый обработчик и добавляем новый
        nextBtn.removeEventListener('click', nextStep);
        nextBtn.addEventListener('click', createScheme);
    } else {
        nextBtn.textContent = 'Далее';
        // Удаляем старый обработчик и добавляем новый
        nextBtn.removeEventListener('click', createScheme);
        nextBtn.addEventListener('click', nextStep);
    }
    
    // Обновляем предварительный просмотр на этапе стилей
    if (currentStep === 3) {
        updatePreview();
    }
    
    // Обновляем финальный просмотр на этапе подтверждения
    if (currentStep === 4) {
        updateFinalPreview();
    }
}

// Обновление финального просмотра
function updateFinalPreview() {
    if (!currentImage || !finalCanvas) return;
    
    const canvasWidth = finalCanvas.width;
    const canvasHeight = finalCanvas.height;
    
    // Очищаем canvas
    finalCtx.clearRect(0, 0, canvasWidth, canvasHeight);
    
    // Устанавливаем качество рендеринга
    finalCtx.imageSmoothingEnabled = true;
    finalCtx.imageSmoothingQuality = 'high';
    
    // Применяем стиль
    applyStyle(finalCtx, canvasWidth, canvasHeight);
    
    // Рисуем изображение с сохранением пропорций
    const imgAspect = currentImage.width / currentImage.height;
    const canvasAspect = canvasWidth / canvasHeight;
    
    let drawWidth, drawHeight, offsetX, offsetY;
    
    if (imgAspect > canvasAspect) {
        // Изображение шире, чем canvas
        drawWidth = canvasWidth;
        drawHeight = canvasWidth / imgAspect;
        offsetX = 0;
        offsetY = (canvasHeight - drawHeight) / 2;
    } else {
        // Изображение выше, чем canvas
        drawHeight = canvasHeight;
        drawWidth = canvasHeight * imgAspect;
        offsetX = (canvasWidth - drawWidth) / 2;
        offsetY = 0;
    }
    
    finalCtx.drawImage(currentImage, offsetX, offsetY, drawWidth, drawHeight);
}

// Возврат к выбору стилей
function goBackToStyles() {
    currentStep = 3;
    updateStepDisplay();
}

// Создание схемы
function createScheme() {
    showSchemeModal();
    
    // Имитация процесса создания схемы
    let progress = 0;
    const progressFill = document.getElementById('progress-fill');
    const progressText = document.getElementById('progress-text');
    
    const interval = setInterval(() => {
        progress += Math.random() * 15;
        if (progress > 100) progress = 100;
        
        progressFill.style.width = progress + '%';
        
        if (progress < 30) {
            progressText.textContent = 'Анализируем изображение...';
        } else if (progress < 60) {
            progressText.textContent = 'Применяем стиль...';
        } else if (progress < 90) {
            progressText.textContent = 'Создаем схему мозаики...';
        } else {
            progressText.textContent = 'Завершаем обработку...';
        }
        
        if (progress >= 100) {
            clearInterval(interval);
            setTimeout(() => {
                hideSchemeModal();
                showSchemeComplete();
            }, 1000);
        }
    }, 200);
}

// Показать модальное окно создания схемы
function showSchemeModal() {
    document.getElementById('scheme-modal').classList.add('show');
}

// Скрыть модальное окно создания схемы
function hideSchemeModal() {
    document.getElementById('scheme-modal').classList.remove('show');
}

// Показать завершение создания схемы
function showSchemeComplete() {
    showNotification('Схема успешно создана!', 'success');
    
    // Здесь будет переход на страницу с результатом
    setTimeout(() => {
        // window.location.href = 'result.html?coupon=' + couponData.number;
        showNotification('Функция скачивания схемы будет добавлена позже', 'info');
    }, 2000);
}

// Показать уведомление
function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.innerHTML = `
        <div class="notification-content">
            <i class="fas fa-${type === 'error' ? 'exclamation-circle' : type === 'success' ? 'check-circle' : 'info-circle'}"></i>
            <span>${message}</span>
        </div>
    `;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.classList.add('show');
    }, 100);
    
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    }, 5000);
}
