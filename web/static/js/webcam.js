/**
 * WebcamCapture - Reusable webcam capture module
 * Handles camera initialization, image capture, and cleanup
 * 
 * @author Absensi System
 * @version 1.0
 */

class WebcamCapture {
    constructor() {
        this.stream = null;
        this.video = null;
        this.canvas = null;
    }

    /**
     * Initialize camera and start video stream
     * @param {HTMLVideoElement} videoElement - Video element to display stream
     * @param {Object} constraints - Camera constraints (optional)
     * @returns {Promise<void>}
     */
    async initialize(videoElement, constraints = null) {
        this.video = videoElement;

        // Default constraints optimized for face recognition
        const defaultConstraints = {
            video: {
                width: { ideal: 1280 },
                height: { ideal: 720 },
                facingMode: 'user' // Front camera
            },
            audio: false
        };

        const finalConstraints = constraints || defaultConstraints;

        try {
            // Request camera access
            this.stream = await navigator.mediaDevices.getUserMedia(finalConstraints);
            
            // Attach stream to video element
            this.video.srcObject = this.stream;
            
            // Wait for video to be ready
            await new Promise((resolve) => {
                this.video.onloadedmetadata = () => {
                    this.video.play();
                    resolve();
                };
            });

            return { success: true };
        } catch (error) {
            return this._handleCameraError(error);
        }
    }

    /**
     * Capture current video frame as base64 image
     * @param {number} quality - JPEG quality (0.0 - 1.0), default 0.85
     * @returns {string} Base64 encoded image (without data:image prefix)
     */
    capture(quality = 0.85) {
        if (!this.video || !this.stream) {
            throw new Error('Camera not initialized');
        }

        // Create canvas if not exists
        if (!this.canvas) {
            this.canvas = document.createElement('canvas');
        }

        // Set canvas size to match video
        this.canvas.width = this.video.videoWidth;
        this.canvas.height = this.video.videoHeight;

        // Draw current video frame to canvas
        const ctx = this.canvas.getContext('2d');
        ctx.drawImage(this.video, 0, 0);

        // Convert to base64 JPEG
        const dataURL = this.canvas.toDataURL('image/jpeg', quality);
        
        // Remove data:image/jpeg;base64, prefix
        return dataURL.split(',')[1];
    }

    /**
     * Get full data URL (with prefix)
     * @param {number} quality - JPEG quality (0.0 - 1.0)
     * @returns {string} Full data URL
     */
    captureDataURL(quality = 0.85) {
        if (!this.canvas) {
            this.canvas = document.createElement('canvas');
        }

        this.canvas.width = this.video.videoWidth;
        this.canvas.height = this.video.videoHeight;

        const ctx = this.canvas.getContext('2d');
        ctx.drawImage(this.video, 0, 0);

        return this.canvas.toDataURL('image/jpeg', quality);
    }

    /**
     * Stop camera and release resources
     */
    stop() {
        if (this.stream) {
            this.stream.getTracks().forEach(track => track.stop());
            this.stream = null;
        }

        if (this.video) {
            this.video.srcObject = null;
        }
    }

    /**
     * Check if camera is currently active
     * @returns {boolean}
     */
    isActive() {
        return this.stream !== null && this.stream.active;
    }

    /**
     * Get video dimensions
     * @returns {Object} {width, height}
     */
    getDimensions() {
        if (!this.video) {
            return { width: 0, height: 0 };
        }
        return {
            width: this.video.videoWidth,
            height: this.video.videoHeight
        };
    }

    /**
     * Handle camera errors and return user-friendly messages
     * @private
     */
    _handleCameraError(error) {
        let message = 'Terjadi kesalahan saat mengakses kamera';
        
        if (error.name === 'NotAllowedError' || error.name === 'PermissionDeniedError') {
            message = 'Akses kamera ditolak. Silakan izinkan akses kamera di pengaturan browser.';
        } else if (error.name === 'NotFoundError' || error.name === 'DevicesNotFoundError') {
            message = 'Kamera tidak ditemukan. Pastikan kamera terpasang dengan benar.';
        } else if (error.name === 'NotReadableError' || error.name === 'TrackStartError') {
            message = 'Kamera sedang digunakan oleh aplikasi lain. Tutup aplikasi lain dan coba lagi.';
        } else if (error.name === 'OverconstrainedError') {
            message = 'Kamera tidak mendukung resolusi yang diminta.';
        } else if (error.name === 'TypeError') {
            message = 'Browser tidak mendukung akses kamera. Gunakan browser modern (Chrome, Firefox, Edge).';
        }

        return {
            success: false,
            error: error.name,
            message: message
        };
    }

    /**
     * Check if browser supports getUserMedia
     * @static
     * @returns {boolean}
     */
    static isSupported() {
        return !!(navigator.mediaDevices && navigator.mediaDevices.getUserMedia);
    }
}

/**
 * CameraModal - Reusable camera modal component
 * Provides UI for camera capture with preview
 */
class CameraModal {
    constructor(options = {}) {
        this.options = {
            title: options.title || 'Ambil Foto Wajah',
            captureButtonText: options.captureButtonText || 'Ambil Foto',
            cancelButtonText: options.cancelButtonText || 'Batal',
            tips: options.tips || [
                'Pastikan wajah terlihat jelas',
                'Pencahayaan cukup',
                'Hadap kamera langsung'
            ],
            onCapture: options.onCapture || null,
            onCancel: options.onCancel || null,
            quality: options.quality || 0.85
        };

        this.webcam = new WebcamCapture();
        this.modal = null;
        this.videoElement = null;
        this.isCapturing = false;
    }

    /**
     * Show camera modal
     */
    async show() {
        // Create modal HTML
        this._createModal();
        
        // Add to DOM
        document.body.appendChild(this.modal);
        
        // Initialize camera
        await this._initializeCamera();
    }

    /**
     * Hide and cleanup modal
     */
    hide() {
        this.webcam.stop();
        
        if (this.modal && this.modal.parentNode) {
            this.modal.parentNode.removeChild(this.modal);
        }
        
        this.modal = null;
        this.videoElement = null;
    }

    /**
     * Create modal HTML structure
     * @private
     */
    _createModal() {
        this.modal = document.createElement('div');
        this.modal.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
        
        const tipsHTML = this.options.tips.map(tip => 
            `<li class="text-sm text-gray-600">• ${tip}</li>`
        ).join('');

        this.modal.innerHTML = `
            <div class="bg-white rounded-xl shadow-2xl max-w-2xl w-full mx-4 overflow-hidden">
                <!-- Header -->
                <div class="bg-gradient-to-r from-blue-600 to-blue-700 text-white px-6 py-4 flex justify-between items-center">
                    <h3 class="text-xl font-bold">${this.options.title}</h3>
                    <button id="modal-close" class="text-white hover:text-gray-200 transition">
                        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                        </svg>
                    </button>
                </div>

                <!-- Content -->
                <div class="p-6">
                    <!-- Camera Preview -->
                    <div id="camera-container" class="mb-4">
                        <div id="camera-loading" class="bg-gray-100 rounded-lg flex items-center justify-center" style="height: 480px;">
                            <div class="text-center">
                                <div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mb-4"></div>
                                <p class="text-gray-600">Membuka kamera...</p>
                            </div>
                        </div>
                        <video id="camera-video" class="w-full rounded-lg bg-black hidden" autoplay playsinline></video>
                        <div id="camera-error" class="hidden bg-red-50 border border-red-200 rounded-lg p-4">
                            <div class="flex items-start">
                                <svg class="w-6 h-6 text-red-500 mr-3 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                                </svg>
                                <div>
                                    <p class="text-red-800 font-semibold mb-1">Gagal Membuka Kamera</p>
                                    <p id="camera-error-message" class="text-red-600 text-sm"></p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Tips -->
                    <div class="bg-blue-50 rounded-lg p-4 mb-4">
                        <p class="font-semibold text-blue-900 mb-2">Tips:</p>
                        <ul class="space-y-1">
                            ${tipsHTML}
                        </ul>
                    </div>

                    <!-- Buttons -->
                    <div class="flex gap-3">
                        <button id="capture-button" disabled class="flex-1 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed text-white font-semibold py-3 px-6 rounded-lg transition flex items-center justify-center gap-2">
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"></path>
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"></path>
                            </svg>
                            <span id="capture-button-text">${this.options.captureButtonText}</span>
                        </button>
                        <button id="cancel-button" class="px-6 py-3 border-2 border-gray-300 hover:border-gray-400 text-gray-700 font-semibold rounded-lg transition">
                            ${this.options.cancelButtonText}
                        </button>
                    </div>
                </div>
            </div>
        `;

        // Attach event listeners
        this.modal.querySelector('#modal-close').addEventListener('click', () => this._handleCancel());
        this.modal.querySelector('#cancel-button').addEventListener('click', () => this._handleCancel());
        this.modal.querySelector('#capture-button').addEventListener('click', () => this._handleCapture());
    }

    /**
     * Initialize camera
     * @private
     */
    async _initializeCamera() {
        const loadingEl = this.modal.querySelector('#camera-loading');
        const videoEl = this.modal.querySelector('#camera-video');
        const errorEl = this.modal.querySelector('#camera-error');
        const errorMessageEl = this.modal.querySelector('#camera-error-message');
        const captureButton = this.modal.querySelector('#capture-button');

        this.videoElement = videoEl;

        // Check browser support
        if (!WebcamCapture.isSupported()) {
            loadingEl.classList.add('hidden');
            errorEl.classList.remove('hidden');
            errorMessageEl.textContent = 'Browser tidak mendukung akses kamera. Gunakan browser modern (Chrome, Firefox, Edge).';
            return;
        }

        // Initialize webcam
        const result = await this.webcam.initialize(videoEl);

        if (result.success) {
            // Success - show video
            loadingEl.classList.add('hidden');
            videoEl.classList.remove('hidden');
            captureButton.disabled = false;
        } else {
            // Error - show error message
            loadingEl.classList.add('hidden');
            errorEl.classList.remove('hidden');
            errorMessageEl.textContent = result.message;
        }
    }

    /**
     * Handle capture button click
     * @private
     */
    async _handleCapture() {
        if (this.isCapturing) return;

        this.isCapturing = true;
        const captureButton = this.modal.querySelector('#capture-button');
        const captureButtonText = this.modal.querySelector('#capture-button-text');
        
        // Update button state
        captureButton.disabled = true;
        captureButtonText.textContent = 'Mengambil foto...';

        try {
            // Capture image
            const imageData = this.webcam.capture(this.options.quality);
            
            // Call callback
            if (this.options.onCapture) {
                await this.options.onCapture(imageData);
            }

            // Close modal
            this.hide();
        } catch (error) {
            console.error('Capture error:', error);
            alert('Gagal mengambil foto: ' + error.message);
            
            // Reset button
            captureButton.disabled = false;
            captureButtonText.textContent = this.options.captureButtonText;
        } finally {
            this.isCapturing = false;
        }
    }

    /**
     * Handle cancel button click
     * @private
     */
    _handleCancel() {
        if (this.options.onCancel) {
            this.options.onCancel();
        }
        this.hide();
    }
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { WebcamCapture, CameraModal };
}
