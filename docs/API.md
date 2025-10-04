# QuasarFlow API Documentation

> Complete API reference for QuasarFlow blockchain abstraction API

## Base Information

- **Base URL**: `http://localhost:8080` (development)
- **API Version**: v1
- **Content-Type**: `application/json`
- **Authentication**: None (for development)

## Response Format

All API responses follow this consistent structure:

### Success Response
```json
{
  "success": true,
  "data": { ... },
  "error": null
}
```

### Error Response
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message"
  }
}
```

## Endpoints

### 1. Health Check

Check if the API is running and database is connected.

**Endpoint**: `GET /health`

**Response**:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "database": "healthy"
  }
}
```

---

### 2. Create Wallet

Generate a new Stellar wallet with encrypted private key storage.

**Endpoint**: `POST /api/v1/wallets`

**Request Body**:
```json
{
  "network": "local"  // "local", "testnet", or "mainnet"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "public_key": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "network": "local",
    "created_at": "2025-01-27T12:34:56Z"
  }
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/wallets \
  -H "Content-Type: application/json" \
  -d '{"network": "local"}'
```

---

### 3. Get Wallet Details

Retrieve wallet information by ID.

**Endpoint**: `GET /api/v1/wallets/{id}`

**Path Parameters**:
- `id` (string): Wallet UUID

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "public_key": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "network": "local",
    "created_at": "2025-01-27T12:34:56Z",
    "updated_at": "2025-01-27T12:34:56Z"
  }
}
```

**Example**:
```bash
curl http://localhost:8080/api/v1/wallets/a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

---

### 4. List Wallets

Get paginated list of all wallets.

**Endpoint**: `GET /api/v1/wallets`

**Query Parameters**:
- `limit` (integer, optional): Number of wallets to return (default: 10, max: 100)
- `offset` (integer, optional): Number of wallets to skip (default: 0)

**Response**:
```json
{
  "success": true,
  "data": {
    "wallets": [
      {
        "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "public_key": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
        "network": "local",
        "created_at": "2025-01-27T12:34:56Z"
      }
    ],
    "total": 1,
    "limit": 10,
    "offset": 0
  }
}
```

**Example**:
```bash
curl "http://localhost:8080/api/v1/wallets?limit=5&offset=0"
```

---

### 5. Get Wallet Balance

Query blockchain for wallet's current balances.

**Endpoint**: `GET /api/v1/wallets/{id}/balance`

**Path Parameters**:
- `id` (string): Wallet UUID

**Response**:
```json
{
  "success": true,
  "data": {
    "public_key": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "network": "local",
    "balances": [
      {
        "asset_type": "native",
        "asset_code": "XLM",
        "asset_issuer": "",
        "amount": "10000.0000000",
        "limit": null
      }
    ]
  }
}
```

**Example**:
```bash
curl http://localhost:8080/api/v1/wallets/a1b2c3d4-e5f6-7890-abcd-ef1234567890/balance
```

---

### 6. Fund Wallet (Development Networks Only)

Fund a wallet using Friendbot (testnet/local networks only).

**Endpoint**: `POST /api/v1/wallets/{id}/fund`

**Path Parameters**:
- `id` (string): Wallet UUID

**Request Body** (optional):
```json
{
  "amount": "1000"  // Optional, defaults to Friendbot default
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "wallet_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "public_key": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "network": "local",
    "transaction_id": "",
    "message": "Wallet successfully funded with 10000 XLM",
    "success": true
  }
}
```

**Error Response** (mainnet):
```json
{
  "success": true,
  "data": {
    "wallet_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "public_key": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "network": "mainnet",
    "message": "Friendbot is not available on mainnet. Please fund this wallet manually.",
    "success": false
  }
}
```

**Examples**:
```bash
# Fund with default amount
curl -X POST http://localhost:8080/api/v1/wallets/a1b2c3d4-e5f6-7890-abcd-ef1234567890/fund

# Fund with specific amount
curl -X POST http://localhost:8080/api/v1/wallets/a1b2c3d4-e5f6-7890-abcd-ef1234567890/fund \
  -H "Content-Type: application/json" \
  -d '{"amount": "5000"}'
```

---

### 7. Send Payment

Send XLM or assets from one wallet to another address.

**Endpoint**: `POST /api/v1/wallets/{id}/payment`

**Path Parameters**:
- `id` (string): Source wallet UUID

**Request Body**:
```json
{
  "to_address": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
  "amount": "100.50",
  "asset_code": "XLM",      // Optional, defaults to "XLM"
  "asset_issuer": "",       // Required for non-native assets
  "memo": "Payment memo"    // Optional
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "transaction_hash": "abc123def456...",
    "from_address": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "to_address": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "amount": "100.50",
    "asset_code": "XLM",
    "asset_issuer": "",
    "memo": "Payment memo",
    "network": "local",
    "ledger": 12345,
    "success": true
  }
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/wallets/a1b2c3d4-e5f6-7890-abcd-ef1234567890/payment \
  -H "Content-Type: application/json" \
  -d '{
    "to_address": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "amount": "50.0",
    "memo": "Test payment"
  }'
```

---

### 8. Get Transaction History

Retrieve transaction history for a wallet.

**Endpoint**: `GET /api/v1/wallets/{id}/transactions`

**Path Parameters**:
- `id` (string): Wallet UUID

**Query Parameters**:
- `limit` (integer, optional): Number of transactions to return (default: 10, max: 200)
- `order` (string, optional): Sort order - "asc" or "desc" (default: "desc")
- `cursor` (string, optional): Pagination cursor for next page

**Response**:
```json
{
  "success": true,
  "data": {
    "wallet_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "public_key": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "network": "local",
    "transactions": [
      {
        "id": "123456789",
        "hash": "abc123def456...",
        "ledger": 12345,
        "created_at": "2025-01-27T12:34:56Z",
        "source_account": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
        "type": "payment",
        "type_i": 1,
        "operation_count": 1,
        "successful": true,
        "max_fee": 100,
        "fee_charged": 100,
        "memo_type": "text",
        "memo": "Payment memo"
      }
    ],
    "operations": [
      {
        "id": "123456789-1",
        "type": "payment",
        "created_at": "2025-01-27T12:34:56Z",
        "transaction_id": "abc123def456...",
        "from": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
        "to": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
        "amount": "100.50",
        "asset_type": "native",
        "asset_code": "XLM"
      }
    ],
    "has_next": false,
    "next_cursor": ""
  }
}
```

**Examples**:
```bash
# Get recent transactions
curl http://localhost:8080/api/v1/wallets/a1b2c3d4-e5f6-7890-abcd-ef1234567890/transactions

# Get with pagination
curl "http://localhost:8080/api/v1/wallets/a1b2c3d4-e5f6-7890-abcd-ef1234567890/transactions?limit=5&order=asc"

# Get next page
curl "http://localhost:8080/api/v1/wallets/a1b2c3d4-e5f6-7890-abcd-ef1234567890/transactions?cursor=123456789"
```

---

## Error Codes

| Code | Description |
|------|-------------|
| `WALLET_NOT_FOUND` | Wallet with specified ID does not exist |
| `INVALID_WALLET_ID` | Wallet ID format is invalid |
| `INVALID_REQUEST_BODY` | Request body is malformed or missing required fields |
| `INVALID_NETWORK` | Network must be 'local', 'testnet', or 'mainnet' |
| `INSUFFICIENT_BALANCE` | Wallet doesn't have enough funds for the transaction |
| `TRANSACTION_FAILED` | Stellar transaction failed (check details in message) |
| `FRIENDBOT_UNAVAILABLE` | Friendbot service is not available |
| `RATE_LIMIT_EXCEEDED` | Too many requests (when rate limiting is enabled) |

---

## Development Workflow

### 1. Setup Local Environment
```bash
# Start all services
./scripts/setup-local.sh

# Or manually with docker-compose
docker-compose up -d
```

### 2. Create and Fund a Wallet
```bash
# 1. Create wallet
WALLET_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/wallets \
  -H "Content-Type: application/json" \
  -d '{"network": "local"}')

# 2. Extract wallet ID and public key
WALLET_ID=$(echo $WALLET_RESPONSE | jq -r '.data.id')
PUBLIC_KEY=$(echo $WALLET_RESPONSE | jq -r '.data.public_key')

# 3. Fund the wallet
curl -X POST http://localhost:8080/api/v1/wallets/$WALLET_ID/fund

# 4. Check balance
curl http://localhost:8080/api/v1/wallets/$WALLET_ID/balance
```

### 3. Send Test Payment
```bash
# Create second wallet for testing
WALLET2_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/wallets \
  -H "Content-Type: application/json" \
  -d '{"network": "local"}')

WALLET2_PUBLIC=$(echo $WALLET2_RESPONSE | jq -r '.data.public_key')

# Send payment from first wallet to second
curl -X POST http://localhost:8080/api/v1/wallets/$WALLET_ID/payment \
  -H "Content-Type: application/json" \
  -d "{
    \"to_address\": \"$WALLET2_PUBLIC\",
    \"amount\": \"100.0\",
    \"memo\": \"Test payment\"
  }"
```

### 4. Check Transaction History
```bash
curl http://localhost:8080/api/v1/wallets/$WALLET_ID/transactions
```

---

## Network Configuration

### Local Development (Recommended)
- **Network**: `local`
- **Stellar Horizon**: `http://localhost:8000`
- **Friendbot**: `http://localhost:8000/friendbot`
- **Network Passphrase**: `"Standalone Network ; February 2017"`

### Testnet
- **Network**: `testnet`
- **Stellar Horizon**: `https://horizon-testnet.stellar.org`
- **Friendbot**: `https://horizon-testnet.stellar.org/friendbot`
- **Network Passphrase**: Stellar test network

### Mainnet (Production)
- **Network**: `mainnet`
- **Stellar Horizon**: `https://horizon.stellar.org`
- **Friendbot**: Not available (manual funding required)
- **Network Passphrase**: Stellar public network

---

## Rate Limits

Current implementation has no rate limits, but the following are planned:

- **Wallet Creation**: 10 per minute per IP
- **Payments**: 60 per minute per wallet
- **Balance Queries**: 100 per minute per IP
- **Friendbot Funding**: 1 per minute per wallet

---

## Best Practices

### Security
1. **Never expose private keys** - They are always encrypted at rest
2. **Use HTTPS in production** - All API calls should be over secure connections
3. **Rotate encryption keys** - Regularly update your `ENCRYPTION_KEY`
4. **Monitor transactions** - Set up alerts for unusual activity

### Performance
1. **Cache balance queries** - Stellar balances don't change instantly
2. **Use pagination** - Don't fetch large transaction histories at once
3. **Batch operations** - Group multiple operations when possible

### Development
1. **Use local network** - Faster for development and testing
2. **Fund test wallets** - Use Friendbot for development wallets
3. **Check transaction status** - Always verify transactions succeeded
4. **Handle errors gracefully** - Network issues are common in blockchain

---

## Support

For issues and questions:
- **GitHub Issues**: [Project Issues](https://github.com/QuasarAPI/quasarflow-api/issues)
- **Documentation**: Check the `docs/` directory
- **Logs**: Use `docker-compose logs quasarflow-api` for debugging