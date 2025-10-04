# 🔐 Wallet Ownership Verification

This document describes the wallet ownership verification system implemented using Stellar Go SDK and SEP-10 Web Authentication standards.

## Overview

The QuasarFlow API now supports multiple methods for external wallet users to prove ownership of their Stellar wallets without requiring them to register with the platform. This enables users with existing wallets to use the API's functionality.

## Supported Verification Methods

### 1. Message Signing (SEP-10) ⭐ **Recommended**
Uses cryptographic message signing to prove wallet ownership.

### 2. Transaction Verification
Uses recent signed transactions as proof of ownership.

### 3. Account Activity Verification
Verifies ownership based on account existence and recent activity.

## API Endpoints

### Public Endpoints (No Authentication Required)

All ownership verification endpoints are public and don't require API authentication, allowing external users to verify their wallet ownership.

#### 1. Generate Challenge
```http
GET /api/v1/accounts/{public_key}/challenge
```

**Response:**
```json
{
  "challenge": "1703123456.789012345.quasarflow-api.com.GABC123...",
  "message": "Sign this challenge with your private key to verify ownership",
  "public_key": "GABC123...",
  "instructions": "Use Stellar SDK to sign the challenge with your private key"
}
```

#### 2. Verify Ownership (Message Signing)
```http
POST /api/v1/accounts/{public_key}/verify-ownership
```

**Request Body:**
```json
{
  "signature": "base64_encoded_signature",
  "message": "challenge_string"
}
```

**Response:**
```json
{
  "is_owner": true,
  "message": "Ownership verified successfully"
}
```

#### 3. Verify Ownership (Transaction)
```http
POST /api/v1/accounts/{public_key}/verify-transaction
```

**Request Body:**
```json
{
  "transaction_hash": "transaction_hash_string"
}
```

**Response:**
```json
{
  "is_owner": true,
  "message": "Ownership verified via transaction"
}
```

#### 4. Verify Ownership (Account Activity)
```http
GET /api/v1/accounts/{public_key}/verify-account
```

**Response:**
```json
{
  "is_owner": true,
  "message": "Account exists and has recent activity"
}
```

#### 5. Get Account Balance
```http
GET /api/v1/accounts/{public_key}/balance
```

**Response:**
```json
{
  "public_key": "GABC123...",
  "message": "Balance endpoint not yet implemented for external wallets",
  "note": "Use the wallet endpoints for registered wallets"
}
```

#### 6. Get Account Transaction History
```http
GET /api/v1/accounts/{public_key}/transactions?limit=10&offset=0
```

**Response:**
```json
{
  "public_key": "GABC123...",
  "limit": 10,
  "offset": 0,
  "message": "Transaction history endpoint not yet implemented for external wallets",
  "note": "Use the wallet endpoints for registered wallets"
}
```

## Client Implementation Examples

### JavaScript/Node.js Example

```javascript
const StellarSdk = require('stellar-sdk');

async function verifyWalletOwnership(secretKey, publicKey) {
    try {
        // 1. Parse keypair
        const keypair = StellarSdk.Keypair.fromSecret(secretKey);
        
        // Verify public key matches
        if (keypair.publicKey() !== publicKey) {
            throw new Error('Private key does not match public key');
        }
        
        // 2. Get challenge from API
        const challengeResponse = await fetch(
            `http://localhost:8080/api/v1/accounts/${publicKey}/challenge`
        );
        const { challenge } = await challengeResponse.json();
        
        // 3. Sign the challenge
        const message = Buffer.from(challenge);
        const signature = keypair.sign(message);
        const signatureBase64 = signature.toString('base64');
        
        // 4. Verify ownership
        const verifyResponse = await fetch(
            `http://localhost:8080/api/v1/accounts/${publicKey}/verify-ownership`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    signature: signatureBase64,
                    message: challenge
                })
            }
        );
        
        const result = await verifyResponse.json();
        return result.is_owner;
        
    } catch (error) {
        console.error('Verification failed:', error);
        return false;
    }
}

// Usage
const secretKey = 'S...'; // User's private key
const publicKey = 'G...'; // User's public key

verifyWalletOwnership(secretKey, publicKey)
    .then(isOwner => {
        console.log('Is owner:', isOwner);
    });
```

### Go Example

```go
package main

import (
    "encoding/base64"
    "fmt"
    "github.com/stellar/go/keypair"
)

func verifyWalletOwnership(secretKey, publicKey, challenge string) (bool, error) {
    // 1. Parse keypair
    kp, err := keypair.Parse(secretKey)
    if err != nil {
        return false, fmt.Errorf("failed to parse secret key: %w", err)
    }
    
    // 2. Verify public key matches
    if kp.Address() != publicKey {
        return false, fmt.Errorf("private key does not match public key")
    }
    
    // 3. Sign the challenge
    message := []byte(challenge)
    signature, err := kp.Sign(message)
    if err != nil {
        return false, fmt.Errorf("failed to sign message: %w", err)
    }
    
    // 4. Encode signature
    signatureBase64 := base64.StdEncoding.EncodeToString(signature)
    
    // 5. Send to API (implement HTTP client)
    // ... HTTP POST request implementation ...
    
    return true, nil
}

func main() {
    secretKey := "SCZANGBA5YHTNYVVV4C3U252E2B6P6F5T3U6MM63WBSBZATAQI3EBTQ4"
    publicKey := "GABC123..."
    challenge := "1703123456.789012345.quasarflow-api.com.GABC123..."
    
    isOwner, err := verifyWalletOwnership(secretKey, publicKey, challenge)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Is owner: %v\n", isOwner)
}
```

### Python Example

```python
import base64
from stellar_sdk import Keypair

def verify_wallet_ownership(secret_key, public_key, challenge):
    try:
        # 1. Parse keypair
        keypair = Keypair.from_secret(secret_key)
        
        # 2. Verify public key matches
        if keypair.public_key != public_key:
            raise ValueError("Private key does not match public key")
        
        # 3. Sign the challenge
        message = challenge.encode('utf-8')
        signature = keypair.sign(message)
        signature_base64 = base64.b64encode(signature).decode('utf-8')
        
        # 4. Send to API (implement HTTP client)
        # ... HTTP POST request implementation ...
        
        return True
        
    except Exception as e:
        print(f"Verification failed: {e}")
        return False

# Usage
secret_key = "S..."
public_key = "G..."
challenge = "1703123456.789012345.quasarflow-api.com.G..."

is_owner = verify_wallet_ownership(secret_key, public_key, challenge)
print(f"Is owner: {is_owner}")
```

## Security Features

### 1. Challenge-Response Authentication
- **SEP-10 Compliant**: Follows Stellar's standard for web authentication
- **Timestamp-based**: Challenges include timestamps to prevent replay attacks
- **Domain Verification**: Challenges include domain information
- **Unique Nonces**: Each challenge includes a unique nonce

### 2. Cryptographic Verification
- **Ed25519 Signatures**: Uses Stellar's native signature algorithm
- **Message Integrity**: Verifies message hasn't been tampered with
- **Public Key Validation**: Ensures public key format is valid

### 3. Transaction Verification
- **Recent Activity**: Only accepts transactions from last 24 hours
- **Source Verification**: Ensures transaction was signed by the specified wallet
- **Replay Protection**: Prevents reuse of old transactions

### 4. Account Verification
- **Network Validation**: Verifies account exists on Stellar network
- **Activity Check**: Ensures account has recent activity (within 30 days)
- **Format Validation**: Validates Stellar public key format

## Error Handling

The API returns appropriate HTTP status codes:

- **200 OK**: Verification successful
- **400 Bad Request**: Invalid request format or parameters
- **401 Unauthorized**: Verification failed (invalid signature/transaction)
- **500 Internal Server Error**: Server error during verification

## Rate Limiting

All ownership verification endpoints are subject to the same rate limiting as other API endpoints to prevent abuse.

## Future Enhancements

### Planned Features
1. **Full Balance API**: Direct balance fetching for external wallets
2. **Transaction History**: Complete transaction history for external wallets
3. **Multi-signature Support**: Verification for multi-signature wallets
4. **Session Management**: JWT tokens for verified ownership sessions
5. **Webhook Integration**: Real-time notifications for wallet events

### Integration Possibilities
1. **Third-party Wallets**: Integration with popular Stellar wallets
2. **Mobile Apps**: SDK for mobile applications
3. **Browser Extensions**: Support for browser-based wallets
4. **Hardware Wallets**: Integration with hardware wallet providers

## Best Practices

### For Developers
1. **Always validate public keys** before making API calls
2. **Handle errors gracefully** and provide user feedback
3. **Cache verification results** to avoid repeated API calls
4. **Use HTTPS** in production environments
5. **Implement proper logging** for debugging

### For Users
1. **Never share private keys** with third parties
2. **Verify domain** before entering private keys
3. **Use secure connections** (HTTPS) only
4. **Keep private keys safe** and backed up
5. **Report suspicious activity** immediately

## Support

For technical support or questions about wallet ownership verification:

- **Documentation**: Check this guide and API documentation
- **Issues**: Report bugs or request features via GitHub issues
- **Community**: Join our community discussions
- **Security**: Report security issues privately via security@quasarflow.com

---

*This verification system enables secure, decentralized wallet ownership proof while maintaining user privacy and security.*
