/**
 * PasswordValidator - Reusable password validation component
 * Provides real-time validation with visual feedback
 * 
 * @author Sistem Absensi Kantor
 * @version 1.0
 * 
 * Usage:
 *   const validator = new PasswordValidator('password-input-id', 'validation-container-id');
 *   
 *   // Check if valid before submit
 *   if (validator.isValid()) {
 *       // Submit form
 *   }
 * 
 * Features:
 * - Real-time validation as user types
 * - Visual feedback with checkmarks
 * - Password strength meter
 * - Consistent validation rules across the system
 */

class PasswordValidator {
    constructor(inputId, containerId) {
        this.input = document.getElementById(inputId);
        this.container = document.getElementById(containerId);
        
        if (this.input && this.container) {
            this.init();
        }
    }
    
    /**
     * Initialize event listeners
     */
    init() {
        // Listen to input changes for real-time validation
        this.input.addEventListener('input', () => {
            this.validate();
        });
        
        // Show validation on focus if has value
        this.input.addEventListener('focus', () => {
            if (this.input.value.length > 0) {
                this.container.classList.remove('hidden');
            }
        });
        
        // Keep showing if has value on blur
        this.input.addEventListener('blur', () => {
            if (this.input.value.length === 0) {
                this.container.classList.add('hidden');
            }
        });
    }
    
    /**
     * Validate password and update UI
     * @returns {boolean} Whether password is valid
     */
    validate() {
        const password = this.input.value;
        const rules = this.checkRules(password);
        this.updateUI(rules);
        return rules.isValid();
    }
    
    /**
     * Check password against all rules
     * @param {string} password - Password to check
     * @returns {Object} Validation results
     */
    checkRules(password) {
        const rules = {
            length: password.length >= 8,
            lengthCount: password.length,
            uppercase: /[A-Z]/.test(password),
            lowercase: /[a-z]/.test(password),
            digit: /[0-9]/.test(password),
            special: /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)
        };
        
        // Add isValid method
        rules.isValid = function() {
            return this.length && this.uppercase && this.lowercase && 
                   this.digit && this.special;
        };
        
        // Add strength calculation
        rules.strength = function() {
            const score = [this.length, this.uppercase, this.lowercase, 
                          this.digit, this.special].filter(Boolean).length;
            
            if (score <= 2) {
                return { 
                    level: 'weak', 
                    text: 'Lemah', 
                    color: 'red', 
                    width: '20%',
                    bgClass: 'bg-red-500'
                };
            }
            if (score <= 3) {
                return { 
                    level: 'medium', 
                    text: 'Sedang', 
                    color: 'yellow', 
                    width: '50%',
                    bgClass: 'bg-yellow-500'
                };
            }
            if (score <= 4) {
                return { 
                    level: 'good', 
                    text: 'Baik', 
                    color: 'blue', 
                    width: '75%',
                    bgClass: 'bg-blue-500'
                };
            }
            return { 
                level: 'strong', 
                text: 'Kuat', 
                color: 'green', 
                width: '100%',
                bgClass: 'bg-green-500'
            };
        };
        
        return rules;
    }
    
    /**
     * Update UI with validation results
     * @param {Object} rules - Validation results
     */
    updateUI(rules) {
        // Show/hide container based on input value
        if (this.input.value.length === 0) {
            this.container.classList.add('hidden');
            return;
        }
        
        this.container.classList.remove('hidden');
        
        // Update each rule
        this.updateRule('rule-length', rules.length, 
                       `Minimal 8 karakter (${rules.lengthCount}/8)`);
        this.updateRule('rule-uppercase', rules.uppercase, 'Huruf besar (A-Z)');
        this.updateRule('rule-lowercase', rules.lowercase, 'Huruf kecil (a-z)');
        this.updateRule('rule-digit', rules.digit, 'Angka (0-9)');
        this.updateRule('rule-special', rules.special, 'Karakter spesial (!@#$%^&*...)');
        
        // Update strength meter
        const strength = rules.strength();
        this.updateStrength(strength);
    }
    
    /**
     * Update individual rule UI
     * @param {string} ruleId - Rule element ID
     * @param {boolean} passed - Whether rule passed
     * @param {string} text - Rule text to display
     */
    updateRule(ruleId, passed, text) {
        const rule = this.container.querySelector(`#${ruleId}`);
        if (!rule) return;
        
        const icon = rule.querySelector('.rule-icon');
        const textEl = rule.querySelector('.rule-text');
        
        if (passed) {
            icon.textContent = '✓';
            icon.className = 'rule-icon text-green-600 font-bold';
            rule.className = 'flex items-center gap-2 text-green-600';
        } else {
            icon.textContent = '✗';
            icon.className = 'rule-icon text-red-600 font-bold';
            rule.className = 'flex items-center gap-2 text-red-600';
        }
        
        textEl.textContent = text;
    }
    
    /**
     * Update password strength meter
     * @param {Object} strength - Strength information
     */
    updateStrength(strength) {
        const bar = this.container.querySelector('#strength-bar');
        const text = this.container.querySelector('#strength-text');
        
        if (!bar || !text) return;
        
        bar.style.width = strength.width;
        bar.className = `h-2 rounded-full transition-all duration-300 ${strength.bgClass}`;
        text.textContent = strength.text;
        text.className = `text-xs font-semibold text-${strength.color}-600`;
    }
    
    /**
     * Check if current password is valid
     * @returns {boolean} Whether password meets all requirements
     */
    isValid() {
        const rules = this.checkRules(this.input.value);
        return rules.isValid();
    }
    
    /**
     * Get validation error messages
     * @returns {Array<string>} Array of error messages
     */
    getErrors() {
        const password = this.input.value;
        const rules = this.checkRules(password);
        const errors = [];
        
        if (!rules.length) {
            errors.push('Password minimal 8 karakter');
        }
        if (!rules.uppercase) {
            errors.push('Harus ada huruf besar (A-Z)');
        }
        if (!rules.lowercase) {
            errors.push('Harus ada huruf kecil (a-z)');
        }
        if (!rules.digit) {
            errors.push('Harus ada angka (0-9)');
        }
        if (!rules.special) {
            errors.push('Harus ada karakter spesial (!@#$%^&*...)');
        }
        
        return errors;
    }
}

/**
 * Create validation indicator HTML
 * @param {string} containerId - Container element ID
 * @returns {string} HTML string
 */
function createValidationIndicatorHTML(containerId) {
    return `
        <div id="${containerId}" class="mt-2 p-3 bg-gray-50 rounded-lg border border-gray-200 hidden">
            <p class="text-xs font-semibold text-gray-700 mb-2">Aturan Password:</p>
            <ul class="space-y-1 text-xs">
                <li id="rule-length" class="flex items-center gap-2">
                    <span class="rule-icon text-red-600 font-bold">✗</span>
                    <span class="rule-text">Minimal 8 karakter (0/8)</span>
                </li>
                <li id="rule-uppercase" class="flex items-center gap-2">
                    <span class="rule-icon text-red-600 font-bold">✗</span>
                    <span class="rule-text">Huruf besar (A-Z)</span>
                </li>
                <li id="rule-lowercase" class="flex items-center gap-2">
                    <span class="rule-icon text-red-600 font-bold">✗</span>
                    <span class="rule-text">Huruf kecil (a-z)</span>
                </li>
                <li id="rule-digit" class="flex items-center gap-2">
                    <span class="rule-icon text-red-600 font-bold">✗</span>
                    <span class="rule-text">Angka (0-9)</span>
                </li>
                <li id="rule-special" class="flex items-center gap-2">
                    <span class="rule-icon text-red-600 font-bold">✗</span>
                    <span class="rule-text">Karakter spesial (!@#$%^&*...)</span>
                </li>
            </ul>
            
            <!-- Password Strength Bar -->
            <div class="mt-3">
                <div class="flex items-center justify-between mb-1">
                    <span class="text-xs font-semibold text-gray-700">Kekuatan Password:</span>
                    <span id="strength-text" class="text-xs font-semibold text-gray-500">-</span>
                </div>
                <div class="w-full bg-gray-200 rounded-full h-2">
                    <div id="strength-bar" class="h-2 rounded-full transition-all duration-300 bg-gray-300" style="width: 0%"></div>
                </div>
            </div>
        </div>
    `;
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { PasswordValidator, createValidationIndicatorHTML };
}
