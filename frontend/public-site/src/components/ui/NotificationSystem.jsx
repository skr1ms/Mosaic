import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { CheckCircle, XCircle, AlertTriangle, Info, X } from 'lucide-react';

const NotificationSystem = ({ notifications, onRemove }) => {
  const getIcon = type => {
    switch (type) {
      case 'success':
        return <CheckCircle className="w-5 h-5 text-brand-secondary" />;
      case 'error':
        return <XCircle className="w-5 h-5 text-red-600" />;
      case 'warning':
        return <AlertTriangle className="w-5 h-5 text-yellow-600" />;
      case 'info':
        return <Info className="w-5 h-5 text-brand-primary" />;
      default:
        return <Info className="w-5 h-5 text-brand-primary" />;
    }
  };

  const getBackgroundColor = type => {
    switch (type) {
      case 'success':
        return 'bg-brand-secondary/5 border-brand-secondary/20';
      case 'error':
        return 'bg-red-50 border-red-200';
      case 'warning':
        return 'bg-yellow-50 border-yellow-200';
      case 'info':
        return 'bg-brand-primary/5 border-brand-primary/20';
      default:
        return 'bg-brand-primary/5 border-brand-primary/20';
    }
  };

  const getTextColor = type => {
    switch (type) {
      case 'success':
        return 'text-brand-secondary/80';
      case 'error':
        return 'text-red-800';
      case 'warning':
        return 'text-yellow-800';
      case 'info':
        return 'text-brand-primary/80';
      default:
        return 'text-brand-primary/80';
    }
  };

  return (
    <div className="fixed top-4 left-4 right-4 sm:left-auto sm:right-4 sm:w-auto z-50 space-y-2 sm:min-w-[300px] sm:max-w-[400px]">
      <AnimatePresence>
        {notifications.map(notification => (
          <motion.div
            key={`${notification.id}-${notification.dedupeKey || notification.message}`}
            initial={{ opacity: 0, x: 100, scale: 0.8 }}
            animate={{ opacity: 1, x: 0, scale: 1 }}
            exit={{ opacity: 0, x: 100, scale: 0.8 }}
            transition={{
              type: 'spring',
              stiffness: 300,
              damping: 25,
              duration: 0.3,
            }}
            className={`w-full px-3 sm:px-4 py-3 rounded-lg border ${getBackgroundColor(notification.type)} shadow-lg backdrop-blur-sm`}
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2 flex-1 min-w-0">
                <motion.div
                  className="flex-shrink-0"
                  initial={{ scale: 0 }}
                  animate={{ scale: 1 }}
                  transition={{ delay: 0.1, type: 'spring', stiffness: 400 }}
                >
                  {getIcon(notification.type)}
                </motion.div>

                <div className="flex-1 min-w-0">
                  {notification.message ? (
                    <motion.p
                      className={`text-sm font-medium ${getTextColor(notification.type)} truncate`}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: 0.15 }}
                    >
                      {notification.message}
                    </motion.p>
                  ) : notification.title ? (
                    <motion.p
                      className={`text-sm font-medium ${getTextColor(notification.type)} truncate`}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: 0.15 }}
                    >
                      {notification.title}
                    </motion.p>
                  ) : null}
                </div>
              </div>

              <div className="flex-shrink-0 ml-2">
                <motion.button
                  onClick={() => onRemove(notification.id)}
                  className={`inline-flex rounded-md p-1 ${getTextColor(notification.type)} hover:bg-white hover:bg-opacity-50 transition-colors`}
                  whileHover={{ scale: 1.1 }}
                  whileTap={{ scale: 0.9 }}
                >
                  <X className="w-4 h-4" />
                </motion.button>
              </div>
            </div>
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  );
};

export default NotificationSystem;
