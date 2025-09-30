import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

console.log('Public Site API Base URL:', API_BASE_URL);
console.log('Environment:', import.meta.env.VITE_ENVIRONMENT);
console.log('Debug mode:', import.meta.env.VITE_DEBUG);

apiClient.interceptors.response.use(
  response => response,
  error => {
    const status = error?.response?.status;
    const data = error?.response?.data || {};
    const problemType = data?.type;
    const message =
      data?.detail ||
      data?.title ||
      data?.message ||
      error.message ||
      'Request failed';
    const shaped = new Error(message);
    shaped.name = error?.name || 'AxiosError';
    shaped.status = status;
    shaped.problemType = problemType;
    shaped.original = error;
    console.error('API Error:', { status, problemType, message });
    return Promise.reject(shaped);
  }
);

export class MosaicAPI {
  static async startSupportChat(title = '') {
    const { data } = await apiClient.post(
      '/public/support/start' +
        (title ? `?title=${encodeURIComponent(title)}` : '')
    );
    return data;
  }

  static async getSupportMessages(chatId, token) {
    const { data } = await apiClient.get(`/public/support/messages`, {
      params: { chat_id: chatId },
      headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    });
    return data?.messages || data?.data?.messages || data?.data || [];
  }

  static async sendSupportMessage(chatId, content, token) {
    const { data } = await apiClient.post(
      `/public/support/messages`,
      { chat_id: chatId, content },
      {
        headers: token ? { Authorization: `Bearer ${token}` } : undefined,
      }
    );
    return data?.message || data?.data || data;
  }

  static async updateSupportMessage(id, content, token) {
    const { data } = await apiClient.patch(
      `/public/support/messages/${id}`,
      { content },
      {
        headers: token ? { Authorization: `Bearer ${token}` } : undefined,
      }
    );
    return data?.message || data?.data || data;
  }

  static async deleteSupportMessage(id, token) {
    const { data } = await apiClient.delete(`/public/support/messages/${id}`, {
      headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    });
    return data?.result || data?.data || data;
  }
  static async validateCoupon(code) {
    const { data } = await apiClient.post(`/coupons/code/${code}/validate`);
    return data;
  }

  static async activateCoupon(idOrCode, activationData) {
    const isUuid = idOrCode.includes('-') || idOrCode.length === 36;
    const endpoint = isUuid 
      ? `/coupons/${idOrCode}/activate`
      : `/coupons/code/${idOrCode}/activate`;
    
    const { data } = await apiClient.put(endpoint, activationData);
    return data;
  }

  static async getCouponByCode(code) {
    const { data } = await apiClient.get(`/coupons/code/${code}`);
    return data;
  }

  static async sendSchemaToEmail(couponId, payload) {
    const { data } = await apiClient.put(
      `/coupons/${couponId}/send-schema`,
      payload
    );
    return data;
  }

  static async reactivateCoupon(code, email = null) {
    const payload = { code };
    if (email) {
      payload.email = email;
    }
    const { data } = await apiClient.post(
      `/coupons/code/${code}/reactivate`,
      payload
    );
    return data;
  }

  static async searchSchemaPage(imageId, pageNumber) {
    const { data } = await apiClient.post(`/images/${imageId}/search-page`, {
      image_id: imageId,
      page_number: pageNumber,
    });
    return data;
  }

  static async getBrandingInfo() {
    const { data } = await apiClient.get(`/branding`);
    return data;
  }

  static async uploadImage(formData) {
    const { data } = await apiClient.post('/images/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return data;
  }

  static async generatePreview(formData) {
    const { data } = await apiClient.post('/preview/generate-variant', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      timeout: 120000,
    });
    return data;
  }

  static async generatePreviewVariant(formData) {
    const { data } = await apiClient.post(
      '/preview/generate-variant',
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        timeout: 120000,
      }
    );
    return data;
  }

  static async generateAIPreview(formData) {
    const { data } = await apiClient.post('/preview/generate-ai', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      timeout: 180000,
    });
    return data;
  }

  static async generateStyleVariants(formData) {
    const { data } = await apiClient.post('/preview/generate-style-variants', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      timeout: 180000,
    });
    return data;
  }

  static async generateAllPreviews(imageId, size, useAI = false, imageFile = null) {
    const formData = new FormData();
    
    if (imageId) {
      formData.append('image_id', imageId);
    }
    
    formData.append('size', size);
    formData.append('use_ai', useAI.toString());
    
    if (imageFile) {
      formData.append('image', imageFile, 'image.jpg');
    }

    const { data } = await apiClient.post('/preview/generate-all', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      timeout: 180000,
    });
    return data;
  }

  static async generateVariants(formData) {
    const { data } = await apiClient.post(
      '/preview/generate-variants',
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        timeout: 120000,
      }
    );
    return data;
  }

  static async editImage(imageId, params) {
    const { data } = await apiClient.post(`/images/${imageId}/edit`, params);
    return data;
  }

  static async processImage(imageId, params) {
    const { data } = await apiClient.post(`/images/${imageId}/process`, params);
    return data;
  }

  static async generateSchema(imageId, params) {
    const { data } = await apiClient.post(
      `/images/${imageId}/generate-schema`,
      params
    );
    return data;
  }

  static async getProcessingStatus(imageId) {
    const { data } = await apiClient.get(`/images/${imageId}/status`);
    return data;
  }

  static async downloadSchemaArchive(schemaUUID) {
    const url = `${API_BASE_URL}/public/schemas/${schemaUUID}/download`;
    window.open(url, '_blank');
  }

  static async sendSchemaToEmail(imageId, payload) {
    const body = typeof payload === 'string' ? { email: payload } : payload;
    const { data } = await apiClient.post(
      `/images/${imageId}/send-email`,
      body
    );
    return data;
  }

  static async initiateCouponOrder(params) {
    const { data } = await apiClient.post('/payment/purchase', params);
    return data;
  }

  static async getOrderStatus(orderNumber) {
    const { data } = await apiClient.get(
      `/payment/orders/${orderNumber}/status`
    );
    return data;
  }

  static async getAvailableSizes() {
    const { data } = await apiClient.get('/sizes');
    return data;
  }

  static async getAvailableStyles() {
    const { data } = await apiClient.get('/styles');
    return data;
  }

  static async getPartnerArticleGrid(partnerId) {
    const { data } = await apiClient.get(
      `/admin/partners/${partnerId}/articles/grid`
    );
    return data;
  }

  static async generateProductURL(partnerId, params) {
    const { data } = await apiClient.post(
      `/partners/${partnerId}/articles/generate-url`,
      params
    );
    return data;
  }

  static async checkMarketplaceStatus(params) {
    const { data } = await apiClient.get('/marketplace/status', {
      params: {
        marketplace: params.marketplace,
        partner_id: params.partnerId,
        size: params.size,
        style: params.style,
        sku: params.sku
      }
    });
    return data;
  }
}

export default apiClient;
