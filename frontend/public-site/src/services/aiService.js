import { MosaicAPI } from '../api/client'

class AIService {
  constructor() {
    // В development используем переменную из Docker Compose, в production - localhost
    this.baseUrl = process.env.REACT_APP_AI_SERVICE_URL || 'http://localhost:8080'
    this.apiKey = process.env.REACT_APP_AI_API_KEY
  }

  /**
   * Обработка изображения через Stable Diffusion
   * @param {string} imageUrl - URL изображения для обработки
   * @param {Object} options - Параметры обработки
   * @returns {Promise<Object>} - Результат обработки
   */
  async processImageWithStableDiffusion(imageUrl, options = {}) {
    try {
      const {
        style = 'enhanced',
        lighting = 'natural',
        contrast = 'normal',
        strength = 0.75
      } = options

      const response = await fetch(`${this.baseUrl}/api/ai/process-image`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.apiKey}`
        },
        body: JSON.stringify({
          image_url: imageUrl,
          style,
          lighting,
          contrast,
          strength,
          model: 'stable-diffusion-v1-5'
        })
      })

      if (!response.ok) {
        throw new Error(`AI processing failed: ${response.statusText}`)
      }

      const result = await response.json()
      return result
    } catch (error) {
      console.error('AI processing error:', error)
      throw new Error('Failed to process image with AI')
    }
  }

  /**
   * Генерация схемы мозаики
   * @param {string} imageUrl - URL обработанного изображения
   * @param {Object} options - Параметры генерации
   * @returns {Promise<Object>} - Результат генерации
   */
  async generateMosaicSchema(imageUrl, options = {}) {
    try {
      const {
        size = '40x50',
        style = 'enhanced',
        quality = 'high'
      } = options

      const response = await fetch(`${this.baseUrl}/api/ai/generate-schema`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.apiKey}`
        },
        body: JSON.stringify({
          image_url: imageUrl,
          size,
          style,
          quality,
          algorithm: 'adaptive-color-matching'
        })
      })

      if (!response.ok) {
        throw new Error(`Schema generation failed: ${response.statusText}`)
      }

      const result = await response.json()
      return result
    } catch (error) {
      console.error('Schema generation error:', error)
      throw new Error('Failed to generate mosaic schema')
    }
  }

  /**
   * Улучшение качества изображения
   * @param {string} imageUrl - URL изображения для улучшения
   * @param {Object} options - Параметры улучшения
   * @returns {Promise<Object>} - Результат улучшения
   */
  async enhanceImageQuality(imageUrl, options = {}) {
    try {
      const {
        upscale = 2,
        denoise = true,
        sharpen = true
      } = options

      const response = await fetch(`${this.baseUrl}/api/ai/enhance`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.apiKey}`
        },
        body: JSON.stringify({
          image_url: imageUrl,
          upscale,
          denoise,
          sharpen,
          model: 'real-esrgan'
        })
      })

      if (!response.ok) {
        throw new Error(`Image enhancement failed: ${response.statusText}`)
      }

      const result = await response.json()
      return result
    } catch (error) {
      console.error('Image enhancement error:', error)
      throw new Error('Failed to enhance image quality')
    }
  }

  /**
   * Стилизация изображения
   * @param {string} imageUrl - URL изображения для стилизации
   * @param {string} style - Стиль для применения
   * @returns {Promise<Object>} - Результат стилизации
   */
  async stylizeImage(imageUrl, style) {
    try {
      const stylePrompts = {
        'artistic': 'artistic painting style, oil painting, masterpiece',
        'pop_art': 'pop art style, bright colors, comic book style',
        'watercolor': 'watercolor painting style, soft colors, artistic',
        'sketch': 'pencil sketch style, black and white, artistic drawing'
      }

      const prompt = stylePrompts[style] || stylePrompts['artistic']

      const response = await fetch(`${this.baseUrl}/api/ai/stylize`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.apiKey}`
        },
        body: JSON.stringify({
          image_url: imageUrl,
          prompt,
          style_strength: 0.8,
          model: 'stable-diffusion-v1-5'
        })
      })

      if (!response.ok) {
        throw new Error(`Image stylization failed: ${response.statusText}`)
      }

      const result = await response.json()
      return result
    } catch (error) {
      console.error('Image stylization error:', error)
      throw new Error('Failed to stylize image')
    }
  }

  /**
   * Получение статуса обработки
   * @param {string} taskId - ID задачи обработки
   * @returns {Promise<Object>} - Статус задачи
   */
  async getProcessingStatus(taskId) {
    try {
      const response = await fetch(`${this.baseUrl}/api/ai/status/${taskId}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${this.apiKey}`
        }
      })

      if (!response.ok) {
        throw new Error(`Failed to get status: ${response.statusText}`)
      }

      const result = await response.json()
      return result
    } catch (error) {
      console.error('Status check error:', error)
      throw new Error('Failed to check processing status')
    }
  }

  /**
   * Отмена задачи обработки
   * @param {string} taskId - ID задачи для отмены
   * @returns {Promise<Object>} - Результат отмены
   */
  async cancelProcessing(taskId) {
    try {
      const response = await fetch(`${this.baseUrl}/api/ai/cancel/${taskId}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.apiKey}`
        }
      })

      if (!response.ok) {
        throw new Error(`Failed to cancel task: ${response.statusText}`)
      }

      const result = await response.json()
      return result
    } catch (error) {
      console.error('Task cancellation error:', error)
      throw new Error('Failed to cancel processing task')
    }
  }

  /**
   * Получение доступных стилей обработки
   * @returns {Promise<Array>} - Список доступных стилей
   */
  async getAvailableStyles() {
    try {
      const response = await fetch(`${this.baseUrl}/api/ai/styles`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${this.apiKey}`
        }
      })

      if (!response.ok) {
        throw new Error(`Failed to get styles: ${response.statusText}`)
      }

      const result = await response.json()
      return result.styles || []
    } catch (error) {
      console.error('Styles fetch error:', error)
      // Возвращаем базовые стили в случае ошибки
      return [
        { id: 'original', name: 'Original', description: 'No processing' },
        { id: 'enhanced', name: 'Enhanced', description: 'AI quality improvement' },
        { id: 'artistic', name: 'Artistic', description: 'Artistic stylization' }
      ]
    }
  }

  /**
   * Получение информации о моделях AI
   * @returns {Promise<Object>} - Информация о моделях
   */
  async getModelInfo() {
    try {
      const response = await fetch(`${this.baseUrl}/api/ai/models`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${this.apiKey}`
        }
      })

      if (!response.ok) {
        throw new Error(`Failed to get model info: ${response.statusText}`)
      }

      const result = await response.json()
      return result
    } catch (error) {
      console.error('Model info fetch error:', error)
      return {
        models: [],
        default: 'stable-diffusion-v1-5'
      }
    }
  }
}

// Создаем экземпляр сервиса
const aiService = new AIService()

export default aiService