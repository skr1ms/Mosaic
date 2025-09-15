import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

const AIQueueStatus = ({ imageId, onClose, isVisible }) => {
  const { t } = useTranslation();
  const [queueStatus, setQueueStatus] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (isVisible && imageId) {
      fetchQueueStatus();
      const interval = setInterval(fetchQueueStatus, 30000);
      return () => clearInterval(interval);
    }
  }, [isVisible, imageId]);

  const fetchQueueStatus = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/public/ai-queue/status');
      if (response.ok) {
        const data = await response.json();
        setQueueStatus(data);
        setError(null);
      } else {
        throw new Error('Failed to fetch queue status');
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const getHealthColor = health => {
    switch (health) {
      case 'healthy':
        return 'text-brand-secondary';
      case 'moderate':
        return 'text-yellow-600';
      case 'congested':
        return 'text-red-600';
      default:
        return 'text-gray-600';
    }
  };

  const getHealthIcon = health => {
    switch (health) {
      case 'healthy':
        return '🟢';
      case 'moderate':
        return '🟡';
      case 'congested':
        return '🔴';
      default:
        return '⚪';
    }
  };

  if (!isVisible) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4 shadow-xl">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-semibold text-gray-900">
            {t('ai_queue.status_title')}
          </h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <svg
              className="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        {loading && (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-brand-primary"></div>
          </div>
        )}

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-red-400"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-red-800">
                  {t('ai_queue.error_loading')}
                </p>
              </div>
            </div>
          </div>
        )}

        {queueStatus && !loading && (
          <div className="space-y-4">
            {}
            <div className="bg-gray-50 rounded-lg p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-700">
                  {t('ai_queue.queue_health')}
                </span>
                <span
                  className={`flex items-center text-sm font-semibold ${getHealthColor(queueStatus.queue_health)}`}
                >
                  {getHealthIcon(queueStatus.queue_health)}
                  <span className="ml-1">
                    {t(`ai_queue.health.${queueStatus.queue_health}`)}
                  </span>
                </span>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="bg-brand-primary/5 rounded-lg p-4 text-center">
                <div className="text-2xl font-bold text-brand-primary">
                  {queueStatus.queue_size}
                </div>
                <div className="text-sm text-brand-primary/80">
                  {t('ai_queue.tasks_in_queue')}
                </div>
              </div>

              <div className="bg-brand-secondary/5 rounded-lg p-4 text-center">
                <div className="text-2xl font-bold text-brand-secondary">
                  {queueStatus.available_workers}
                </div>
                <div className="text-sm text-brand-secondary/80">
                  {t('ai_queue.available_workers')}
                </div>
              </div>
            </div>

            {}
            <div className="bg-yellow-50 rounded-lg p-4">
              <div className="flex items-center">
                <svg
                  className="h-5 w-5 text-yellow-600 mr-2"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z"
                    clipRule="evenodd"
                  />
                </svg>
                <span className="text-sm font-medium text-yellow-800">
                  {t('ai_queue.estimated_wait')}:
                </span>
              </div>
              <div className="mt-1 text-lg font-semibold text-yellow-900">
                {queueStatus.estimated_wait_time}
              </div>
            </div>

            {}
            <div className="bg-brand-accent/5 rounded-lg p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-brand-accent/80">
                  {t('ai_queue.currently_processing')}
                </span>
                <span className="text-lg font-bold text-brand-accent">
                  {queueStatus.currently_processing} /{' '}
                  {queueStatus.max_concurrent}
                </span>
              </div>
              <div className="mt-2 w-full bg-brand-accent/20 rounded-full h-2">
                <div
                  className="bg-brand-accent h-2 rounded-full transition-all duration-300"
                  style={{
                    width: `${(queueStatus.currently_processing / queueStatus.max_concurrent) * 100}%`,
                  }}
                ></div>
              </div>
            </div>

            {}
            {imageId && (
              <div className="bg-brand-primary/5 rounded-lg p-4">
                <div className="flex items-center">
                  <svg
                    className="h-5 w-5 text-brand-primary mr-2"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M4 4a2 2 0 00-2 2v4a2 2 0 002 2V6h10a2 2 0 00-2-2H4zm2 6a2 2 0 012-2h8a2 2 0 012 2v4a2 2 0 01-2 2H8a2 2 0 01-2-2v-4zm6 4a2 2 0 100 4 2 2 0 000-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  <span className="text-sm font-medium text-brand-primary">
                    {t('ai_queue.your_task')}
                  </span>
                </div>
                <div className="mt-1 text-sm text-brand-primary/80">
                  ID: {imageId}
                </div>
              </div>
            )}
          </div>
        )}

        <div className="mt-6 flex space-x-3">
          <button
            onClick={fetchQueueStatus}
            disabled={loading}
            className="flex-1 bg-brand-primary text-white px-4 py-2 rounded-md hover:bg-brand-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {loading ? t('ai_queue.refreshing') : t('ai_queue.refresh')}
          </button>

          <button
            onClick={onClose}
            className="flex-1 bg-gray-300 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-400 transition-colors"
          >
            {t('ai_queue.close')}
          </button>
        </div>

        <div className="mt-4 text-xs text-gray-500 text-center">
          {t('ai_queue.auto_refresh_note')}
        </div>
      </div>
    </div>
  );
};

export default AIQueueStatus;
