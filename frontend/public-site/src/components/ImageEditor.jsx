import React, { useState, useRef, useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import ReactCrop from 'react-image-crop';
import { ZoomIn, ZoomOut, RotateCw, Upload, Crop } from 'lucide-react';
import useCouponStore from '../store/couponStore';
import { useUIStore } from '../store/partnerStore';
import 'react-image-crop/dist/ReactCrop.css';

const getCroppedImg = (
  image,
  pixelCrop,
  rotation = 0,
  scale = 1,
  position = { x: 0, y: 0 }
) => {
  const canvas = document.createElement('canvas');
  const ctx = canvas.getContext('2d');

  if (!ctx || !pixelCrop) {
    return null;
  }

  const naturalWidth = image.naturalWidth;
  const naturalHeight = image.naturalHeight;

  canvas.width = pixelCrop.width;
  canvas.height = pixelCrop.height;

  ctx.save();
  ctx.translate(canvas.width / 2, canvas.height / 2);
  ctx.scale(1 / scale, 1 / scale);
  ctx.rotate((-rotation * Math.PI) / 180);
  ctx.translate(-canvas.width / 2, -canvas.height / 2);

  const offsetX = pixelCrop.x - position.x / scale;
  const offsetY = pixelCrop.y - position.y / scale;

  ctx.drawImage(
    image,
    offsetX,
    offsetY,
    pixelCrop.width / scale,
    pixelCrop.height / scale,
    0,
    0,
    pixelCrop.width,
    pixelCrop.height
  );

  ctx.restore();

  return canvas.toDataURL('image/jpeg');
};

const ImageEditor = ({
  imageUrl,
  onSave,
  onCancel,
  title,
  showCropHint = true,
  aspectRatio,
  fileName: propFileName,
}) => {
  const { t } = useTranslation();
  const imgRef = useRef(null);
  const fileInputRef = useRef(null);
  const containerRef = useRef(null);
  const couponStore = useCouponStore();
  
  const [crop, setCrop] = useState(null);
  const [completedCrop, setCompletedCrop] = useState(null);
  const [scale, setScale] = useState(1);
  const [rotate, setRotate] = useState(0);
  const [fileName, setFileName] = useState('');
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [lastMousePos, setLastMousePos] = useState({ x: 0, y: 0 });
  const [isCropping, setIsCropping] = useState(false);

  useEffect(() => {
    if (propFileName) {
      setFileName(propFileName);
      return;
    }
    if (imageUrl) {
      if (imageUrl.startsWith('data:')) {
        if (couponStore.selectedFile?.name) {
          setFileName(couponStore.selectedFile.name);
        } else {
          setFileName('uploaded_image.jpg');
        }
      } else {
        const parts = imageUrl.split('/');
        const name = parts[parts.length - 1];
        setFileName(name);
      }
    }
  }, [imageUrl, propFileName, couponStore.selectedFile]);

  useEffect(() => {
    if (couponStore.editorParams) {
      setCrop(couponStore.editorParams.crop || null);
      setScale(couponStore.editorParams.scale || 1);
      setRotate(couponStore.editorParams.rotate || 0);
      setPosition(couponStore.editorParams.position || { x: 0, y: 0 });
    } else {
      setCrop(null);
      setCompletedCrop(null);
      setScale(1);
      setRotate(0);
      setPosition({ x: 0, y: 0 });
      setIsCropping(false);
    }
  }, [imageUrl, couponStore.editorParams]);

  function onImageLoad(e) {
    imgRef.current = e.currentTarget;
    setCrop(null);
    setCompletedCrop(null);
  }

  const compressImage = (file, maxWidth = 1920, quality = 0.8) => {
    return new Promise(resolve => {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      const img = new Image();

      img.onload = () => {
        let { width, height } = img;
        if (width > maxWidth) {
          height = (height * maxWidth) / width;
          width = maxWidth;
        }

        canvas.width = width;
        canvas.height = height;

        ctx.drawImage(img, 0, 0, width, height);

        const compressedDataUrl = canvas.toDataURL('image/jpeg', quality);
        resolve(compressedDataUrl);
      };

      img.src = URL.createObjectURL(file);
    });
  };

  const [isProcessing, setIsProcessing] = useState(false);

  const handleFileUpload = async event => {
    if (isProcessing) {
      return;
    }

    const file = event.target.files[0];

    if (file && file.type.startsWith('image/')) {
      setIsProcessing(true);
      try {
        const compressedImageUrl = await compressImage(file);
        
        sessionStorage.setItem('diamondMosaic_fileUrl', compressedImageUrl);
        setFileName(file.name);
        
        couponStore.setSelectedFile(file);
        couponStore.setPreviewUrl(compressedImageUrl);
        
        if (onSave) {
          onSave(compressedImageUrl, { newImage: true, fileName: file.name });
        }
      } catch (error) {
        alert(t('image_editor.processing_error'));
      } finally {
        setIsProcessing(false);
      }
    } else {
      setIsProcessing(false);
    }

    event.target.value = '';
  };

  const handleSave = () => {
    if (completedCrop && imgRef.current) {
      const croppedImageUrl = getCroppedImg(
        imgRef.current,
        completedCrop,
        rotate,
        scale,
        position
      );
      if (croppedImageUrl) {
        const editorParams = { crop, rotate, scale, position };
        couponStore.setEditorParams(editorParams);
        couponStore.setEditedImageUrl(croppedImageUrl);
        onSave(croppedImageUrl, editorParams);
      }
    } else if (imgRef.current) {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      canvas.width = imgRef.current.naturalWidth;
      canvas.height = imgRef.current.naturalHeight;
      
      ctx.save();
      ctx.translate(canvas.width / 2, canvas.height / 2);
      ctx.rotate((rotate * Math.PI) / 180);
      ctx.scale(scale, scale);
      ctx.translate(-canvas.width / 2, -canvas.height / 2);
      ctx.drawImage(imgRef.current, 0, 0);
      ctx.restore();
      
      const editedUrl = canvas.toDataURL('image/jpeg');
      const editorParams = { crop: null, rotate, scale, position };
      couponStore.setEditorParams(editorParams);
      couponStore.setEditedImageUrl(editedUrl);
      onSave(editedUrl, editorParams);
    }
  };

  const getImageStyles = () => {
    return {
      transform: `translate(${position.x}px, ${position.y}px) rotate(${rotate}deg) scale(${scale})`,
      width: '100%',
      height: '100%',
      objectFit: 'cover',
      transformOrigin: 'center center',
      display: 'block',
      cursor: isDragging ? 'grabbing' : isCropping ? 'crosshair' : 'grab',
      userSelect: 'none',
    };
  };

  const handleRotate = () => {
    setRotate(prevRotate => (prevRotate + 90) % 360);
    setPosition({ x: 0, y: 0 });
    setCrop(null);
    setCompletedCrop(null);
  };

  const handleZoomIn = () => {
    setScale(Math.min(scale + 0.1, 3));
  };

  const handleZoomOut = () => {
    setScale(Math.max(scale - 0.1, 0.1));
  };

  const handleReset = () => {
    setRotate(0);
    setScale(1);
    setPosition({ x: 0, y: 0 });
    setCrop(null);
    setCompletedCrop(null);
    setIsCropping(false);
  };

  const handleWheel = useCallback(event => {
    if (imgRef.current) {
      event.preventDefault();
      event.stopPropagation();
      const delta = event.deltaY > 0 ? -0.1 : 0.1;
      setScale(prev => Math.max(0.1, Math.min(3, prev + delta)));
    }
  });

  const handleMouseDown = e => {
    if (e.button === 2) {
      e.preventDefault();
      setIsDragging(true);
      setLastMousePos({ x: e.clientX, y: e.clientY });
    }
  };

  const handleMouseMove = e => {
    if (isDragging) {
      e.preventDefault();
      const deltaX = e.clientX - lastMousePos.x;
      const deltaY = e.clientY - lastMousePos.y;
      setPosition(prev => ({
        x: prev.x + deltaX,
        y: prev.y + deltaY,
      }));
      setLastMousePos({ x: e.clientX, y: e.clientY });
    }
  };

  const handleMouseUp = e => {
    setIsDragging(false);
  };

  const handleContextMenu = e => {
    e.preventDefault();
  };

  useEffect(() => {
    const container = containerRef.current;
    if (container) {
      container.addEventListener('wheel', handleWheel, { passive: false });
      return () => {
        container.removeEventListener('wheel', handleWheel);
      };
    }
  });

  useEffect(() => {
    const handleGlobalMouseMove = e => {
      if (isDragging) {
        handleMouseMove(e);
      }
    };
    const handleGlobalMouseUp = e => {
      if (isDragging) {
        handleMouseUp(e);
      }
    };
    if (isDragging) {
      document.addEventListener('mousemove', handleGlobalMouseMove);
      document.addEventListener('mouseup', handleGlobalMouseUp);
      document.addEventListener('contextmenu', handleContextMenu);
      document.body.style.cursor = 'grabbing';
      document.body.style.overflow = 'hidden';
      return () => {
        document.removeEventListener('mousemove', handleGlobalMouseMove);
        document.removeEventListener('mouseup', handleGlobalMouseUp);
        document.removeEventListener('contextmenu', handleContextMenu);
        document.body.style.cursor = 'default';
        document.body.style.overflow = 'auto';
      };
    }
  });

  return (
    <div className="bg-white rounded-xl sm:rounded-2xl shadow-lg">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between p-4 sm:p-6 border-b border-gray-200 gap-3 sm:gap-0">
        <h2 className="text-lg sm:text-xl font-semibold text-gray-900">
          {title}
        </h2>
        <button
          onClick={() => fileInputRef.current?.click()}
          disabled={isProcessing}
          className={`flex items-center gap-2 px-4 py-2 sm:px-5 sm:py-2 rounded-lg transition-colors text-sm sm:text-base w-full sm:w-auto justify-center sm:justify-start ${
            isProcessing
              ? 'bg-gray-400 text-gray-600 cursor-not-allowed'
              : 'bg-blue-500 text-white hover:bg-blue-600'
          }`}
        >
          {isProcessing ? (
            <>
              <div className="w-4 h-4 sm:w-5 sm:h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
              <span>{t('common.processing') || 'Processing...'}</span>
            </>
          ) : (
            <>
              <Upload size={16} className="sm:w-5 sm:h-5" />
              <span>{t('diamond_mosaic_page.image_editor.tools.replace')}</span>
            </>
          )}
        </button>
      </div>

      <div className="p-3 sm:p-4 lg:p-6">
        <div
          ref={containerRef}
          className="overflow-hidden mb-4 sm:mb-6 relative flex items-center justify-center mx-auto transition-all duration-300"
          style={{
            width: '100%',
            height: rotate === 90 || rotate === 270 ? '500px' : '500px',
            maxWidth: '800px',
            maxHeight: rotate === 90 || rotate === 270 ? '600px' : '70vh',
            minHeight: '400px',
          }}
        >
          <div
            className="relative"
            style={{
              width: '100%',
              height: '100%',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              overflow: 'hidden',
            }}
            onMouseDown={handleMouseDown}
            onMouseUp={handleMouseUp}
            onContextMenu={handleContextMenu}
          >
            <ReactCrop
              crop={isCropping ? crop : null}
              onChange={c => isCropping && !isDragging && setCrop(c)}
              onComplete={c => isCropping && !isDragging && setCompletedCrop(c)}
              aspect={aspectRatio}
              disabled={isDragging || !isCropping}
              style={{
                maxWidth: '100%',
                maxHeight: '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                pointerEvents: isDragging ? 'none' : 'auto',
              }}
            >
              <img
                ref={imgRef}
                src={imageUrl}
                alt="Edit"
                style={getImageStyles()}
                onLoad={onImageLoad}
                draggable={false}
              />
            </ReactCrop>
          </div>
        </div>

        <div className="flex flex-col sm:flex-row flex-wrap items-center justify-center gap-2 sm:gap-3 mb-4 sm:mb-6 p-3 sm:p-4 bg-gray-50 rounded-xl">
          <button
            onClick={() => setIsCropping(!isCropping)}
            className={`flex items-center px-4 py-3 sm:py-2 rounded-xl font-medium transition-all duration-200 shadow-sm hover:shadow-md text-sm sm:text-base w-full sm:w-auto justify-center ${
              isCropping
                ? 'bg-emerald-500 hover:bg-emerald-600 text-white'
                : 'bg-white hover:bg-gray-50 text-gray-700 border border-gray-200'
            }`}
            title={t('diamond_mosaic_page.image_editor.tooltips.toggle_crop')}
          >
            <Crop className="w-4 h-4 mr-2 flex-shrink-0" />
            <span>{t('diamond_mosaic_page.image_editor.tools.crop')}</span>
          </button>

          <button
            onClick={handleRotate}
            className="flex items-center px-4 py-3 sm:py-2 bg-white hover:bg-gray-50 text-gray-700 rounded-xl font-medium transition-all duration-200 shadow-sm hover:shadow-md border border-gray-200 text-sm sm:text-base w-full sm:w-auto justify-center"
            title={t('diamond_mosaic_page.image_editor.tooltips.rotate_90')}
          >
            <RotateCw className="w-4 h-4 mr-2 flex-shrink-0" />
            <span>{t('diamond_mosaic_page.image_editor.tools.rotate')}</span>
          </button>

          <div className="flex items-center gap-1 sm:gap-2 w-full sm:w-auto justify-center">
            <button
              onClick={handleZoomOut}
              className="w-10 h-10 sm:w-12 sm:h-10 bg-white hover:bg-gray-50 text-gray-700 rounded-xl flex items-center justify-center transition-all duration-200 shadow-sm hover:shadow-md border border-gray-200"
              title={t('diamond_mosaic_page.image_editor.tooltips.zoom_out')}
            >
              <ZoomOut className="w-4 h-4 sm:w-5 sm:h-5" />
            </button>
            <span className="px-3 py-2 bg-white border border-gray-200 rounded-xl text-xs sm:text-sm font-medium text-gray-700 min-w-14 sm:min-w-16 text-center">
              {Math.round(scale * 100)}%
            </span>
            <button
              onClick={handleZoomIn}
              className="w-10 h-10 sm:w-12 sm:h-10 bg-white hover:bg-gray-50 text-gray-700 rounded-xl flex items-center justify-center transition-all duration-200 shadow-sm hover:shadow-md border border-gray-200"
              title={t('diamond_mosaic_page.image_editor.tooltips.zoom_in')}
            >
              <ZoomIn className="w-4 h-4 sm:w-5 sm:h-5" />
            </button>
          </div>

          <button
            onClick={handleReset}
            className="flex items-center px-4 py-3 sm:py-2 bg-orange-500 hover:bg-orange-600 text-white rounded-xl font-medium transition-all duration-200 shadow-sm hover:shadow-md text-sm sm:text-base w-full sm:w-auto justify-center"
            title={t('diamond_mosaic_page.image_editor.tooltips.reset_all')}
          >
            <span>{t('diamond_mosaic_page.image_editor.tools.reset')}</span>
          </button>
        </div>

        <div className="text-center bg-white border border-gray-200 rounded-lg p-4 sm:p-6 mb-4 sm:mb-6">
          {fileName && (
            <div className="flex items-center justify-center mb-2 sm:mb-3 flex-wrap">
              <div className="w-2 h-2 bg-green-500 rounded-full mr-2 flex-shrink-0"></div>
              <p className="text-sm sm:text-base font-medium text-gray-800 break-all text-center">
                {fileName}
              </p>
            </div>
          )}
          {isCropping ? (
            <p className="text-sm sm:text-base text-gray-700 mb-2 leading-relaxed">
              {t(
                'diamond_mosaic_page.image_editor.instructions.cropping_active'
              )}
            </p>
          ) : (
            <p className="text-sm sm:text-base text-gray-700 mb-2 leading-relaxed">
              {t('diamond_mosaic_page.image_editor.instructions.general')}
            </p>
          )}
          <p className="text-xs sm:text-sm text-gray-500 leading-relaxed">
            {t('diamond_mosaic_page.image_editor.instructions.controls')}
          </p>
        </div>

        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          onChange={handleFileUpload}
          className="hidden"
        />
      </div>
    </div>
  );
};

export default ImageEditor;
