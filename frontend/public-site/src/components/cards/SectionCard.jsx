import React from 'react';
import { motion } from 'framer-motion';
import { Clock } from 'lucide-react';

const SectionCard = ({
  icon,
  title,
  description,
  buttonText,
  buttonIcon,
  onClick,
  className = '',
  gradient = 'from-blue-500 to-purple-500',
  active = false,
  disabled = false,
  comingSoon,
}) => {
  return (
    <motion.div
      whileHover={!disabled ? { y: -4, scale: 1.02 } : {}}
      whileTap={!disabled ? { scale: 0.98 } : {}}
      className={`relative bg-white/80 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 overflow-hidden ${
        disabled ? 'opacity-60' : 'hover:shadow-2xl'
      } transition-all duration-300 ${className} h-full min-h-[400px]`}
    >
      <div className="p-6 sm:p-8 flex flex-col h-full">
        <div
          className={`w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-gradient-to-r ${active ? 'from-brand-primary to-brand-secondary' : gradient} flex items-center justify-center text-white mb-4 sm:mb-6 mx-auto`}
        >
          {icon}
        </div>

        <h3 className="text-xl sm:text-2xl font-bold text-gray-900 mb-4 px-2 text-center">
          {title}
        </h3>

        <p className="text-gray-600 mb-6 px-2 text-sm sm:text-base text-center flex-grow">
          {description}
        </p>

        <div className="mt-auto">
          {disabled && comingSoon ? (
            <div className="flex items-center justify-center space-x-2 py-3 px-4 sm:px-6 bg-gray-100 text-gray-500 rounded-lg min-h-[48px]">
              <Clock className="w-5 h-5" />
              <span className="font-medium">{comingSoon}</span>
            </div>
          ) : (
            buttonText && (
              <button
                onClick={onClick}
                disabled={disabled}
                className={`section-btn w-full py-3 sm:py-4 px-4 sm:px-6 rounded-lg font-semibold text-sm sm:text-base transition-all duration-200 flex items-center justify-center space-x-2 focus:ring-2 focus:ring-offset-2 touch-target ${
                  active
                    ? `bg-gradient-to-r from-brand-primary to-brand-secondary text-white hover:from-brand-primary/90 hover:to-brand-secondary/90 active:from-brand-primary/80 active:to-brand-secondary/80 hover:shadow-lg focus:ring-brand-primary`
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200 active:bg-gray-300 focus:ring-gray-300'
                } ${disabled ? 'cursor-not-allowed' : 'cursor-pointer'}`}
              >
                {buttonIcon && (
                  <span className="flex items-center shrink-0">
                    {buttonIcon}
                  </span>
                )}
                <span className="btn-text-safe">{buttonText}</span>
              </button>
            )
          )}
        </div>
      </div>

      {active && (
        <div className="absolute top-4 right-4">
          <div className="w-3 h-3 bg-brand-secondary rounded-full animate-pulse" />
        </div>
      )}
    </motion.div>
  );
};

export default SectionCard;
