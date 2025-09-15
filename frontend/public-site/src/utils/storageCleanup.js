export const cleanupImageStorage = () => {
  try {
    const keys = Object.keys(localStorage);
    keys.forEach(key => {
      if (
        (key.startsWith('preview_') ||
          key.startsWith('style_') ||
          key.startsWith('temp_') ||
          key.startsWith('mosaic_cache_') ||
          key.startsWith('old_diamondMosaic_') ||
          key.startsWith('diamondMosaic_') ||
          key.startsWith('coupon')) &&
        key !== 'diamondMosaic_projectSettings'
      ) {
        localStorage.removeItem(key);
      }
    });

    const sessionKeys = Object.keys(sessionStorage);
    sessionKeys.forEach(key => {
      if (key !== 'diamondMosaic_fileUrl') {
        sessionStorage.removeItem(key);
      }
    });

    console.log('Aggressive storage cleanup completed');
  } catch (error) {
    console.warn('Error during storage cleanup:', error);
  }
};

export const getStorageInfo = () => {
  try {
    let totalSize = 0;
    let itemCount = 0;

    for (let key in localStorage) {
      if (localStorage.hasOwnProperty(key)) {
        const item = localStorage.getItem(key);
        totalSize += key.length + (item ? item.length : 0);
        itemCount++;
      }
    }

    return {
      totalSize,
      itemCount,
      estimatedSizeMB: (totalSize / (1024 * 1024)).toFixed(2),
    };
  } catch (error) {
    return { error: error.message };
  }
};

export const emergencyCleanup = () => {
  try {
    const projectSettings = localStorage.getItem(
      'diamondMosaic_projectSettings'
    );

    localStorage.clear();

    if (projectSettings) {
      localStorage.setItem('diamondMosaic_projectSettings', projectSettings);
    }

    sessionStorage.clear();

    console.log(
      'Emergency cleanup completed - removed all data except project settings'
    );
  } catch (error) {
    console.error('Emergency cleanup failed:', error);
  }
};

export const compressImageData = imageData => {
  try {
    if (typeof imageData === 'string' && imageData.startsWith('data:image/')) {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      const img = new Image();

      return new Promise(resolve => {
        img.onload = () => {
          const maxSize = 800;
          let { width, height } = img;

          if (width > maxSize || height > maxSize) {
            const ratio = Math.min(maxSize / width, maxSize / height);
            width *= ratio;
            height *= ratio;
          }

          canvas.width = width;
          canvas.height = height;

          ctx.drawImage(img, 0, 0, width, height);
          const compressedData = canvas.toDataURL('image/jpeg', 0.7);

          resolve(compressedData);
        };

        img.onerror = () => resolve(imageData);
        img.src = imageData;
      });
    }

    return Promise.resolve(imageData);
  } catch (error) {
    console.warn('Error compressing image:', error);
    return Promise.resolve(imageData);
  }
};
