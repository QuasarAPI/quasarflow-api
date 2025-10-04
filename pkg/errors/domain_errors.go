package errors

import "fmt"

// Stellar domain-specific errors
var (
	// ErrInvalidPublicKey is returned when a Stellar public key is invalid
	ErrInvalidPublicKey = NewValidationError(
		"Invalid Stellar public key",
		"Public key must start with 'G' and be 56 characters long",
	)

	// ErrInvalidNetwork is returned when the specified Stellar network is invalid
	ErrInvalidNetwork = NewValidationError(
		"Invalid network",
		"Network must be either 'testnet' or 'mainnet'",
	)

	// ErrInvalidAssetType is returned when an invalid Stellar asset type is specified
	ErrInvalidAssetType = NewValidationError(
		"Invalid asset type",
		"Asset type must be one of: native, credit_alphanum4, credit_alphanum12",
	)

	// ErrNegativeAmount is returned when attempting to use a negative amount
	ErrNegativeAmount = NewValidationError(
		"Invalid amount",
		"Amount cannot be negative",
	)

	// ErrInvalidIssuer is returned when an invalid asset issuer is specified
	ErrInvalidIssuer = NewValidationError(
		"Invalid asset issuer",
		"Asset issuer must be a valid Stellar public key",
	)
)

// Cryptography-specific errors
var (
	// ErrInvalidKeySize is returned when the encryption key size is not 32 bytes
	ErrInvalidKeySize = NewCryptoError(
		"Invalid encryption key size",
		fmt.Errorf("encryption key must be exactly 32 bytes"),
	)

	// ErrEncryptionFailed is returned when data encryption fails
	ErrEncryptionFailed = NewCryptoError(
		"Encryption failed",
		nil,
	)

	// ErrDecryptionFailed is returned when data decryption fails
	ErrDecryptionFailed = NewCryptoError(
		"Decryption failed",
		nil,
	)

	// ErrInvalidCiphertext is returned when the ciphertext is invalid or corrupted
	ErrInvalidCiphertext = NewCryptoError(
		"Invalid ciphertext",
		fmt.Errorf("ciphertext is corrupted or invalid"),
	)
)

// Erros específicos de blockchain
var (
	ErrHorizonConnection = NewBlockchainError(
		"Failed to connect to Stellar Horizon",
		nil,
	)

	ErrTransactionFailed = NewBlockchainError(
		"Transaction failed",
		nil,
	)

	ErrAccountNotFound = NewBlockchainError(
		"Account not found",
		nil,
	)
)

// Authentication-specific errors
var (
	// ErrMissingAuthHeader is returned when Authorization header is missing
	ErrMissingAuthHeader = &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    "Authorization header is required",
		StatusCode: 401,
	}

	// ErrInvalidAuthFormat is returned when Authorization header format is invalid
	ErrInvalidAuthFormat = &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    "Invalid authorization header format",
		StatusCode: 401,
	}

	// ErrInvalidToken is returned when JWT token is invalid or expired
	ErrInvalidToken = &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    "Invalid or expired token",
		StatusCode: 401,
	}

	// ErrInvalidSigningMethod is returned when JWT signing method is unexpected
	ErrInvalidSigningMethod = &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    "Unexpected signing method",
		StatusCode: 401,
	}

	// ErrInvalidTokenClaims is returned when JWT claims are invalid
	ErrInvalidTokenClaims = &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    "Invalid token claims",
		StatusCode: 401,
	}

	// ErrInvalidTokenIssuer is returned when JWT issuer is invalid
	ErrInvalidTokenIssuer = &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    "Invalid token issuer",
		StatusCode: 401,
	}

	// ErrTokenExpired is returned when JWT token has expired
	ErrTokenExpired = &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    "Token has expired",
		StatusCode: 401,
	}

	// ErrUserRoleNotFound is returned when user role is not found in context
	ErrUserRoleNotFound = &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    "User role not found",
		StatusCode: 403,
	}

	// ErrInsufficientPermissions is returned when user lacks required permissions
	ErrInsufficientPermissions = &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    "Insufficient permissions",
		StatusCode: 403,
	}
)
