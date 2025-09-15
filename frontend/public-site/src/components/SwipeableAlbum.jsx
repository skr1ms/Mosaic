import React, { useState, useRef, useEffect } from 'react';
import { motion } from 'framer-motion';
import { ChevronLeft, ChevronRight, Loader2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';

const SwipeableAlbum = ({
  previews,
  selectedPreview,
  onPreviewSelect,
  isGeneratingVariants,
}) => {
  const { t } = useTranslation();
  const [currentPage, setCurrentPage] = useState(0);
  const scrollContainerRef = useRef(null);

  const [itemsPerPage, setItemsPerPage] = useState(4);
  const totalPages = Math.ceil(previews.length / itemsPerPage);

  useEffect(() => {
    const updateItemsPerPage = () => {
      setItemsPerPage(4);
    };

    updateItemsPerPage();
    window.addEventListener('resize', updateItemsPerPage);

    return () => window.removeEventListener('resize', updateItemsPerPage);
  }, []);

  useEffect(() => {
    if (selectedPreview >= 0) {
      const targetPage = Math.floor(selectedPreview / itemsPerPage);
      setCurrentPage(targetPage);
    }
  }, [selectedPreview, itemsPerPage]);

  const goToPage = pageIndex => {
    let newPage = pageIndex;
    if (pageIndex < 0) {
      newPage = totalPages - 1;
    } else if (pageIndex >= totalPages) {
      newPage = 0;
    }
    setCurrentPage(newPage);
  };

  const goNext = () => {
    goToPage(currentPage + 1);
  };

  const goPrev = () => {
    goToPage(currentPage - 1);
  };

  const getCurrentPagePreviews = () => {
    const startIndex = currentPage * itemsPerPage;
    const endIndex = startIndex + itemsPerPage;
    return previews.slice(startIndex, endIndex);
  };

  if (isGeneratingVariants && previews.length === 0) {
    return (
      <div className="w-full max-w-4xl mb-6 sm:mb-8">
        <div className="text-center py-8">
          <Loader2 className="w-8 h-8 animate-spin text-purple-600 mx-auto mb-2" />
          <p className="text-gray-600">
            {t('diamond_mosaic_preview_album.generating_variants')}
          </p>
        </div>
      </div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.2 }}
      className="w-full max-w-4xl mb-6 sm:mb-8"
    >
      {}
      <h2 className="text-xl sm:text-2xl font-bold text-gray-800 mb-4 sm:mb-6 text-center px-4">
        {t('diamond_mosaic_preview_album.style_variants_title')}
      </h2>

      {}
      <div className="flex items-center justify-center w-full max-w-5xl mx-auto">
        {}
        {totalPages > 1 && (
          <button
            onClick={goPrev}
            className="flex-shrink-0 w-10 h-10 sm:w-12 sm:h-12 bg-white/95 backdrop-blur-sm hover:bg-white shadow-lg rounded-full flex items-center justify-center transition-all duration-300 touch-target group border border-gray-200 hover:border-purple-300 mr-2 sm:mr-3"
            aria-label="Предыдущая страница превьюшек"
          >
            <ChevronLeft className="w-5 h-5 sm:w-6 sm:h-6 text-purple-600 group-hover:text-purple-700" />
          </button>
        )}

        {}
        <div
          ref={scrollContainerRef}
          className="overflow-hidden max-w-2xl flex justify-center"
        >
          <motion.div
            className="flex gap-3 sm:gap-4 transition-transform duration-300 ease-in-out justify-center items-center"
            animate={{ x: 0 }}
            key={currentPage}
            drag="x"
            dragConstraints={{ left: 0, right: 0 }}
            onDragEnd={(event, info) => {
              const threshold = 50;
              if (info.offset.x > threshold) {
                goPrev();
              } else if (info.offset.x < -threshold) {
                goNext();
              }
            }}
          >
            {getCurrentPagePreviews().map((preview, pageIndex) => {
              const actualIndex = currentPage * itemsPerPage + pageIndex;
              return (
                <motion.div
                  key={preview.id}
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ delay: pageIndex * 0.1 }}
                  className={`flex-shrink-0 flex flex-col items-center p-2 sm:p-3 rounded-lg cursor-pointer transition-all touch-target ${
                    selectedPreview === actualIndex
                      ? 'bg-purple-100 border-2 border-purple-500 shadow-md'
                      : 'bg-white border border-gray-200 hover:border-purple-300 hover:shadow-sm active:bg-gray-50'
                  }`}
                  onClick={() => onPreviewSelect(actualIndex)}
                  style={{
                    width: `${95 / Math.min(itemsPerPage, getCurrentPagePreviews().length)}%`,
                    maxWidth: '140px',
                    minWidth: '100px',
                  }}
                >
                  {}
                  <div className="w-20 h-20 sm:w-24 sm:h-24 md:w-28 md:h-28 rounded-lg overflow-hidden bg-gray-100 mb-2 flex-shrink-0 mx-auto">
                    {preview.url ? (
                      <img
                        src={preview.url}
                        alt={preview.title}
                        className="w-full h-full object-cover"
                      />
                    ) : preview.error ? (
                      <div className="w-full h-full flex items-center justify-center text-red-400">
                        ❌
                      </div>
                    ) : (
                      <div className="w-full h-full flex items-center justify-center">
                        <Loader2 className="w-4 h-4 animate-spin text-purple-600" />
                      </div>
                    )}
                  </div>

                  {}
                  <div className="text-center w-full">
                    <p className="font-medium text-gray-800 text-xs sm:text-sm truncate w-full leading-tight">
                      {preview.title}
                    </p>
                    <p className="text-xs text-gray-500 capitalize leading-tight">
                      {preview.type}
                    </p>
                    {preview.isMain && (
                      <span className="inline-block px-2 py-1 text-xs bg-purple-100 text-purple-700 rounded mt-1 font-medium">
                        {t('diamond_mosaic_preview_album.main_preview')}
                      </span>
                    )}
                  </div>
                </motion.div>
              );
            })}
          </motion.div>
        </div>

        {}
        {totalPages > 1 && (
          <button
            onClick={goNext}
            className="flex-shrink-0 w-10 h-10 sm:w-12 sm:h-12 bg-white/95 backdrop-blur-sm hover:bg-white shadow-lg rounded-full flex items-center justify-center transition-all duration-300 touch-target group border border-gray-200 hover:border-purple-300 ml-2 sm:ml-3"
            aria-label="Следующая страница превьюшек"
          >
            <ChevronRight className="w-5 h-5 sm:w-6 sm:h-6 text-purple-600 group-hover:text-purple-700" />
          </button>
        )}
      </div>

      {}
      {totalPages > 1 && (
        <div className="flex justify-center mt-4 gap-2">
          {Array.from({ length: totalPages }, (_, index) => (
            <button
              key={index}
              onClick={() => setCurrentPage(index)}
              className={`w-2 h-2 rounded-full transition-all duration-300 ${
                currentPage === index
                  ? 'bg-purple-600 w-6'
                  : 'bg-gray-300 hover:bg-purple-400'
              }`}
            />
          ))}
        </div>
      )}
    </motion.div>
  );
};

export default SwipeableAlbum;
