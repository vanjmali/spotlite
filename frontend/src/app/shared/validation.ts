// ============================================================================
// VALIDATION PATTERNS & TYPES
// ============================================================================

/**
 * Validation result interface for form field validation
 */
export interface ValidationResult {
  isValid: boolean;
  error?: string;
}

// ============================================================================
// REGEX PATTERNS
// ============================================================================

/**
 * Email validation pattern
 * Matches: simple@example.com, user.name+tag@domain.co.uk
 */
export const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

/**
 * Password criteria patterns
 * Must contain: uppercase, lowercase, digit, special char
 */
export const PASSWORD_PATTERNS = {
  // At least 1 uppercase letter
  uppercase: /[A-Z]/,

  // At least 1 lowercase letter
  lowercase: /[a-z]/,

  // At least 1 digit
  digit: /\d/,

  // At least 1 special character (matching backend: !@#$%^&*.)
  special: /[!@#$%^&*.]/,
} as const; /**
 * Minimum password length requirement
 */
export const MIN_PASSWORD_LENGTH = 10;

/**
 * Username validation pattern
 * Allows: letters, numbers, dots, underscores
 * Length: 2+ characters (minimum enforced in TextInputComponent)
 */
export const USERNAME_PATTERN = /^[a-zA-Z0-9._]{2,}$/;

/**
 * OTP code pattern
 * Matches: exactly 6 digits
 */
export const OTP_PATTERN = /^\d{6}$/;

// ============================================================================
// ERROR MESSAGES
// ============================================================================

export const VALIDATION_MESSAGES = {
  EMAIL_REQUIRED: 'Email address is required',
  EMAIL_INVALID: 'Please enter a valid email address',
  PASSWORD_REQUIRED: 'Password is required',
  PASSWORD_CRITERIA:
    'Password must contain uppercase, lowercase, number, special character, and be at least 10 characters',
  TEXT_REQUIRED: (label: string) => `${label} is required`,
  OTP_REQUIRED: 'OTP code is required',
  OTP_INVALID: 'Invalid OTP code.',
  PASSWORDS_MISMATCH: 'Passwords do not match.',
  USERNAME_REQUIRED: 'Username is required',
  EMAIL_IN_USE: 'Email is already in use.',
  USERNAME_IN_USE: 'Username is already in use.',
  REGISTRATION_FAILED: 'Registration failed. Please try again.',
  INVALID_CREDENTIALS: 'Invalid email or password.',
  LOGIN_FAILED: 'Login failed. Check your credentials and try again.',
  OTP_EXPIRED: 'Verification code has expired',
  OTP_INVALID_CODE: 'Invalid code',
  OTP_VERIFICATION_FAILED: 'OTP verification failed',
  OTP_RESEND_FAILED: 'Failed to resend OTP. Please try again.',
  MISSING_EMAIL: 'Missing email',
  PASSWORD_RESET_FAILED: 'Failed to reset password. Please try again.',
  INVALID_RECOVERY_LINK: 'Invalid or expired recovery link',
} as const;
