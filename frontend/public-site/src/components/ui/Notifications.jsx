import React, { useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X, CheckCircle, AlertCircle, Info } from 'lucide-react';
import { useUIStore } from '../../store/partnerStore';

const Notifications = () => {
  const { notifications, removeNotification } = useUIStore();

  const getIcon = type => {
    switch (type) {
      case 'success':
        return (
          <CheckCircle className="w-4 h-4 sm:w-5 sm:h-5 text-brand-secondary" />
        );
      case 'error':
        return <AlertCircle className="w-4 h-4 sm:w-5 sm:h-5 text-red-500" />;
      default:
        return <Info className="w-4 h-4 sm:w-5 sm:h-5 text-brand-primary" />;
    }
  };

  const getColors = type => {
    switch (type) {
      case 'success':
        return 'border-brand-secondary/20 bg-brand-secondary/5';
      case 'error':
        return 'border-red-200 bg-red-50';
      default:
        return 'border-brand-primary/20 bg-brand-primary/5';
    }
  };

  return (
    <div className="fixed top-4 right-4 left-4 md:left-auto z-[9999] space-y-3 max-w-sm md:max-w-none mx-auto md:mx-0">
      <AnimatePresence>
        {notifications.map(notification => (
          <NotificationItem
            key={notification.id}
            notification={notification}
            onRemove={removeNotification}
            getIcon={getIcon}
            getColors={getColors}
          />
        ))}
      </AnimatePresence>
    </div>
  );
};

const NotificationItem = ({ notification, onRemove, getIcon, getColors }) => {
  useEffect(() => {
    const timer = setTimeout(() => {
      onRemove(notification.id);
    }, notification.duration || 5000);

    return () => clearTimeout(timer);
  }, [notification.id, notification.duration, onRemove]);

  return (
    <motion.div
      initial={{ opacity: 0, x: 300, scale: 0.8 }}
      animate={{ opacity: 1, x: 0, scale: 1 }}
      exit={{ opacity: 0, x: 300, scale: 0.8 }}
      transition={{ duration: 0.3, ease: 'easeOut' }}
      className={`flex items-start space-x-2 sm:space-x-3 p-3 sm:p-4 rounded-lg border shadow-lg w-full md:min-w-80 md:max-w-sm ${getColors(notification.type)}`}
    >
      <div className="flex-shrink-0">{getIcon(notification.type)}</div>

      <div className="flex-1 min-w-0">
        {notification.title && (
          <p className="text-xs sm:text-sm font-medium text-gray-900 mb-1 break-words">
            {notification.title}
          </p>
        )}
        <p className="text-xs sm:text-sm text-gray-700 break-words">
          {notification.message}
        </p>
      </div>

      <button
        onClick={() => onRemove(notification.id)}
        className="flex-shrink-0 p-1 sm:p-2 text-gray-400 hover:text-gray-600 transition-colors touch-manipulation"
      >
        <X className="w-3 h-3 sm:w-4 sm:h-4" />
      </button>
    </motion.div>
  );
};

export default Notifications;
