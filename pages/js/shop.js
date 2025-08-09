// Глобальные переменные магазина
let selectedSize = null;
let selectedStyle = null;
let cartTotal = 0;

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    initializeShop();
    setupShopEventListeners();
});

// Инициализация магазина
function initializeShop() {
    // Загружаем данные партнера для брендинга
    loadPartnerData();
    
    // Инициализируем корзину
    updateCart();
}

// Настройка обработчиков событий
function setupShopEventListeners() {
    // Выбор размера
    document.querySelectorAll('.size-option').forEach(option => {
        option.addEventListener('click', function() {
            selectSize(this.dataset.size, parseInt(this.dataset.price));
        });
    });
    
    // Выбор стиля
    document.querySelectorAll('.style-option').forEach(option => {
        option.addEventListener('click', function() {
            selectStyle(this.dataset.style, parseInt(this.dataset.price));
        });
    });
    
    // Кнопка покупки
    document.getElementById('buy-button').addEventListener('click', showPaymentModal);
    
    // Форма оплаты
    document.getElementById('payment-form').addEventListener('submit', handlePayment);
    
    // Валидация полей формы
    setupFormValidation();
}

// Загрузка данных партнера
function loadPartnerData() {
    // Здесь будет API запрос для получения данных партнера
    // Пока используем заглушку
    const partnerData = {
        name: 'Алмазная мозаика',
        logo: 'assets/logo.svg',
        email: 'info@diamond-art.ru',
        phone: '+7 (999) 123-45-67',
        telegram: 'https://t.me/diamond_art',
        whatsapp: 'https://wa.me/79991234567',
        instagram: 'https://instagram.com/diamond_art',
        ozonLink: 'https://ozon.ru/diamond-art',
        wildberriesLink: 'https://wildberries.ru/diamond-art',
        yandexMarketLink: 'https://market.yandex.ru/diamond-art'
    };
    
    applyPartnerBranding(partnerData);
}

// Применение брендинга партнера
function applyPartnerBranding(partnerData) {
    if (!partnerData) return;
    
    // Логотип и название
    const logoImg = document.getElementById('partner-logo');
    const partnerName = document.getElementById('partner-name');
    
    if (logoImg && partnerData.logo) {
        logoImg.src = partnerData.logo;
    }
    
    if (partnerName) {
        partnerName.textContent = partnerData.name;
    }
    
    // Контакты
    const partnerEmail = document.getElementById('partner-email');
    const partnerPhone = document.getElementById('partner-phone');
    
    if (partnerEmail) {
        partnerEmail.textContent = partnerData.email;
    }
    
    if (partnerPhone) {
        partnerPhone.textContent = partnerData.phone;
    }
    
    // Социальные сети
    const telegramLink = document.getElementById('telegram-link');
    const whatsappLink = document.getElementById('whatsapp-link');
    const instagramLink = document.getElementById('instagram-link');
    
    if (telegramLink && partnerData.telegram) {
        telegramLink.href = partnerData.telegram;
    }
    
    if (whatsappLink && partnerData.whatsapp) {
        whatsappLink.href = partnerData.whatsapp;
    }
    
    if (instagramLink && partnerData.instagram) {
        instagramLink.href = partnerData.instagram;
    }
    
    // Маркетплейсы
    const ozonShopLink = document.getElementById('ozon-shop-link');
    const wildberriesShopLink = document.getElementById('wildberries-shop-link');
    const yandexMarketShopLink = document.getElementById('yandex-market-shop-link');
    
    if (ozonShopLink && partnerData.ozonLink) {
        ozonShopLink.href = partnerData.ozonLink;
    }
    
    if (wildberriesShopLink && partnerData.wildberriesLink) {
        wildberriesShopLink.href = partnerData.wildberriesLink;
    }
    
    if (yandexMarketShopLink && partnerData.yandexMarketLink) {
        yandexMarketShopLink.href = partnerData.yandexMarketLink;
    }
}

// Выбор размера
function selectSize(size, price) {
    selectedSize = { size, price };
    
    // Обновляем визуальное выделение
    document.querySelectorAll('.size-option').forEach(option => {
        option.classList.remove('selected');
    });
    
    document.querySelector(`[data-size="${size}"]`).classList.add('selected');
    
    updateCart();
}

// Выбор стиля
function selectStyle(style, price) {
    selectedStyle = { style, price };
    
    // Обновляем визуальное выделение
    document.querySelectorAll('.style-option').forEach(option => {
        option.classList.remove('selected');
    });
    
    document.querySelector(`[data-style="${style}"]`).classList.add('selected');
    
    updateCart();
}

// Обновление корзины
function updateCart() {
    const sizeItem = document.getElementById('size-item');
    const styleItem = document.getElementById('style-item');
    const totalPrice = document.getElementById('total-price');
    const buyButton = document.getElementById('buy-button');
    
    // Обновляем размер
    if (selectedSize) {
        sizeItem.querySelector('.item-value').textContent = selectedSize.size;
        sizeItem.querySelector('.item-price').textContent = formatPrice(selectedSize.price);
    } else {
        sizeItem.querySelector('.item-value').textContent = 'Не выбран';
        sizeItem.querySelector('.item-price').textContent = '0 ₽';
    }
    
    // Обновляем стиль
    if (selectedStyle) {
        const styleNames = {
            'grayscale': 'Оттенки серого',
            'flesh': 'Оттенки телесного',
            'pop-art': 'Поп-арт',
            'max-colors': 'Максимум цветов'
        };
        
        styleItem.querySelector('.item-value').textContent = styleNames[selectedStyle.style] || selectedStyle.style;
        
        if (selectedStyle.price === 0) {
            styleItem.querySelector('.item-price').textContent = 'Бесплатно';
        } else {
            styleItem.querySelector('.item-price').textContent = formatPrice(selectedStyle.price);
        }
    } else {
        styleItem.querySelector('.item-value').textContent = 'Не выбран';
        styleItem.querySelector('.item-price').textContent = '0 ₽';
    }
    
    // Вычисляем общую стоимость
    const sizePrice = selectedSize ? selectedSize.price : 0;
    const stylePrice = selectedStyle ? selectedStyle.price : 0;
    cartTotal = sizePrice + stylePrice;
    
    totalPrice.textContent = formatPrice(cartTotal);
    
    // Активируем кнопку покупки
    buyButton.disabled = !selectedSize || !selectedStyle;
}

// Форматирование цены
function formatPrice(price) {
    return new Intl.NumberFormat('ru-RU').format(price) + ' ₽';
}

// Показать модальное окно оплаты
function showPaymentModal() {
    if (!selectedSize || !selectedStyle) {
        showNotification('Пожалуйста, выберите размер и стиль', 'error');
        return;
    }
    
    // Обновляем сводку заказа
    document.getElementById('summary-size').textContent = selectedSize.size;
    
    const styleNames = {
        'grayscale': 'Оттенки серого',
        'flesh': 'Оттенки телесного',
        'pop-art': 'Поп-арт',
        'max-colors': 'Максимум цветов'
    };
    
    document.getElementById('summary-style').textContent = styleNames[selectedStyle.style] || selectedStyle.style;
    document.getElementById('summary-total').textContent = formatPrice(cartTotal);
    
    // Показываем модальное окно
    document.getElementById('payment-modal').classList.add('show');
}

// Закрыть модальное окно оплаты
function closePaymentModal() {
    document.getElementById('payment-modal').classList.remove('show');
    document.getElementById('payment-form').reset();
}

// Настройка валидации формы
function setupFormValidation() {
    // Форматирование номера карты
    const cardNumber = document.getElementById('card-number');
    cardNumber.addEventListener('input', function() {
        let value = this.value.replace(/\D/g, '');
        value = value.replace(/(\d{4})(?=\d)/g, '$1 ');
        this.value = value.substring(0, 19);
    });
    
    // Форматирование срока действия
    const cardExpiry = document.getElementById('card-expiry');
    cardExpiry.addEventListener('input', function() {
        let value = this.value.replace(/\D/g, '');
        if (value.length >= 2) {
            value = value.substring(0, 2) + '/' + value.substring(2, 4);
        }
        this.value = value.substring(0, 5);
    });
    
    // Форматирование CVV
    const cardCvv = document.getElementById('card-cvv');
    cardCvv.addEventListener('input', function() {
        this.value = this.value.replace(/\D/g, '').substring(0, 3);
    });
}

// Обработка оплаты
function handlePayment(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const email = document.getElementById('email').value;
    const cardNumber = document.getElementById('card-number').value;
    const cardExpiry = document.getElementById('card-expiry').value;
    const cardCvv = document.getElementById('card-cvv').value;
    
    // Валидация
    if (!email || !cardNumber || !cardExpiry || !cardCvv) {
        showNotification('Пожалуйста, заполните все поля', 'error');
        return;
    }
    
    if (!isValidEmail(email)) {
        showNotification('Пожалуйста, введите корректный email', 'error');
        return;
    }
    
    if (cardNumber.replace(/\s/g, '').length !== 16) {
        showNotification('Номер карты должен содержать 16 цифр', 'error');
        return;
    }
    
    if (cardExpiry.length !== 5) {
        showNotification('Введите корректный срок действия карты', 'error');
        return;
    }
    
    if (cardCvv.length !== 3) {
        showNotification('CVV должен содержать 3 цифры', 'error');
        return;
    }
    
    // Показываем индикатор загрузки
    showPaymentLoading();
    
    // Имитация процесса оплаты
    setTimeout(() => {
        hidePaymentLoading();
        processPaymentSuccess();
    }, 3000);
}

// Показать загрузку оплаты
function showPaymentLoading() {
    const submitButton = document.querySelector('#payment-form button[type="submit"]');
    submitButton.disabled = true;
    submitButton.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Обрабатываем оплату...';
}

// Скрыть загрузку оплаты
function hidePaymentLoading() {
    const submitButton = document.querySelector('#payment-form button[type="submit"]');
    submitButton.disabled = false;
    submitButton.innerHTML = '<i class="fas fa-lock"></i> Оплатить';
}

// Обработка успешной оплаты
function processPaymentSuccess() {
    // Генерируем номер купона
    const couponNumber = generateCouponNumber();
    
    // Закрываем модальное окно
    closePaymentModal();
    
    // Показываем уведомление об успехе
    showNotification('Оплата прошла успешно! Купон отправлен на email', 'success');
    
    // Здесь будет отправка купона на email
    console.log('Купон отправлен:', {
        number: couponNumber,
        size: selectedSize.size,
        style: selectedStyle.style,
        email: document.getElementById('email').value
    });
    
    // Сброс выбора
    resetSelection();
}

// Генерация номера купона
function generateCouponNumber() {
    return Math.random().toString().substring(2, 14);
}

// Сброс выбора
function resetSelection() {
    selectedSize = null;
    selectedStyle = null;
    
    document.querySelectorAll('.size-option, .style-option').forEach(option => {
        option.classList.remove('selected');
    });
    
    updateCart();
}

// Валидация email
function isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
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

// Закрытие модального окна при клике вне его
document.addEventListener('click', function(event) {
    const modal = document.getElementById('payment-modal');
    if (event.target === modal) {
        closePaymentModal();
    }
});

// Закрытие модального окна по Escape
document.addEventListener('keydown', function(event) {
    if (event.key === 'Escape') {
        closePaymentModal();
    }
});
