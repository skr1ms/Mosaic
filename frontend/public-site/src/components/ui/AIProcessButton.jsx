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

      const aiParams = await showAIParametersModal();
      if (!aiParams) {
        setIsProcessing(false);
        return;
      }

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

        if (onProcess) {
          onProcess(data);
        }
      } else {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to start AI processing');
      }
    } catch (error) {
      console.error('AI processing error:', error);
      alert(t('ai_process.error') + error.message);
    } finally {
      setIsProcessing(false);
    }
  };

  const showAIParametersModal = () => {
    return new Promise(resolve => {
      const modal = document.createElement('div');
      modal.className =
        'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';

      const modalContent = `
        <div class="bg-white rounded-lg p-4 sm:p-6 max-w-md w-full mx-4 shadow-xl max-h-[90vh] overflow-y-auto">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">
            ${t('ai_process.parameters_title')}
          </h3>
          
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                ${t('ai_process.style')}
              </label>
              <select id="ai-style" class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-primary" style="font-size: 16px;">
                <option value="grayscale">${t('ai_process.style.grayscale')}</option>
                <option value="skin_tones">${t('ai_process.style.skin_tones')}</option>
                <option value="pop_art">${t('ai_process.style.pop_art')}</option>
                <option value="max_colors">${t('ai_process.style.max_colors')}</option>
              </select>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                ${t('ai_process.lighting')}
              </label>
              <select id="ai-lighting" class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-primary" style="font-size: 16px;">
                <option value="">${t('ai_process.lighting.none')}</option>
                <option value="sun">${t('ai_process.lighting.sun')}</option>
                <option value="moon">${t('ai_process.lighting.moon')}</option>
                <option value="venus">${t('ai_process.lighting.venus')}</option>
              </select>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                ${t('ai_process.contrast')}
              </label>
              <select id="ai-contrast" class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-primary" style="font-size: 16px;">
                <option value="">${t('ai_process.contrast.none')}</option>
                <option value="low">${t('ai_process.contrast.low')}</option>
                <option value="high">${t('ai_process.contrast.high')}</option>
              </select>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                ${t('ai_process.brightness')} (-100 до 100)
              </label>
              <input type="range" id="ai-brightness" min="-100" max="100" value="0" class="w-full">
              <span id="brightness-value" class="text-sm text-gray-600">0</span>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                ${t('ai_process.saturation')} (-100 до 100)
              </label>
              <input type="range" id="ai-saturation" min="-100" max="100" value="0" class="w-full">
              <span id="saturation-value" class="text-sm text-gray-600">0</span>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                ${t('ai_process.priority')}
              </label>
              <select id="ai-priority" class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-primary" style="font-size: 16px;">
                <option value="5">${t('ai_process.priority.normal')}</option>
                <option value="7">${t('ai_process.priority.high')}</option>
                <option value="10">${t('ai_process.priority.urgent')}</option>
              </select>
            </div>
          </div>
          
          <div class="mt-6 flex flex-col sm:flex-row space-y-2 sm:space-y-0 sm:space-x-3">
            <button id="ai-process-start" class="flex-1 bg-brand-primary text-white px-4 py-3 sm:py-2 rounded-md hover:bg-brand-primary/90 active:bg-brand-primary/80 transition-colors text-sm font-medium touch-target">
              ${t('ai_process.start')}
            </button>
            <button id="ai-process-cancel" class="flex-1 bg-gray-300 text-gray-700 px-4 py-3 sm:py-2 rounded-md hover:bg-gray-400 active:bg-gray-500 transition-colors text-sm font-medium touch-target">
              ${t('ai_process.cancel')}
            </button>
          </div>
        </div>
      `;

      modal.innerHTML = modalContent;
      document.body.appendChild(modal);

      const brightnessSlider = modal.querySelector('#ai-brightness');
      const brightnessValue = modal.querySelector('#brightness-value');
      const saturationSlider = modal.querySelector('#ai-saturation');
      const saturationValue = modal.querySelector('#ai-saturation-value');

      brightnessSlider.addEventListener('input', e => {
        brightnessValue.textContent = e.target.value;
      });

      saturationSlider.addEventListener('input', e => {
        saturationValue.textContent = e.target.value;
      });

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

      modal
        .querySelector('#ai-process-cancel')
        .addEventListener('click', () => {
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
        className="inline-flex items-center justify-center w-full sm:w-auto px-3 sm:px-4 py-2 sm:py-2 border border-transparent text-xs sm:text-sm font-medium rounded-md shadow-sm text-white bg-gradient-to-r from-brand-primary to-brand-secondary hover:from-brand-primary/90 hover:to-brand-secondary/90 active:from-brand-primary/80 active:to-brand-secondary/80 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-brand-primary disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 touch-target"
      >
        {isProcessing ? (
          <>
            <svg
              className="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              ></circle>
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {t('ai_process.processing')}
          </>
        ) : (
          <>
            <svg
              className="-ml-1 mr-2 h-4 w-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M13 10V3L4 14h7v7l9-11h-7z"
              />
            </svg>
            {t('ai_process.start_ai')}
          </>
        )}
      </button>

      {}
      <AIQueueStatus
        imageId={imageId}
        isVisible={showQueueStatus}
        onClose={() => setShowQueueStatus(false)}
      />
    </>
  );
};

export default AIProcessButton;
