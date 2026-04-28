/**
 * Guided Face Scan Module
 * Comprehensive face enrollment with 5-step guided scan
 * 
 * Usage:
 *   const scanner = new GuidedFaceScan();
 *   scanner.start();
 */

class GuidedFaceScan {
    constructor(options = {}) {
        this.options = {
            onComplete: options.onComplete || null,
            onCancel: options.onCancel || null,
            onError: options.onError || null,
            ...options
        };
        
        // Scan state
        this.SCAN_STEPS = [
            { 
                name: 'frontal', 
                icon: '⏺️',
                instruction: 'Hadap kamera langsung', 
                subtext: 'Posisikan wajah di tengah dan lihat ke kamera',
                check: (pose) => Math.abs(pose.yaw) < 10 && Math.abs(pose.pitch) < 10 && Math.abs(pose.roll) < 10
            },
            { 
                name: 'left', 
                icon: '⬅️',
                instruction: 'Putar kepala ke kiri', 
                subtext: 'Putar kepala Anda ke arah kiri',
                check: (pose) => pose.yaw < -20
            },
            { 
                name: 'right', 
                icon: '➡️',
                instruction: 'Putar kepala ke kanan', 
                subtext: 'Putar kepala Anda ke arah kanan',
                check: (pose) => pose.yaw > 20
            },
            { 
                name: 'up', 
                icon: '⬆️',
                instruction: 'Angkat kepala sedikit', 
                subtext: 'Angkat dagu Anda ke atas',
                check: (pose) => pose.pitch < -15
            },
            { 
                name: 'down', 
                icon: '⬇️',
                instruction: 'Turunkan kepala sedikit', 
                subtext: 'Turunkan dagu Anda ke bawah',
                check: (pose) => pose.pitch > 15
            }
        ];
        
        this.currentStepIndex = 0;
        this.capturedPhotos = [];
        this.poseStableTime = 0;
        this.lastPoseCorrect = false;
        this.STABILITY_DURATION = 1000; // 1 second
        
        this.faceMesh = null;
        this.camera = null;
        this.isRunning = false;
        this.isScanning = false;
        
        this.fps = 0;
        this.lastTime = Date.now();
        this.frameCount = 0;
    }
    
    async start() {
        try {
            // Create modal
            this.createModal();
            
            // Initialize MediaPipe
            await this.initMediaPipe();
            
            // Start camera
            await this.startCamera();
            
            // Start scanning
            this.startScanning();
            
        } catch (error) {
            console.error('Failed to start guided scan:', error);
            if (this.options.onError) {
                this.options.onError(error);
            }
            this.cleanup();
        }
    }
    
    createModal() {
        // Create modal HTML
        const modalHTML = `
            <div id="guided-scan-modal" class="fixed inset-0 bg-black bg-opacity-75 z-50 flex items-center justify-center p-4">
                <div class="bg-white rounded-xl shadow-2xl w-full max-w-4xl max-h-[95vh] overflow-y-auto">
                    <!-- Header -->
                    <div class="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex justify-between items-center z-10">
                        <h2 class="text-2xl font-bold text-gray-800">📸 Daftar Wajah - Scan Lengkap</h2>
                        <button onclick="window.guidedScanInstance.cancel()" class="text-gray-500 hover:text-gray-700 transition">
                            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                            </svg>
                        </button>
                    </div>
                    
                    <!-- Content -->
                    <div class="p-6">
                        <!-- Progress -->
                        <div class="mb-6">
                            <div class="flex justify-between items-center mb-2">
                                <span class="text-sm font-semibold text-gray-700">Progress</span>
                                <span id="scan-progress-text" class="text-sm font-semibold text-blue-600">0/5</span>
                            </div>
                            <div class="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
                                <div id="scan-progress-bar" class="h-full bg-blue-600 transition-all duration-300" style="width: 0%"></div>
                            </div>
                        </div>
                        
                        <!-- Photo Slots -->
                        <div class="grid grid-cols-5 gap-3 mb-6">
                            <div id="photo-slot-frontal" class="photo-slot aspect-square bg-gray-100 border-2 border-dashed border-gray-300 rounded-lg flex flex-col items-center justify-center">
                                <div class="text-3xl mb-1">⏺️</div>
                                <div class="text-xs text-gray-600 text-center">Frontal</div>
                            </div>
                            <div id="photo-slot-left" class="photo-slot aspect-square bg-gray-100 border-2 border-dashed border-gray-300 rounded-lg flex flex-col items-center justify-center">
                                <div class="text-3xl mb-1">⬅️</div>
                                <div class="text-xs text-gray-600 text-center">Kiri</div>
                            </div>
                            <div id="photo-slot-right" class="photo-slot aspect-square bg-gray-100 border-2 border-dashed border-gray-300 rounded-lg flex flex-col items-center justify-center">
                                <div class="text-3xl mb-1">➡️</div>
                                <div class="text-xs text-gray-600 text-center">Kanan</div>
                            </div>
                            <div id="photo-slot-up" class="photo-slot aspect-square bg-gray-100 border-2 border-dashed border-gray-300 rounded-lg flex flex-col items-center justify-center">
                                <div class="text-3xl mb-1">⬆️</div>
                                <div class="text-xs text-gray-600 text-center">Atas</div>
                            </div>
                            <div id="photo-slot-down" class="photo-slot aspect-square bg-gray-100 border-2 border-dashed border-gray-300 rounded-lg flex flex-col items-center justify-center">
                                <div class="text-3xl mb-1">⬇️</div>
                                <div class="text-xs text-gray-600 text-center">Bawah</div>
                            </div>
                        </div>
                        
                        <!-- Video Container -->
                        <div class="relative bg-black rounded-lg overflow-hidden mb-4" style="aspect-ratio: 4/3;">
                            <video id="scan-video" autoplay playsinline class="w-full h-full object-cover"></video>
                            <canvas id="scan-canvas" class="absolute inset-0 w-full h-full"></canvas>
                            
                            <!-- Pose Guide -->
                            <div id="scan-pose-guide" class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-48 h-60 border-4 border-dashed border-white opacity-50 rounded-full pointer-events-none"></div>
                            
                            <!-- Instruction Overlay -->
                            <div id="scan-instruction-overlay" class="absolute inset-0 bg-black bg-opacity-70 flex flex-col items-center justify-center text-white p-6 text-center">
                                <div id="scan-instruction-icon" class="text-6xl mb-4 animate-pulse">👤</div>
                                <div id="scan-instruction-text" class="text-2xl font-bold mb-2">Siap untuk memulai?</div>
                                <div id="scan-instruction-subtext" class="text-sm opacity-90">Klik "Mulai Scan" untuk memulai pendaftaran wajah</div>
                                <div class="w-full max-w-xs mt-4">
                                    <div class="w-full h-2 bg-white bg-opacity-30 rounded-full overflow-hidden">
                                        <div id="scan-progress-fill" class="h-full bg-green-500 transition-all duration-300" style="width: 0%"></div>
                                    </div>
                                </div>
                            </div>
                        </div>
                        
                        <!-- Controls -->
                        <div class="flex gap-3">
                            <button id="scan-start-btn" onclick="window.guidedScanInstance.startScanning()" class="flex-1 bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 px-6 rounded-lg transition">
                                Mulai Scan
                            </button>
                            <button id="scan-submit-btn" onclick="window.guidedScanInstance.submit()" disabled class="flex-1 bg-green-600 hover:bg-green-700 text-white font-semibold py-3 px-6 rounded-lg transition disabled:opacity-50 disabled:cursor-not-allowed">
                                Submit Foto
                            </button>
                            <button onclick="window.guidedScanInstance.cancel()" class="bg-gray-500 hover:bg-gray-600 text-white font-semibold py-3 px-6 rounded-lg transition">
                                Batal
                            </button>
                        </div>
                        
                        <!-- Tips -->
                        <div class="mt-4 bg-blue-50 border border-blue-200 rounded-lg p-4">
                            <p class="text-sm text-blue-800 font-semibold mb-2">💡 Tips:</p>
                            <ul class="text-sm text-blue-700 space-y-1">
                                <li>• Pastikan pencahayaan cukup</li>
                                <li>• Ikuti instruksi di layar untuk setiap posisi</li>
                                <li>• Tahan posisi stabil selama 1 detik untuk capture otomatis</li>
                                <li>• Total 5 foto akan diambil dari berbagai sudut</li>
                            </ul>
                        </div>
                    </div>
                </div>
            </div>
        `;
        
        // Append to body
        document.body.insertAdjacentHTML('beforeend', modalHTML);
        
        // Store references
        this.modal = document.getElementById('guided-scan-modal');
        this.video = document.getElementById('scan-video');
        this.canvas = document.getElementById('scan-canvas');
        this.ctx = this.canvas.getContext('2d');
        
        // Store instance globally for onclick handlers
        window.guidedScanInstance = this;
    }
    
    async initMediaPipe() {
        // Load MediaPipe Face Mesh
        this.faceMesh = new FaceMesh({
            locateFile: (file) => {
                return `https://cdn.jsdelivr.net/npm/@mediapipe/face_mesh/${file}`;
            }
        });
        
        this.faceMesh.setOptions({
            maxNumFaces: 1,
            refineLandmarks: true,
            minDetectionConfidence: 0.5,
            minTrackingConfidence: 0.5
        });
        
        this.faceMesh.onResults(this.onResults.bind(this));
    }
    
    async startCamera() {
        this.camera = new Camera(this.video, {
            onFrame: async () => {
                if (this.isRunning) {
                    await this.faceMesh.send({image: this.video});
                }
            },
            width: 640,
            height: 480
        });
        
        await this.camera.start();
        this.isRunning = true;
    }
    
    startScanning() {
        this.isScanning = true;
        this.currentStepIndex = 0;
        this.capturedPhotos = [];
        this.updateInstruction(this.SCAN_STEPS[0]);
        document.getElementById('scan-instruction-overlay').classList.remove('hidden');
        document.getElementById('scan-start-btn').disabled = true;
    }
    
    onResults(results) {
        // Update canvas size
        this.canvas.width = this.video.videoWidth;
        this.canvas.height = this.video.videoHeight;
        
        // Clear canvas
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        
        // Process results if scanning
        if (this.isScanning && results.multiFaceLandmarks && results.multiFaceLandmarks.length > 0) {
            const landmarks = results.multiFaceLandmarks[0];
            
            // Draw face mesh
            drawConnectors(this.ctx, landmarks, FACEMESH_TESSELATION, {color: '#C0C0C070', lineWidth: 1});
            drawConnectors(this.ctx, landmarks, FACEMESH_FACE_OVAL, {color: '#E0E0E0', lineWidth: 2});
            
            // Calculate head pose
            const pose = this.calculateHeadPose(landmarks);
            
            // Check if we're done
            if (this.currentStepIndex >= this.SCAN_STEPS.length) {
                this.completeScan();
                return;
            }
            
            // Get current step
            const currentStep = this.SCAN_STEPS[this.currentStepIndex];
            
            // Check if pose matches current step
            const poseCorrect = currentStep.check(pose);
            
            // Update pose guide
            const poseGuide = document.getElementById('scan-pose-guide');
            if (poseCorrect) {
                poseGuide.classList.add('border-green-500', 'border-solid');
                poseGuide.classList.remove('border-white', 'border-dashed');
                
                // Check stability
                if (!this.lastPoseCorrect) {
                    this.poseStableTime = Date.now();
                }
                
                const stableDuration = Date.now() - this.poseStableTime;
                
                // Update subtext with countdown
                const remaining = Math.max(0, this.STABILITY_DURATION - stableDuration);
                if (remaining > 0) {
                    document.getElementById('scan-instruction-subtext').textContent = 
                        `Tahan posisi... ${Math.ceil(remaining / 1000)}s`;
                }
                
                // Auto-capture if stable for required duration
                if (stableDuration >= this.STABILITY_DURATION) {
                    this.capturePhoto(currentStep.name);
                    this.currentStepIndex++;
                    this.updateProgress();
                    
                    // Move to next step
                    if (this.currentStepIndex < this.SCAN_STEPS.length) {
                        this.updateInstruction(this.SCAN_STEPS[this.currentStepIndex]);
                    }
                    
                    this.poseStableTime = 0;
                    this.lastPoseCorrect = false;
                } else {
                    this.lastPoseCorrect = true;
                }
            } else {
                poseGuide.classList.remove('border-green-500', 'border-solid');
                poseGuide.classList.add('border-white', 'border-dashed');
                this.poseStableTime = 0;
                this.lastPoseCorrect = false;
                
                // Reset subtext
                document.getElementById('scan-instruction-subtext').textContent = currentStep.subtext;
            }
        }
        
        // Calculate FPS
        this.frameCount++;
        const now = Date.now();
        if (now - this.lastTime >= 1000) {
            this.fps = this.frameCount;
            this.frameCount = 0;
            this.lastTime = now;
        }
    }
    
    calculateHeadPose(landmarks) {
        const noseTip = landmarks[1];
        const chin = landmarks[152];
        const leftEye = landmarks[33];
        const rightEye = landmarks[263];
        
        // Calculate yaw (left/right rotation)
        const eyeDistance = Math.abs(rightEye.x - leftEye.x);
        const noseToLeftEye = Math.abs(noseTip.x - leftEye.x);
        const noseToRightEye = Math.abs(noseTip.x - rightEye.x);
        const yaw = ((noseToRightEye - noseToLeftEye) / eyeDistance) * 90;
        
        // Calculate pitch (up/down rotation)
        const eyeY = (leftEye.y + rightEye.y) / 2;
        const noseToChin = Math.abs(chin.y - noseTip.y);
        const noseToEye = Math.abs(noseTip.y - eyeY);
        const pitch = ((noseToEye / noseToChin) - 0.5) * 60;
        
        // Calculate roll (tilt)
        const eyeDeltaY = rightEye.y - leftEye.y;
        const eyeDeltaX = rightEye.x - leftEye.x;
        const roll = Math.atan2(eyeDeltaY, eyeDeltaX) * (180 / Math.PI);
        
        return { pitch, yaw, roll };
    }
    
    capturePhoto(stepName) {
        // Create temporary canvas for capture
        const captureCanvas = document.createElement('canvas');
        captureCanvas.width = this.video.videoWidth;
        captureCanvas.height = this.video.videoHeight;
        const captureCtx = captureCanvas.getContext('2d');
        
        // Draw video frame
        captureCtx.drawImage(this.video, 0, 0);
        
        // Convert to data URL
        const photoDataURL = captureCanvas.toDataURL('image/jpeg', 0.9);
        
        // Strip data URL prefix (data:image/jpeg;base64,) to get pure base64
        const photoData = photoDataURL.replace(/^data:image\/\w+;base64,/, '');
        
        // Store photo with pure base64 data
        this.capturedPhotos.push({
            step: stepName,
            data: photoData,
            timestamp: Date.now()
        });
        
        // Update UI (use data URL for display)
        const photoSlot = document.getElementById(`photo-slot-${stepName}`);
        photoSlot.innerHTML = `
            <img src="${photoDataURL}" alt="${stepName}" class="w-full h-full object-cover rounded-lg">
            <div class="absolute top-1 right-1 bg-green-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-xs">✓</div>
        `;
        photoSlot.classList.remove('border-dashed', 'border-gray-300');
        photoSlot.classList.add('border-solid', 'border-green-500', 'relative');
        
        // Flash success
        this.flashSuccess();
    }
    
    flashSuccess() {
        const overlay = document.getElementById('scan-instruction-overlay');
        overlay.style.background = 'rgba(34, 197, 94, 0.8)';
        setTimeout(() => {
            overlay.style.background = 'rgba(0,0,0,0.7)';
        }, 300);
    }
    
    updateInstruction(step) {
        document.getElementById('scan-instruction-icon').textContent = step.icon;
        document.getElementById('scan-instruction-text').textContent = step.instruction;
        document.getElementById('scan-instruction-subtext').textContent = step.subtext;
    }
    
    updateProgress() {
        const progress = this.capturedPhotos.length;
        const total = this.SCAN_STEPS.length;
        const percentage = (progress / total) * 100;
        
        document.getElementById('scan-progress-text').textContent = `${progress}/${total}`;
        document.getElementById('scan-progress-bar').style.width = `${percentage}%`;
        document.getElementById('scan-progress-fill').style.width = `${percentage}%`;
    }
    
    completeScan() {
        this.isScanning = false;
        document.getElementById('scan-instruction-overlay').classList.add('hidden');
        document.getElementById('scan-submit-btn').disabled = false;
    }
    
    async submit() {
        if (this.capturedPhotos.length !== this.SCAN_STEPS.length) {
            alert('Harap selesaikan semua 5 langkah scan terlebih dahulu');
            return;
        }
        
        // Prepare data
        const enrollmentData = {
            photos: this.capturedPhotos,
            timestamp: Date.now(),
            metadata: {
                fps: this.fps,
                device: 'web',
                version: '1.0'
            }
        };
        
        // Call completion callback
        if (this.options.onComplete) {
            await this.options.onComplete(enrollmentData);
        }
        
        // Cleanup
        this.cleanup();
    }
    
    cancel() {
        if (this.options.onCancel) {
            this.options.onCancel();
        }
        this.cleanup();
    }
    
    cleanup() {
        // Stop camera
        if (this.camera) {
            this.camera.stop();
            this.isRunning = false;
        }
        
        // Close recognizer
        if (this.faceMesh) {
            this.faceMesh.close();
        }
        
        // Remove modal
        if (this.modal) {
            this.modal.remove();
        }
        
        // Clear global reference
        delete window.guidedScanInstance;
    }
}

// Export for use
window.GuidedFaceScan = GuidedFaceScan;
