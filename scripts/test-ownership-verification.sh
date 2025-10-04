#!/bin/bash

# Test script for wallet ownership verification
# This script demonstrates how to use the ownership verification endpoints

set -e

# Configuration
API_BASE_URL="http://localhost:8080"
PUBLIC_KEY="GDQNY3PBO1OKUMOSUQDKHXFEH3A3XEGXD6M2W42BHN7PDUNM6BB6AWD2"  # Example public key
PRIVATE_KEY="SDQNY3PBO1OKUMOSUQDKHXFEH3A3XEGXD6M2W42BHN7PDUNM6BB6AWD2"  # Example private key (for demo only)

echo "🔐 Testing Wallet Ownership Verification"
echo "========================================"
echo "API Base URL: $API_BASE_URL"
echo "Public Key: $PUBLIC_KEY"
echo ""

# Test 1: Generate Challenge
echo "📋 Test 1: Generate Challenge"
echo "-----------------------------"
CHALLENGE_RESPONSE=$(curl -s -X GET "$API_BASE_URL/api/v1/accounts/$PUBLIC_KEY/challenge")
echo "Response: $CHALLENGE_RESPONSE"

# Extract challenge from response
CHALLENGE=$(echo $CHALLENGE_RESPONSE | jq -r '.challenge')
echo "Challenge: $CHALLENGE"
echo ""

# Test 2: Verify Account Activity
echo "📋 Test 2: Verify Account Activity"
echo "----------------------------------"
ACCOUNT_RESPONSE=$(curl -s -X GET "$API_BASE_URL/api/v1/accounts/$PUBLIC_KEY/verify-account")
echo "Response: $ACCOUNT_RESPONSE"
echo ""

# Test 3: Get Account Balance (placeholder)
echo "📋 Test 3: Get Account Balance"
echo "------------------------------"
BALANCE_RESPONSE=$(curl -s -X GET "$API_BASE_URL/api/v1/accounts/$PUBLIC_KEY/balance")
echo "Response: $BALANCE_RESPONSE"
echo ""

# Test 4: Get Account Transaction History (placeholder)
echo "📋 Test 4: Get Account Transaction History"
echo "------------------------------------------"
HISTORY_RESPONSE=$(curl -s -X GET "$API_BASE_URL/api/v1/accounts/$PUBLIC_KEY/transactions?limit=5&offset=0")
echo "Response: $HISTORY_RESPONSE"
echo ""

echo "✅ Ownership verification tests completed!"
echo ""
echo "📝 Notes:"
echo "- Message signing verification requires a client-side implementation"
echo "- Transaction verification requires a valid transaction hash"
echo "- Balance and transaction history endpoints are placeholders for future implementation"
echo ""
echo "🔗 For full implementation examples, see:"
echo "   docs/WALLET_OWNERSHIP_VERIFICATION.md"
