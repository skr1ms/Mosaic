class NavigationService {
  constructor() {
    this.state = this.loadState() || this.getInitialState();
    this.listeners = [];
  }

  getInitialState() {
    return {
      currentStep: 'size-selection',
      selectedSize: null,
      selectedFile: null,
      selectedStyle: null,
      selectedPreview: null,
      imageData: {
        original: null,
        edited: null,
        previewUrl: null,
      },
      editorSettings: {
        size: null,
        style: null,
        edits: [],
      },
      previews: [],
      aiPreviews: [],
      useAI: false,
      purchaseData: null,
      history: [],
    };
  }

  loadState() {
    try {
      const savedState = localStorage.getItem('mosaicNavigationState');
      return savedState ? JSON.parse(savedState) : null;
    } catch (error) {
      console.error('Error loading navigation state:', error);
      return null;
    }
  }

  saveState() {
    try {
      localStorage.setItem('mosaicNavigationState', JSON.stringify(this.state));
    } catch (error) {
      console.error('Error saving navigation state:', error);
    }
  }

  updateState(updates) {
    this.state = {
      ...this.state,
      ...updates,
      lastUpdated: Date.now(),
    };
    this.saveState();
    this.notifyListeners();
  }

  goToStep(step) {
    const validSteps = [
      'size-selection',
      'style-selection',
      'preview-album',
      'purchase',
      'success',
    ];
    if (!validSteps.includes(step)) {
      console.error(`Invalid step: ${step}`);
      return false;
    }

    this.state.history.push({
      step: this.state.currentStep,
      timestamp: Date.now(),
    });

    this.updateState({ currentStep: step });
    return true;
  }

  goBack() {
    if (this.state.history.length === 0) return false;

    const previousStep = this.state.history.pop();
    this.updateState({
      currentStep: previousStep.step,
      history: this.state.history,
    });
    return true;
  }

  selectSize(size) {
    this.updateState({
      selectedSize: size,
      editorSettings: {
        ...this.state.editorSettings,
        size,
      },
    });
  }

  uploadImage(file, previewUrl) {
    this.updateState({
      selectedFile: file,
      imageData: {
        ...this.state.imageData,
        original: previewUrl,
        previewUrl: previewUrl,
      },
    });
  }

  updateEditedImage(editedUrl, editorParams) {
    this.updateState({
      imageData: {
        ...this.state.imageData,
        edited: editedUrl,
        previewUrl: editedUrl,
      },
      editorSettings: {
        ...this.state.editorSettings,
        edits: editorParams,
      },
    });
  }

  selectStyle(style) {
    this.updateState({
      selectedStyle: style,
      editorSettings: {
        ...this.state.editorSettings,
        style,
      },
    });
  }

  setPreviews(previews) {
    this.updateState({ previews });
  }

  addAIPreviews(aiPreviews) {
    this.updateState({
      aiPreviews,
      previews: [...this.state.previews, ...aiPreviews],
    });
  }

  selectPreview(previewIndex) {
    this.updateState({ selectedPreview: previewIndex });
  }

  toggleAI(enabled) {
    this.updateState({ useAI: enabled });
  }

  setPurchaseData(data) {
    this.updateState({ purchaseData: data });
  }

  reset() {
    this.state = this.getInitialState();
    this.saveState();
    this.notifyListeners();

    try {
      localStorage.removeItem('diamondMosaic_selectedImage');
      localStorage.removeItem('diamondMosaic_editorSettings');
      localStorage.removeItem('diamondMosaic_purchaseData');
      sessionStorage.removeItem('diamondMosaic_fileUrl');
    } catch (error) {
      console.error('Error clearing storage:', error);
    }
  }

  canProceedToStyles() {
    return this.state.selectedSize && this.state.selectedFile;
  }

  canProceedToAlbum() {
    return this.canProceedToStyles() && this.state.selectedStyle;
  }

  canProceedToPurchase() {
    return this.canProceedToAlbum() && this.state.selectedPreview !== null;
  }

  subscribe(listener) {
    this.listeners.push(listener);
    return () => {
      this.listeners = this.listeners.filter(l => l !== listener);
    };
  }

  notifyListeners() {
    this.listeners.forEach(listener => listener(this.state));
  }

  getState() {
    return { ...this.state };
  }
}

const navigationService = new NavigationService();

export default navigationService;
