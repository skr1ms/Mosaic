import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import {
  Palette,
  Sun,
  Moon,
  Sparkles,
  Contrast,
  Check,
  Clock,
} from 'lucide-react';
import { useUIStore } from '../../store/partnerStore';

const StyleSelector = ({
  imageId,
  onStyleSelected,
  onBack,
  initialOptions = null,
}) => {
  const { t } = useTranslation();
  const { addNotification } = useUIStore();
  const [selectedStyle, setSelectedStyle] = useState(
    initialOptions?.style || null
  );
  const [selectedLighting, setSelectedLighting] = useState(
    initialOptions?.lighting || null
  );
  const [selectedContrast, setSelectedContrast] = useState(
    initialOptions?.contrast || null
  );

  React.useEffect(() => {
    if (!imageId || initialOptions) return;
    try {
      const raw = sessionStorage.getItem(`editor:selectedOptions:${imageId}`);
      if (raw) {
        const opts = JSON.parse(raw);
        setSelectedStyle(opts.style || null);
        setSelectedLighting(opts.lighting || null);
        setSelectedContrast(opts.contrast || null);
      }
    } catch {}
  }, [imageId]);

  const mapToBackendParams = ui => {
    const styleMap = {
      original: { style: 'grayscale', use_ai: false },
      enhanced: { style: 'max_colors', use_ai: true },
    };
    const lightingMap = {
      natural: 'sun',
      moonlight: 'moon',
      venus: 'venus',
    };
    const contrastMap = {
      soft: 'low',
      strong: 'high',
    };

    const mapped = styleMap[ui.style] || { style: 'grayscale', use_ai: false };
    const params = {
      style: mapped.style,
      use_ai: mapped.use_ai,
      lighting: lightingMap[ui.lighting] || 'sun',
    };
    const mappedContrast = contrastMap[ui.contrast];
    if (mappedContrast) params.contrast = mappedContrast;
    return params;
  };

  const handleStyleSelect = style => {
    setSelectedStyle(style);
    try {
      const raw = sessionStorage.getItem(`editor:selectedOptions:${imageId}`);
      const current = raw ? JSON.parse(raw) : {};
      sessionStorage.setItem(
        `editor:selectedOptions:${imageId}`,
        JSON.stringify({ ...current, style })
      );
    } catch {}
  };

  const handleLightingSelect = lighting => {
    setSelectedLighting(lighting);
    try {
      const raw = sessionStorage.getItem(`editor:selectedOptions:${imageId}`);
      const current = raw ? JSON.parse(raw) : {};
      sessionStorage.setItem(
        `editor:selectedOptions:${imageId}`,
        JSON.stringify({ ...current, lighting })
      );
    } catch {}
  };

  const handleContrastSelect = contrast => {
    setSelectedContrast(contrast);
    try {
      const raw = sessionStorage.getItem(`editor:selectedOptions:${imageId}`);
      const current = raw ? JSON.parse(raw) : {};
      sessionStorage.setItem(
        `editor:selectedOptions:${imageId}`,
        JSON.stringify({ ...current, contrast })
      );
    } catch {}
  };

  const handleContinue = () => {
    if (selectedStyle) {
      onStyleSelected({
        style: selectedStyle,
        lighting: selectedLighting || 'natural',
        contrast: selectedContrast || 'soft',
      });
    }
  };

  const styles = [
    {
      id: 'original',
      name: t('editor.styles.original.title'),
      description: t('editor.styles.original.description'),
      icon: Palette,
      color: 'from-gray-400 to-gray-600',
      processingTime: t('editor.processing_time_fast', '1-2 min'),
      useAI: false,
    },
    {
      id: 'enhanced',
      name: t('editor.styles.enhanced.title'),
      description: t('editor.styles.enhanced.description'),
      icon: Sparkles,
      color: 'from-blue-400 to-purple-600',
      processingTime: t('editor.processing_time', '2-5 min'),
      useAI: true,
    },
  ];

  const lightingOptions = [
    {
      id: 'natural',
      name: t('editor.lighting.natural.title'),
      description: t('editor.lighting.natural.description'),
      icon: Sun,
      color: 'from-yellow-400 to-orange-500',
    },
    {
      id: 'moonlight',
      name: t('editor.lighting.moonlight.title'),
      description: t('editor.lighting.moonlight.description'),
      icon: Moon,
      color: 'from-blue-400 to-indigo-500',
    },
    {
      id: 'venus',
      name: t('editor.lighting.venus.title'),
      description: t('editor.lighting.venus.description'),
      icon: Sparkles,
      color: 'from-pink-400 to-purple-500',
    },
  ];

  const contrastOptions = [
    {
      id: 'soft',
      name: t('editor.contrast.soft.title'),
      description: t('editor.contrast.soft.description'),
      icon: Contrast,
      color: 'from-gray-300 to-gray-500',
    },
    {
      id: 'strong',
      name: t('editor.contrast.strong.title'),
      description: t('editor.contrast.strong.description'),
      icon: Contrast,
      color: 'from-gray-600 to-gray-800',
    },
  ];

  return (
    <div className="space-y-6 sm:space-y-8">
      {}
      <div className="text-center px-4">
        <h2 className="text-2xl sm:text-3xl font-bold text-gray-900 mb-3 sm:mb-4 leading-tight">
          {t('editor.styles.title')}
        </h2>
        <p className="text-base sm:text-lg text-gray-600 max-w-2xl mx-auto leading-relaxed">
          {t('editor.styles.subtitle')}
        </p>
      </div>

      <section>
        <h3 className="text-lg sm:text-xl font-semibold text-gray-900 mb-4 sm:mb-6 text-center leading-tight">
          {t('editor.styles.main_style')}
        </h3>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6 max-w-2xl mx-auto">
          {styles.map(style => (
            <motion.div
              key={style.id}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              onClick={() => handleStyleSelect(style.id)}
              className={`relative cursor-pointer rounded-xl p-4 sm:p-6 border-2 transition-all duration-300 touch-target ${
                selectedStyle === style.id
                  ? 'border-brand-primary bg-brand-primary/5 shadow-sm'
                  : 'border-gray-200 hover:border-gray-300 active:bg-gray-50 bg-white'
              }`}
            >
              {selectedStyle === style.id && (
                <div className="absolute -top-1 -right-1 sm:-top-2 sm:-right-2 w-5 h-5 sm:w-6 sm:h-6 bg-brand-primary rounded-full flex items-center justify-center">
                  <Check className="w-3 h-3 sm:w-4 sm:h-4 text-white" />
                </div>
              )}

              <div className="text-center">
                <div
                  className={`w-12 h-12 sm:w-16 sm:h-16 mx-auto mb-3 sm:mb-4 rounded-lg bg-gradient-to-br ${style.color} flex items-center justify-center`}
                >
                  <style.icon className="w-6 h-6 sm:w-8 sm:h-8 text-white" />
                </div>

                <h4 className="text-base sm:text-lg font-semibold text-gray-900 mb-2 leading-tight">
                  {style.name}
                </h4>

                <p className="text-xs sm:text-sm text-gray-600 mb-2 sm:mb-3 leading-relaxed">
                  {style.description}
                </p>

                {}
                <div
                  className={`inline-flex items-center space-x-1 px-2 py-1 rounded-full text-xs font-medium ${
                    style.useAI
                      ? 'bg-brand-primary/10 text-brand-primary border border-brand-primary/20'
                      : 'bg-gray-100 text-gray-600 border border-gray-200'
                  }`}
                >
                  <Clock className="w-3 h-3 flex-shrink-0" />
                  <span>{style.processingTime}</span>
                  {style.useAI && <span className="ml-1">🤖</span>}
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </section>

      {}
      {selectedStyle && (
        <motion.section
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
        >
          <h3 className="text-lg sm:text-xl font-semibold text-gray-900 mb-4 sm:mb-6 text-center leading-tight">
            {t('editor.lighting.title')}
          </h3>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4 lg:gap-6">
            {lightingOptions.map(lighting => (
              <motion.div
                key={lighting.id}
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                onClick={() => handleLightingSelect(lighting.id)}
                className={`relative cursor-pointer rounded-xl p-4 sm:p-6 border-2 transition-all duration-300 touch-target ${
                  selectedLighting === lighting.id
                    ? 'border-brand-primary bg-brand-primary/5 shadow-sm'
                    : 'border-gray-200 hover:border-gray-300 active:bg-gray-50 bg-white'
                }`}
              >
                {selectedLighting === lighting.id && (
                  <div className="absolute -top-1 -right-1 sm:-top-2 sm:-right-2 w-5 h-5 sm:w-6 sm:h-6 bg-brand-primary rounded-full flex items-center justify-center">
                    <Check className="w-3 h-3 sm:w-4 sm:h-4 text-white" />
                  </div>
                )}

                <div className="text-center">
                  <div
                    className={`w-12 h-12 sm:w-16 sm:h-16 mx-auto mb-3 sm:mb-4 rounded-lg bg-gradient-to-br ${lighting.color} flex items-center justify-center`}
                  >
                    <lighting.icon className="w-6 h-6 sm:w-8 sm:h-8 text-white" />
                  </div>

                  <h4 className="text-base sm:text-lg font-semibold text-gray-900 mb-2 leading-tight">
                    {lighting.name}
                  </h4>

                  <p className="text-xs sm:text-sm text-gray-600 leading-relaxed">
                    {lighting.description}
                  </p>
                </div>
              </motion.div>
            ))}
          </div>
        </motion.section>
      )}

      {selectedStyle && (
        <motion.section
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.2 }}
        >
          <h3 className="text-lg sm:text-xl font-semibold text-gray-900 mb-4 sm:mb-6 text-center leading-tight">
            {t('editor.contrast.title')}
          </h3>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4 lg:gap-6">
            {contrastOptions.map(contrast => (
              <motion.div
                key={contrast.id}
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                onClick={() => handleContrastSelect(contrast.id)}
                className={`relative cursor-pointer rounded-xl p-4 sm:p-6 border-2 transition-all duration-300 touch-target ${
                  selectedContrast === contrast.id
                    ? 'border-brand-primary bg-brand-primary/5 shadow-sm'
                    : 'border-gray-200 hover:border-gray-300 active:bg-gray-50 bg-white'
                }`}
              >
                {selectedContrast === contrast.id && (
                  <div className="absolute -top-1 -right-1 sm:-top-2 sm:-right-2 w-5 h-5 sm:w-6 sm:h-6 bg-brand-primary rounded-full flex items-center justify-center">
                    <Check className="w-3 h-3 sm:w-4 sm:h-4 text-white" />
                  </div>
                )}

                <div className="text-center">
                  <div
                    className={`w-12 h-12 sm:w-16 sm:h-16 mx-auto mb-3 sm:mb-4 rounded-lg bg-gradient-to-br ${contrast.color} flex items-center justify-center`}
                  >
                    <contrast.icon className="w-6 h-6 sm:w-8 sm:h-8 text-white" />
                  </div>

                  <h4 className="text-base sm:text-lg font-semibold text-gray-900 mb-2 leading-tight">
                    {contrast.name}
                  </h4>

                  <p className="text-xs sm:text-sm text-gray-600 leading-relaxed">
                    {contrast.description}
                  </p>
                </div>
              </motion.div>
            ))}
          </div>
        </motion.section>
      )}

      <div className="flex flex-col sm:flex-row items-center justify-between pt-6 sm:pt-8 border-t gap-3 sm:gap-4">
        <button
          onClick={onBack}
          className="w-full sm:w-auto flex items-center justify-center space-x-2 px-4 sm:px-6 py-3 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 active:bg-gray-100 transition-colors text-sm sm:text-base font-medium touch-target"
        >
          <span>{t('common.back')}</span>
        </button>

        <button
          onClick={handleContinue}
          disabled={!selectedStyle}
          className={`w-full sm:w-auto px-6 sm:px-8 py-3 rounded-lg font-semibold transition-all duration-300 text-sm sm:text-base touch-target ${
            selectedStyle
              ? 'bg-brand-primary text-white hover:bg-brand-primary/90 active:bg-brand-primary/80 hover:shadow-lg'
              : 'bg-gray-300 text-gray-500 cursor-not-allowed'
          }`}
        >
          {t('common.continue')}
        </button>
      </div>
    </div>
  );
};

export default StyleSelector;
