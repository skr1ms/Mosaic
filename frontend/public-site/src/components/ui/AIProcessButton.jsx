import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import AIQueueStatus from './AIQueueStatus';

const AIProcessButton = ({ imageId, onProcess, disabled = false }) => {
  const { t } = useTranslation();
  const [showQueueStatus, setShowQueueStatus] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const [queueInfo, setQueueInfo] = useState(null);

  const handleAIProcess = async () => {
    if (!imageId || isProcessing) return;

    try {
      setIsProcessing(true);
      
      // Показываем модальное окно с параметрами AI обработки
      const aiParams = await showAIParametersModal();
      if (!aiParams) {
        setIsProcessing(false);
        return;
      }

      // Отправляем запрос на AI обработку
      const response = await fetch(`/api/images/${imageId}/process`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(aiParams),
      });

      if (response.ok) {
        const data = await response.json();
        setQueueInfo(data);
        setShowQueueStatus(true);
        
        // Вызываем callback если передан
        if (onProcess) {
          onProcess(data);
        }
      } else {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to start AI processing');
      }
    } catch (error) {
      console.error('AI processing error:', error);
      alert(t('ai_process.error', 'Ошибка запуска AI обработки: ') + error.message);
    } finally {
      setIsProcessing(false);
    }
  };

  const showAIParametersModal = () => {
    return new Promise((resolve) => {
      // Создаем модальное окно с параметрами
      const modal = document.createElement('div');
      modal.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
      
      const modalContent = `
        <div class="bg-white rounded-lg p-6 max-w-md w-full mx-4 shadow-xl">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">
            ${t('ai_process.parameters_title', 'Параметры AI обработки')}
          </h3>
          
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                ${t('ai_process.style', 'Стиль обработки')}
              </label>
              <select id="ai-style" class="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-brand-primary">
                <option value="grayscale">${t('ai_process.style.grayscale', 'Оттенки серого')}</option>
                <option value="skin_tones">${t('ai_process.style.skin_tones', 'Оттенки телесного')}</option>
                <option value="pop_art">${t('ai_process.style.pop_art', 'Поп-арт')}</option>
                <option value="max_colors">${t('ai_process.style.max_colors', 'Максимум цветов')}</option>
              </select>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                ${t('ai_process.lighting', 'Освещение')}
              </label>
              <select id="ai-lighting" class="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-brand-primary">
                <option value="">${t('ai_process.lighting.none', 'Без изменений')}</option>
                <option value="sun">${t('ai_process.lighting.sun', 'Солнечное')}</option>
                <option value="moon">${t('ai_process.lighting.moon', 'Лунное')}</option>
                <option value="venus">${t('ai_process.lighting.venus', 'Венерианское')}</option>
              </select>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                ${t('ai_process.contrast', 'Контрастность')}
              </label>
              <select id="ai-contrast" class="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-brand-primary">
                <option value="">${t('ai_process.contrast.none', 'Без изменений')}</option>
                <option value="low">${t('ai_process.contrast.low', 'Низкий')}</option>
                <option value="high">${t('ai_process.contrast.high', 'Высокий')}</option>
              </select>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                ${t('ai_process.brightness', 'Яркость')} (-100 до 100)
              </label>
              <input type="range" id="ai-brightness" min="-100" max="100" value="0" class="w-full">
              <span id="brightness-value" class="text-sm text-gray-600">0</span>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                ${t('ai_process.saturation', 'Насыщенность')} (-100 до 100)
              </label>
              <input type="range" id="ai-saturation" min="-100" max="100" value="0" class="w-full">
              <span id="saturation-value" class="text-sm text-gray-600">0</span>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                ${t('ai_process.priority', 'Приоритет')}
              </label>
              <select id="ai-priority" class="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-brand-primary">
                <option value="5">${t('ai_process.priority.normal', 'Обычный')}</option>
                <option value="7">${t('ai_process.priority.high', 'Высокий')}</option>
                <option value="10">${t('ai_process.priority.urgent', 'Срочный')}</option>
              </select>
            </div>
          </div>
          
          <div class="mt-6 flex space-x-3">
            <button id="ai-process-start" class="flex-1 bg-brand-primary text-white px-4 py-2 rounded-md hover:bg-brand-primary/90 transition-colors">
              ${t('ai_process.start', 'Запустить AI обработку')}
            </button>
            <button id="ai-process-cancel" class="flex-1 bg-gray-300 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-400 transition-colors">
              ${t('ai_process.cancel', 'Отмена')}
            </button>
          </div>
        </div>
      `;
      
      modal.innerHTML = modalContent;
      document.body.appendChild(modal);
      
      // Добавляем обработчики событий
      const brightnessSlider = modal.querySelector('#ai-brightness');
      const brightnessValue = modal.querySelector('#brightness-value');
      const saturationSlider = modal.querySelector('#ai-saturation');
      const saturationValue = modal.querySelector('#ai-saturation-value');
      
      brightnessSlider.addEventListener('input', (e) => {
        brightnessValue.textContent = e.target.value;
      });
      
      saturationSlider.addEventListener('input', (e) => {
        saturationValue.textContent = e.target.value;
      });
      
      // Обработчик запуска
      modal.querySelector('#ai-process-start').addEventListener('click', () => {
        const params = {
          style: modal.querySelector('#ai-style').value,
          use_ai: true,
          lighting: modal.querySelector('#ai-lighting').value,
          contrast: modal.querySelector('#ai-contrast').value,
          brightness: parseFloat(modal.querySelector('#ai-brightness').value),
          saturation: parseFloat(modal.querySelector('#ai-saturation').value),
          priority: parseInt(modal.querySelector('#ai-priority').value),
        };
        
        document.body.removeChild(modal);
        resolve(params);
      });
      
      // Обработчик отмены
      modal.querySelector('#ai-process-cancel').addEventListener('click', () => {
        document.body.removeChild(modal);
        resolve(null);
      });
    });
  };

  return (
    <>
      <button
        onClick={handleAIProcess}
        disabled={disabled || isProcessing}
        className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-gradient-to-r from-brand-primary to-brand-secondary hover:from-brand-primary/90 hover:to-brand-secondary/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-brand-primary disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
      >
        {isProcessing ? (
          <>
            <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            {t('ai_process.processing', 'Обработка...')}
          </>
        ) : (
          <>
            <svg className="-ml-1 mr-2 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            {t('ai_process.start_ai', 'AI улучшение')}
          </>
        )}
      </button>

      {/* Модальное окно статуса очереди */}
      <AIQueueStatus
        imageId={imageId}
        isVisible={showQueueStatus}
        onClose={() => setShowQueueStatus(false)}
      />
    </>
  );
};

export default AIProcessButton;
