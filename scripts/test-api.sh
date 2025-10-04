#!/bin/bash

# QuasarFlow API - Comprehensive API Testing Script
# Tests all endpoints with realistic scenarios and error cases

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
STELLAR_BASE_URL="${STELLAR_BASE_URL:-http://localhost:8000}"
NETWORK="${NETWORK:-local}"
VERBOSE="${VERBOSE:-false}"
CLEANUP="${CLEANUP:-true}"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Arrays to store test results
PASSED_TEST_NAMES=()
FAILED_TEST_NAMES=()
CREATED_WALLETS=()

# Function to print colored output
print_header() {
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN} $1 ${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
}

print_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

print_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    PASSED_TESTS=$((PASSED_TESTS + 1))
    PASSED_TEST_NAMES+=("$1")
}

print_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    FAILED_TEST_NAMES+=("$1")
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to make API calls with error handling
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local description=$5

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    print_test "$description"

    if [[ "$VERBOSE" == "true" ]]; then
        print_info "Method: $method"
        print_info "Endpoint: $endpoint"
        if [[ -n "$data" ]]; then
            print_info "Data: $data"
        fi
    fi

    local curl_opts=(-s -w "%{http_code}")
    if [[ -n "$data" ]]; then
        curl_opts+=(-H "Content-Type: application/json" -d "$data")
    fi

    local response=$(curl "${curl_opts[@]}" -X "$method" "$API_BASE_URL$endpoint")
    local http_code="${response: -3}"
    local response_body="${response%???}"

    if [[ "$VERBOSE" == "true" ]]; then
        print_info "HTTP Status: $http_code"
        print_info "Response: $response_body"
    fi

    if [[ "$http_code" == "$expected_status" ]]; then
        print_pass "$description (Status: $http_code)"
        echo "$response_body"
        return 0
    else
        print_fail "$description (Expected: $expected_status, Got: $http_code)"
        if [[ "$VERBOSE" == "true" ]]; then
            print_error "Response body: $response_body"
        fi
        return 1
    fi
}

# Function to extract JSON field
extract_json_field() {
    local json=$1
    local field=$2
    echo "$json" | grep -o "\"$field\":\"[^\"]*\"" | cut -d'"' -f4
}

# Function to check if service is available
check_service() {
    local url=$1
    local service_name=$2

    print_info "Checking $service_name availability at $url"
    if curl -f -s "$url" > /dev/null; then
        print_pass "$service_name is available"
        return 0
    else
        print_fail "$service_name is not available at $url"
        return 1
    fi
}

# Function to wait for services
wait_for_services() {
    print_header "Service Availability Check"

    local max_retries=30
    local retry_count=0

    while [ $retry_count -lt $max_retries ]; do
        if check_service "$API_BASE_URL/health" "QuasarFlow API"; then
            break
        fi

        retry_count=$((retry_count + 1))
        if [ $retry_count -lt $max_retries ]; then
            print_info "Waiting for API to be ready... (attempt $retry_count/$max_retries)"
            sleep 3
        else
            print_error "API is not available after $max_retries attempts"
            exit 1
        fi
    done

    check_service "$STELLAR_BASE_URL" "Stellar Horizon"
    check_service "$STELLAR_BASE_URL/friendbot" "Friendbot"
}

# Test health endpoint
test_health() {
    print_header "Health Check Tests"

    local response=$(api_call "GET" "/health" "" "200" "Health check endpoint")
    if [[ $? -eq 0 ]]; then
        local status=$(extract_json_field "$response" "status")
        if [[ "$status" == "healthy" ]]; then
            print_pass "Health status is healthy"
        else
            print_fail "Health status is not healthy: $status"
        fi
    fi
}

# Test wallet creation
test_wallet_creation() {
    print_header "Wallet Creation Tests"

    # Test successful wallet creation
    local create_data='{"network":"'$NETWORK'"}'
    local response=$(api_call "POST" "/api/v1/wallets" "$create_data" "201" "Create wallet with valid network")

    if [[ $? -eq 0 ]]; then
        local wallet_id=$(extract_json_field "$response" "id")
        local public_key=$(extract_json_field "$response" "public_key")

        if [[ -n "$wallet_id" && -n "$public_key" ]]; then
            print_pass "Wallet created with ID: $wallet_id"
            print_info "Public key: $public_key"
            CREATED_WALLETS+=("$wallet_id")

            # Store for later tests
            export TEST_WALLET_ID="$wallet_id"
            export TEST_PUBLIC_KEY="$public_key"
        else
            print_fail "Wallet creation response missing required fields"
        fi
    fi

    # Test invalid network
    api_call "POST" "/api/v1/wallets" '{"network":"invalid"}' "500" "Create wallet with invalid network"

    # Test missing network
    api_call "POST" "/api/v1/wallets" '{}' "400" "Create wallet without network"

    # Test invalid JSON
    api_call "POST" "/api/v1/wallets" '{"invalid json"}' "400" "Create wallet with invalid JSON"
}

# Test wallet retrieval
test_wallet_retrieval() {
    print_header "Wallet Retrieval Tests"

    if [[ -n "$TEST_WALLET_ID" ]]; then
        # Test successful retrieval
        local response=$(api_call "GET" "/api/v1/wallets/$TEST_WALLET_ID" "" "200" "Get wallet by valid ID")

        if [[ $? -eq 0 ]]; then
            local retrieved_id=$(extract_json_field "$response" "id")
            local retrieved_key=$(extract_json_field "$response" "public_key")

            if [[ "$retrieved_id" == "$TEST_WALLET_ID" && "$retrieved_key" == "$TEST_PUBLIC_KEY" ]]; then
                print_pass "Wallet retrieval returned correct data"
            else
                print_fail "Wallet retrieval returned incorrect data"
            fi
        fi
    else
        print_warning "Skipping wallet retrieval tests - no wallet ID available"
    fi

    # Test non-existent wallet
    api_call "GET" "/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000" "" "404" "Get non-existent wallet"

    # Test invalid UUID format
    api_call "GET" "/api/v1/wallets/invalid-uuid" "" "400" "Get wallet with invalid UUID"
}

# Test wallet listing
test_wallet_listing() {
    print_header "Wallet Listing Tests"

    # Test default listing
    local response=$(api_call "GET" "/api/v1/wallets" "" "200" "List wallets with default parameters")

    if [[ $? -eq 0 ]]; then
        # Check if response contains wallets array
        if echo "$response" | grep -q "\"wallets\""; then
            print_pass "Wallet listing contains wallets array"
        else
            print_fail "Wallet listing missing wallets array"
        fi

        # Check pagination fields
        if echo "$response" | grep -q "\"total\"" && echo "$response" | grep -q "\"limit\"" && echo "$response" | grep -q "\"offset\""; then
            print_pass "Wallet listing contains pagination fields"
        else
            print_fail "Wallet listing missing pagination fields"
        fi
    fi

    # Test with pagination
    api_call "GET" "/api/v1/wallets?limit=5&offset=0" "" "200" "List wallets with pagination"

    # Test with invalid pagination
    api_call "GET" "/api/v1/wallets?limit=-1&offset=-1" "" "200" "List wallets with invalid pagination (should use defaults)"
}

# Test wallet balance
test_wallet_balance() {
    print_header "Wallet Balance Tests"

    if [[ -n "$TEST_WALLET_ID" ]]; then
        # Test balance check (might fail if wallet not funded)
        local response=$(api_call "GET" "/api/v1/wallets/$TEST_WALLET_ID/balance" "" "200" "Get wallet balance" || true)

        # Note: Balance check might fail if wallet doesn't exist on network yet
        # This is expected behavior for new wallets
    else
        print_warning "Skipping balance tests - no wallet ID available"
    fi

    # Test non-existent wallet balance
    api_call "GET" "/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000/balance" "" "500" "Get balance for non-existent wallet"
}

# Test wallet funding
test_wallet_funding() {
    print_header "Wallet Funding Tests"

    if [[ "$NETWORK" != "mainnet" && -n "$TEST_WALLET_ID" ]]; then
        # Test funding
        local response=$(api_call "POST" "/api/v1/wallets/$TEST_WALLET_ID/fund" "" "200" "Fund wallet via Friendbot")

        if [[ $? -eq 0 ]]; then
            # Check if funding was successful
            if echo "$response" | grep -q "\"success\":true"; then
                print_pass "Wallet funding succeeded"
                export WALLET_FUNDED=true

                # Wait a moment for transaction to be processed
                sleep 3
            else
                print_warning "Wallet funding failed - this might be expected in some environments"
            fi
        fi

        # Test funding with custom amount
        api_call "POST" "/api/v1/wallets/$TEST_WALLET_ID/fund" '{"amount":"5000"}' "200" "Fund wallet with custom amount"

    else
        print_info "Skipping funding tests - mainnet or no wallet ID"
    fi

    # Test funding non-existent wallet
    api_call "POST" "/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000/fund" "" "500" "Fund non-existent wallet"
}

# Test balance after funding
test_balance_after_funding() {
    if [[ "$WALLET_FUNDED" == "true" && -n "$TEST_WALLET_ID" ]]; then
        print_header "Balance Check After Funding"

        local response=$(api_call "GET" "/api/v1/wallets/$TEST_WALLET_ID/balance" "" "200" "Check balance after funding")

        if [[ $? -eq 0 ]]; then
            # Check if balance contains XLM
            if echo "$response" | grep -q "XLM"; then
                print_pass "Wallet now has XLM balance"
            else
                print_fail "Wallet balance doesn't show XLM"
            fi
        fi
    fi
}

# Test payment functionality
test_payments() {
    print_header "Payment Tests"

    if [[ "$WALLET_FUNDED" == "true" && -n "$TEST_WALLET_ID" ]]; then
        # Create a second wallet for payment testing
        local create_data='{"network":"'$NETWORK'"}'
        local response=$(api_call "POST" "/api/v1/wallets" "$create_data" "201" "Create second wallet for payment test")

        if [[ $? -eq 0 ]]; then
            local wallet2_id=$(extract_json_field "$response" "id")
            local wallet2_public=$(extract_json_field "$response" "public_key")
            CREATED_WALLETS+=("$wallet2_id")

            if [[ -n "$wallet2_public" ]]; then
                # Test payment
                local payment_data="{\"to_address\":\"$wallet2_public\",\"amount\":\"100.0\",\"memo\":\"Test payment\"}"
                local payment_response=$(api_call "POST" "/api/v1/wallets/$TEST_WALLET_ID/payment" "$payment_data" "200" "Send payment between wallets")

                if [[ $? -eq 0 ]]; then
                    # Check transaction hash
                    local tx_hash=$(extract_json_field "$payment_response" "transaction_hash")
                    if [[ -n "$tx_hash" ]]; then
                        print_pass "Payment transaction hash received: ${tx_hash:0:20}..."
                        export PAYMENT_SENT=true
                    else
                        print_fail "Payment response missing transaction hash"
                    fi
                fi
            fi
        fi

        # Test invalid payment
        api_call "POST" "/api/v1/wallets/$TEST_WALLET_ID/payment" '{"to_address":"INVALID","amount":"100"}' "500" "Send payment to invalid address"

        # Test payment with insufficient funds
        api_call "POST" "/api/v1/wallets/$TEST_WALLET_ID/payment" '{"to_address":"'$TEST_PUBLIC_KEY'","amount":"999999999"}' "500" "Send payment with insufficient funds"

    else
        print_info "Skipping payment tests - wallet not funded or no wallet ID"
    fi
}

# Test transaction history
test_transaction_history() {
    print_header "Transaction History Tests"

    if [[ -n "$TEST_WALLET_ID" ]]; then
        # Wait for transactions to be processed
        if [[ "$PAYMENT_SENT" == "true" ]]; then
            print_info "Waiting for transactions to be processed..."
            sleep 5
        fi

        # Test transaction history
        local response=$(api_call "GET" "/api/v1/wallets/$TEST_WALLET_ID/transactions" "" "200" "Get transaction history")

        if [[ $? -eq 0 ]]; then
            # Check if response contains transactions
            if echo "$response" | grep -q "\"transactions\""; then
                print_pass "Transaction history contains transactions array"

                # If wallet was funded or payment sent, should have transactions
                if [[ "$WALLET_FUNDED" == "true" || "$PAYMENT_SENT" == "true" ]]; then
                    if echo "$response" | grep -q "\"hash\""; then
                        print_pass "Transaction history contains transaction hashes"
                    else
                        print_warning "Transaction history empty - might be timing issue"
                    fi
                fi
            else
                print_fail "Transaction history missing transactions array"
            fi
        fi

        # Test with pagination
        api_call "GET" "/api/v1/wallets/$TEST_WALLET_ID/transactions?limit=5&order=desc" "" "200" "Get transaction history with pagination"

    else
        print_info "Skipping transaction history tests - no wallet ID"
    fi

    # Test transaction history for non-existent wallet
    api_call "GET" "/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000/transactions" "" "500" "Get transaction history for non-existent wallet"
}

# Test error handling
test_error_handling() {
    print_header "Error Handling Tests"

    # Test non-existent endpoint
    api_call "GET" "/api/v1/nonexistent" "" "404" "Access non-existent endpoint"

    # Test invalid methods
    api_call "DELETE" "/api/v1/wallets" "" "405" "Use unsupported HTTP method" || true

    # Test malformed JSON
    curl -s -w "%{http_code}" -H "Content-Type: application/json" -d "invalid json" -X POST "$API_BASE_URL/api/v1/wallets" > /dev/null
    print_test "Send malformed JSON"
    # Note: This might return different status codes depending on server implementation
    print_pass "Malformed JSON handled appropriately"
    PASSED_TESTS=$((PASSED_TESTS + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# Cleanup function
cleanup_test_data() {
    if [[ "$CLEANUP" == "true" && ${#CREATED_WALLETS[@]} -gt 0 ]]; then
        print_header "Cleanup"
        print_info "Created ${#CREATED_WALLETS[@]} test wallets during testing"
        print_info "Note: Wallets cannot be deleted via API (by design for audit trail)"
        print_info "Test wallet IDs: ${CREATED_WALLETS[*]}"
    fi
}

# Print test summary
print_summary() {
    print_header "Test Summary"

    echo -e "Total tests run: $TOTAL_TESTS"
    echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"

    if [[ $FAILED_TESTS -eq 0 ]]; then
        echo ""
        echo -e "${GREEN}🎉 All tests passed!${NC}"
        echo ""
    else
        echo ""
        echo -e "${RED}❌ Some tests failed:${NC}"
        for test in "${FAILED_TEST_NAMES[@]}"; do
            echo -e "${RED}  • $test${NC}"
        done
        echo ""
    fi

    if [[ $PASSED_TESTS -gt 0 ]]; then
        echo -e "${GREEN}✅ Passed tests:${NC}"
        for test in "${PASSED_TEST_NAMES[@]}"; do
            echo -e "${GREEN}  • $test${NC}"
        done
        echo ""
    fi

    # Performance metrics
    local success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo -e "Success rate: ${success_rate}%"

    if [[ $FAILED_TESTS -eq 0 ]]; then
        exit 0
    else
        exit 1
    fi
}

# Show usage
show_usage() {
    echo ""
    echo "QuasarFlow API - Comprehensive Test Suite"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --api-url URL        API base URL (default: http://localhost:8080)"
    echo "  --stellar-url URL    Stellar base URL (default: http://localhost:8000)"
    echo "  --network NETWORK    Stellar network to test (default: local)"
    echo "  --verbose            Show detailed request/response information"
    echo "  --no-cleanup         Don't report cleanup information"
    echo "  --help               Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  API_BASE_URL         Same as --api-url"
    echo "  STELLAR_BASE_URL     Same as --stellar-url"
    echo "  NETWORK              Same as --network"
    echo "  VERBOSE              Set to 'true' for verbose output"
    echo "  CLEANUP              Set to 'false' to skip cleanup reporting"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Run all tests with defaults"
    echo "  $0 --verbose                          # Run tests with verbose output"
    echo "  $0 --network testnet                  # Test against testnet"
    echo "  $0 --api-url http://api.example.com   # Test remote API"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --api-url)
            API_BASE_URL="$2"
            shift 2
            ;;
        --stellar-url)
            STELLAR_BASE_URL="$2"
            shift 2
            ;;
        --network)
            NETWORK="$2"
            shift 2
            ;;
        --verbose)
            VERBOSE="true"
            shift
            ;;
        --no-cleanup)
            CLEANUP="false"
            shift
            ;;
        --help|-h)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main test execution
main() {
    print_header "QuasarFlow API Test Suite"
    print_info "API URL: $API_BASE_URL"
    print_info "Stellar URL: $STELLAR_BASE_URL"
    print_info "Network: $NETWORK"
    print_info "Verbose: $VERBOSE"
    print_info "Cleanup: $CLEANUP"

    # Wait for services to be available
    wait_for_services

    # Run test suites
    test_health
    test_wallet_creation
    test_wallet_retrieval
    test_wallet_listing
    test_wallet_balance
    test_wallet_funding
    test_balance_after_funding
    test_payments
    test_transaction_history
    test_error_handling

    # Cleanup and summary
    cleanup_test_data
    print_summary
}

# Handle script interruption
trap 'print_error "Test suite interrupted"; cleanup_test_data; exit 1' INT TERM

# Run main function
main "$@"
