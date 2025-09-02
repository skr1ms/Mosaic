import axios from 'axios'

// API base URL - зависит от окружения
// В development используем переменную из Docker Compose, в production - /api
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Логируем API URL для отладки
console.log('Public Site API Base URL:', API_BASE_URL)
console.log('Environment:', import.meta.env.VITE_ENVIRONMENT)
console.log('Debug mode:', import.meta.env.VITE_DEBUG)

// Interceptors для обработки ошибок
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error?.response?.status
    const data = error?.response?.data || {}
    const problemType = data?.type
    const message = data?.detail || data?.title || data?.message || error.message || 'Request failed'
    const shaped = new Error(message)
    shaped.name = error?.name || 'AxiosError'
    shaped.status = status
    shaped.problemType = problemType
    shaped.original = error
    console.error('API Error:', { status, problemType, message })
    return Promise.reject(shaped)
  }
)

export class MosaicAPI {
  // Support chat
  static async startSupportChat(title = '') {
    const { data } = await apiClient.post('/public/support/start' + (title ? `?title=${encodeURIComponent(title)}` : ''))
    return data
  }

  static async getSupportMessages(chatId, token) {
    const { data } = await apiClient.get(`/public/support/messages`, {
      params: { chat_id: chatId },
      headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    })
    return data?.messages || data?.data?.messages || data?.data || []
  }

  static async sendSupportMessage(chatId, content, token) {
    const { data } = await apiClient.post(`/public/support/messages`, { chat_id: chatId, content }, {
      headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    })
    return data?.message || data?.data || data
  }

  static async updateSupportMessage(id, content, token) {
    const { data } = await apiClient.patch(`/public/support/messages/${id}`, { content }, {
      headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    })
    return data?.message || data?.data || data
  }

  static async deleteSupportMessage(id, token) {
    const { data } = await apiClient.delete(`/public/support/messages/${id}`, {
      headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    })
    return data?.result || data?.data || data
  }
  // Проверка купона
  static async validateCoupon(code) {
    const { data } = await apiClient.get(`/coupons/${code}`)
    return data
  }

  // Активация купона
  static async activateCoupon(code) {
    const { data } = await apiClient.post(`/coupons/${code}/activate`)
    return data
  }

  // Получение брендинга (публичный эндпоинт)
  static async getBrandingInfo() {
    const { data } = await apiClient.get(`/branding`)
    return data
  }
  
  // Загрузка изображения
  static async uploadImage(formData) {
    const { data } = await apiClient.post('/images/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return data
  }

  // Генерация превью мозаики
  static async generatePreview(formData) {
    const { data } = await apiClient.post('/preview/generate', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      timeout: 120000, // 2 минуты для генерации превью
    })
    return data
  }

  // Генерация вариантов превью с контрастами
  static async generatePreviewVariant(formData) {
    const { data } = await apiClient.post('/preview/generate-variant', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      timeout: 120000,
    })
    return data
  }

  // Генерация AI превью
  static async generateAIPreview(formData) {
    const { data } = await apiClient.post('/preview/generate-ai', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      timeout: 180000, // 3 минуты для AI обработки
    })
    return data
  }

  // Редактирование изображения
  static async editImage(imageId, params) {
    const { data } = await apiClient.post(`/images/${imageId}/edit`, params)
    return data
  }

  // Обработка изображения через AI
  static async processImage(imageId, params) {
    const { data } = await apiClient.post(`/images/${imageId}/process`, params)
    return data
  }

  // Генерация схемы мозаики
  static async generateSchema(imageId, params) {
    const { data } = await apiClient.post(`/images/${imageId}/generate-schema`, params)
    return data
  }

  // Получение статуса обработки
  static async getProcessingStatus(imageId) {
    const { data } = await apiClient.get(`/images/${imageId}/status`)
    return data
  }

  // Скачивание ZIP архива схемы
  static async downloadSchemaArchive(schemaUUID) {
    const url = `${API_BASE_URL}/public/schemas/${schemaUUID}/download`
    window.open(url, '_blank')
  }

  // Отправка схемы на email
  static async sendSchemaToEmail(imageId, payload) {
    // Бэкенд обрабатывает по /api/images/:id/send-email
    const body = typeof payload === 'string' ? { email: payload } : payload
    const { data } = await apiClient.post(`/images/${imageId}/send-email`, body)
    return data
  }

  // Покупка купона — инициируем заказ и получаем ссылку эквайринга
  static async initiateCouponOrder(params) {
    const { data } = await apiClient.post('/payment/purchase', params)
    return data
  }

  // Проверка статуса заказа
  static async getOrderStatus(orderNumber) {
    const { data } = await apiClient.get(`/payment/orders/${orderNumber}/status`)
    return data
  }

  // Получение доступных размеров
  static async getAvailableSizes() {
    const { data } = await apiClient.get('/sizes')
    return data
  }

  // Получение доступных стилей
  static async getAvailableStyles() {
    const { data } = await apiClient.get('/styles')
    return data
  }

  // Получение сетки артикулов партнера
  static async getPartnerArticleGrid(partnerId) {
    const { data } = await apiClient.get(`/admin/partners/${partnerId}/articles/grid`)
    return data
  }

  // Генерация URL товара в маркетплейсе (публичный метод)
  static async generateProductURL(partnerId, params) {
    const { data } = await apiClient.post(`/partners/${partnerId}/articles/generate-url`, params)
    return data
  }
}

export default apiClient
