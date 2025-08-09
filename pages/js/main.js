// Глобальные переменные
let currentLanguage = 'ru';
let partnerData = null;

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
    setupEventListeners();
    loadPartnerData();
});

// Инициализация приложения
function initializeApp() {
    // Загружаем сохраненный язык
    const savedLanguage = localStorage.getItem('language') || 'ru';
    
    // Инициализируем выпадающий список языков
    const currentLangSpan = document.getElementById('current-lang');
    if (currentLangSpan) {
        currentLangSpan.textContent = savedLanguage.toUpperCase();
    }
    
    // Устанавливаем выбранную опцию
    document.querySelectorAll('.lang-option').forEach(option => {
        option.classList.remove('selected');
        if (option.dataset.lang === savedLanguage) {
            option.classList.add('selected');
        }
    });
    
    setLanguage(savedLanguage);
    
    // Инициализируем FAQ
    initializeFAQ();
    
    // Инициализируем валидацию купонов
    initializeCouponValidation();
}

// Настройка обработчиков событий
function setupEventListeners() {
    // Выпадающий список языков
    const langDropdownBtn = document.getElementById('lang-dropdown-btn');
    const langDropdownMenu = document.getElementById('lang-dropdown-menu');
    const langOptions = document.querySelectorAll('.lang-option');
    
    if (langDropdownBtn && langDropdownMenu) {
        // Открытие/закрытие выпадающего списка
        langDropdownBtn.addEventListener('click', function(e) {
            e.stopPropagation();
            const dropdown = this.closest('.language-dropdown');
            dropdown.classList.toggle('active');
        });
        
        // Выбор языка
        langOptions.forEach(option => {
            option.addEventListener('click', function() {
                const lang = this.dataset.lang;
                setLanguage(lang);
                
                // Закрываем выпадающий список
                const dropdown = this.closest('.language-dropdown');
                dropdown.classList.remove('active');
            });
        });
        
        // Закрытие при клике вне выпадающего списка
        document.addEventListener('click', function(e) {
            if (!e.target.closest('.language-dropdown')) {
                const dropdown = document.querySelector('.language-dropdown');
                if (dropdown) {
                    dropdown.classList.remove('active');
                }
            }
        });
    }
    
    // Навигационные ссылки
    document.querySelectorAll('.nav-link').forEach(link => {
        link.addEventListener('click', function(e) {
            const href = this.getAttribute('href');
            
            // Если ссылка ведет на другую страницу (не якорь), не предотвращаем стандартное поведение
            if (href.startsWith('http') || href.includes('.html') || href === 'index.html') {
                return; // Позволяем браузеру обработать переход
            }
            
            // Если это ссылка "Главная" на главной странице, прокручиваем вверх
            if (href === '#' && window.location.pathname.includes('index.html')) {
                e.preventDefault();
                window.scrollTo({ top: 0, behavior: 'smooth' });
                return;
            }
            
            // Для якорных ссылок используем плавную прокрутку
            e.preventDefault();
            const target = href.substring(1);
            navigateToSection(target);
        });
    });
    
    // Ввод купона
    const couponInput = document.getElementById('coupon-code');
    if (couponInput) {
        couponInput.addEventListener('input', function() {
            this.value = this.value.replace(/[^0-9]/g, '').substring(0, 12);
        });
        
        couponInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                activateCoupon();
            }
        });
    }
}

// Загрузка данных партнера
function loadPartnerData() {
    // Получаем домен для определения партнера
    const hostname = window.location.hostname;
    
    // Здесь будет API запрос для получения данных партнера
    // Пока используем заглушку
    partnerData = {
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
    
    applyPartnerBranding();
}

// Применение брендинга партнера
function applyPartnerBranding() {
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
    const ozonLink = document.getElementById('ozon-link');
    const wildberriesLink = document.getElementById('wildberries-link');
    const yandexMarketLink = document.getElementById('yandex-market-link');
    
    if (ozonLink && partnerData.ozonLink) {
        ozonLink.href = partnerData.ozonLink;
    }
    
    if (wildberriesLink && partnerData.wildberriesLink) {
        wildberriesLink.href = partnerData.wildberriesLink;
    }
    
    if (yandexMarketLink && partnerData.yandexMarketLink) {
        yandexMarketLink.href = partnerData.yandexMarketLink;
    }
}

// Переключение языков
function setLanguage(lang) {
    currentLanguage = lang;
    localStorage.setItem('language', lang);
    
    // Обновляем текст в кнопке выпадающего списка
    const currentLangSpan = document.getElementById('current-lang');
    if (currentLangSpan) {
        currentLangSpan.textContent = lang.toUpperCase();
    }
    
    // Обновляем выбранную опцию в выпадающем списке
    document.querySelectorAll('.lang-option').forEach(option => {
        option.classList.remove('selected');
        if (option.dataset.lang === lang) {
            option.classList.add('selected');
        }
    });
    
    // Здесь будет загрузка переводов
    loadTranslations(lang);
}

// Загрузка переводов
function loadTranslations(lang) {
    const translations = {
        ru: {
            'hero-title': 'Создайте свою уникальную картину',
            'hero-subtitle': 'Алмазная мозаика - увлекательное хобби для всей семьи',
            'coupon-placeholder': 'Введите номер купона (12 цифр)',
            'activate-btn': 'Активировать',
            'buy-coupon-title': 'Хочу купить купон',
            'buy-coupon-desc': 'У меня есть набор камней алмазная мозаика, хочу купить купон',
            'go-to-shop': 'Перейти в магазин',
            'diamond-art': 'Алмазная мозаика',
            'paint-by-numbers': 'Картины по номерам',
            'create-picture': 'Создать картину',
            'coming-soon': 'Скоро в продаже',
            'faq-title': 'Часто задаваемые вопросы',
            'where-to-buy': 'Где купить',
            'contacts': 'Контакты',
            'social-networks': 'Социальные сети'
        },
        en: {
            'hero-title': 'Create Your Unique Painting',
            'hero-subtitle': 'Diamond Art - Exciting Hobby for the Whole Family',
            'coupon-placeholder': 'Enter coupon number (12 digits)',
            'activate-btn': 'Activate',
            'buy-coupon-title': 'I Want to Buy a Coupon',
            'buy-coupon-desc': 'I have a diamond art kit, I want to buy a coupon',
            'go-to-shop': 'Go to Shop',
            'diamond-art': 'Diamond Art',
            'paint-by-numbers': 'Paint by Numbers',
            'create-picture': 'Create Picture',
            'coming-soon': 'Coming Soon',
            'faq-title': 'Frequently Asked Questions',
            'where-to-buy': 'Where to Buy',
            'contacts': 'Contacts',
            'social-networks': 'Social Networks'
        },
        es: {
            'hero-title': 'Crea Tu Pintura Única',
            'hero-subtitle': 'Arte de Diamantes - Hobby Emocionante para Toda la Familia',
            'coupon-placeholder': 'Ingrese número de cupón (12 dígitos)',
            'activate-btn': 'Activar',
            'buy-coupon-title': 'Quiero Comprar un Cupón',
            'buy-coupon-desc': 'Tengo un kit de arte de diamantes, quiero comprar un cupón',
            'go-to-shop': 'Ir a la Tienda',
            'diamond-art': 'Arte de Diamantes',
            'paint-by-numbers': 'Pintar por Números',
            'create-picture': 'Crear Pintura',
            'coming-soon': 'Próximamente',
            'faq-title': 'Preguntas Frecuentes',
            'where-to-buy': 'Dónde Comprar',
            'contacts': 'Contactos',
            'social-networks': 'Redes Sociales'
        }
    };
    
    const currentTranslations = translations[lang] || translations.ru;
    
    // Применяем переводы
    Object.keys(currentTranslations).forEach(key => {
        const elements = document.querySelectorAll(`[data-translate="${key}"]`);
        elements.forEach(element => {
            element.textContent = currentTranslations[key];
        });
    });
}

// Инициализация FAQ
function initializeFAQ() {
    document.querySelectorAll('.faq-question').forEach(question => {
        question.addEventListener('click', function() {
            const faqItem = this.parentElement;
            const isActive = faqItem.classList.contains('active');
            
            // Закрываем все FAQ
            document.querySelectorAll('.faq-item').forEach(item => {
                item.classList.remove('active');
            });
            
            // Открываем текущий, если он был закрыт
            if (!isActive) {
                faqItem.classList.add('active');
            }
        });
    });
}

// Инициализация валидации купонов
function initializeCouponValidation() {
    const couponInput = document.getElementById('coupon-code');
    if (couponInput) {
        couponInput.addEventListener('blur', function() {
            validateCoupon(this.value);
        });
    }
}

// Валидация купона
function validateCoupon(code) {
    if (!code) return true;
    
    const isValid = /^\d{12}$/.test(code);
    const input = document.getElementById('coupon-code');
    
    if (isValid) {
        input.style.borderColor = '#10b981';
        return true;
    } else {
        input.style.borderColor = '#ef4444';
        return false;
    }
}

// Активация купона
function activateCoupon() {
    const couponCode = document.getElementById('coupon-code').value;
    
    if (!validateCoupon(couponCode)) {
        showNotification('Пожалуйста, введите корректный 12-значный номер купона', 'error');
        return;
    }
    
    if (!couponCode) {
        showNotification('Пожалуйста, введите номер купона', 'error');
        return;
    }
    
    // Показываем индикатор загрузки
    showLoading();
    
    // Здесь будет API запрос для активации купона
    setTimeout(() => {
        hideLoading();
        // Переход на страницу редактора
        window.location.href = `editor.html?coupon=${couponCode}`;
    }, 2000);
}

// Переход в магазин
function goToShop() {
    window.location.href = 'shop.html';
}

// Переход в редактор
function goToEditor() {
    window.location.href = 'editor.html';
}

// Навигация по разделам
function navigateToSection(sectionId) {
    const section = document.getElementById(sectionId);
    if (section) {
        section.scrollIntoView({ behavior: 'smooth' });
    } else if (sectionId === '') {
        // Если sectionId пустой, прокручиваем вверх
        window.scrollTo({ top: 0, behavior: 'smooth' });
    }
    
    // Обновляем активную ссылку
    document.querySelectorAll('.nav-link').forEach(link => {
        link.classList.remove('active');
    });
    
    const activeLink = document.querySelector(`[href="#${sectionId}"]`);
    if (activeLink) {
        activeLink.classList.add('active');
    }
}

// Показать уведомление
function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.innerHTML = `
        <div class="notification-content">
            <i class="fas fa-${type === 'error' ? 'exclamation-circle' : 'info-circle'}"></i>
            <span>${message}</span>
        </div>
    `;
    
    document.body.appendChild(notification);
    
    // Анимация появления
    setTimeout(() => {
        notification.classList.add('show');
    }, 100);
    
    // Автоматическое скрытие
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    }, 5000);
}

// Показать загрузку
function showLoading() {
    const loading = document.createElement('div');
    loading.className = 'loading-overlay';
    loading.innerHTML = `
        <div class="loading-spinner">
            <i class="fas fa-spinner fa-spin"></i>
            <p>Активируем купон...</p>
        </div>
    `;
    
    document.body.appendChild(loading);
}

// Скрыть загрузку
function hideLoading() {
    const loading = document.querySelector('.loading-overlay');
    if (loading) {
        document.body.removeChild(loading);
    }
}

// Добавляем стили для уведомлений и загрузки
const additionalStyles = `
<style>
.notification {
    position: fixed;
    top: 20px;
    right: 20px;
    background: white;
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0,0,0,0.15);
    padding: 1rem;
    z-index: 10000;
    transform: translateX(100%);
    transition: transform 0.3s ease;
    max-width: 300px;
}

.notification.show {
    transform: translateX(0);
}

.notification-content {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.notification-error {
    border-left: 4px solid #ef4444;
}

.notification-info {
    border-left: 4px solid #3b82f6;
}

.loading-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0,0,0,0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10000;
}

.loading-spinner {
    background: white;
    padding: 2rem;
    border-radius: 12px;
    text-align: center;
    box-shadow: 0 10px 30px rgba(0,0,0,0.2);
}

.loading-spinner i {
    font-size: 2rem;
    color: #3b82f6;
    margin-bottom: 1rem;
}

.loading-spinner p {
    color: #64748b;
    margin: 0;
}
</style>
`;

document.head.insertAdjacentHTML('beforeend', additionalStyles);
