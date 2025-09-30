import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

const useCouponStore = create(
  persist(
    (set, get) => ({
      couponCode: '',
      couponData: null,
      activationStep: 'input',
      selectedFile: null,
      previewUrl: null,
      editedImageUrl: null,
      imageId: null,
      selectedSize: '30x40',
      selectedStyle: null,
      selectedPreview: null,
      previews: [],
      editorParams: null,
      
      setCouponCode: (code) => set({ couponCode: code }),
      
      setCouponData: (data) => set({ couponData: data }),
      
      setActivationStep: (step) => set({ activationStep: step }),
      
      setSelectedFile: (file) => set({ selectedFile: file }),
      
      setPreviewUrl: (url) => set({ previewUrl: url }),
      
      setEditedImageUrl: (url) => set({ editedImageUrl: url }),
      
      setImageId: (id) => set({ imageId: id }),
      
      setSelectedSize: (size) => set({ selectedSize: size }),
      
      setSelectedStyle: (style) => set({ selectedStyle: style }),
      
      setSelectedPreview: (preview) => set({ selectedPreview: preview }),
      
      setPreviews: (previews) => set({ previews: previews }),
      
      setEditorParams: (params) => set({ editorParams: params }),
      
      updateCouponData: (updates) => set((state) => ({
        couponData: { ...state.couponData, ...updates }
      })),
      
      saveImageData: (data) => set({
        selectedFile: data.file,
        previewUrl: data.previewUrl,
        editedImageUrl: data.editedImageUrl,
        imageId: data.imageId,
        editorParams: data.editorParams
      }),
      
      clearSession: () => set({
        couponCode: '',
        couponData: null,
        activationStep: 'input',
        selectedFile: null,
        previewUrl: null,
        editedImageUrl: null,
        imageId: null,
        selectedStyle: null,
        selectedPreview: null,
        previews: [],
        editorParams: null
      }),
      
      resetToUpload: () => set({
        activationStep: 'upload',
        selectedFile: null,
        previewUrl: null,
        editedImageUrl: null,
        imageId: null,
        previews: [],
        editorParams: null
      })
    }),
    {
      name: 'coupon-activation-storage',
      storage: createJSONStorage(() => sessionStorage),
      partialize: (state) => ({
        couponCode: state.couponCode,
        couponData: state.couponData,
        activationStep: state.activationStep,
        previewUrl: state.previewUrl,
        editedImageUrl: state.editedImageUrl,
        imageId: state.imageId,
        selectedSize: state.selectedSize,
        selectedStyle: state.selectedStyle,
        selectedPreview: state.selectedPreview,
        previews: state.previews,
        editorParams: state.editorParams
      })
    }
  )
);

export default useCouponStore;
