# Pawapay Go SDK

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.23-blue)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A comprehensive Go client library for integrating with the [Pawapay](https://www.pawapay.io/) mobile money API. This SDK enables seamless mobile money deposit operations across multiple African countries and mobile money operators.

## Features

- ✅ **Mobile Money Deposits** - Initiate deposits from customers across Africa
- ✅ **Multi-Provider Support** - Works with various mobile money operators (Vodacom, MTN, Airtel, Tigo, etc.)
- ✅ **Multi-Country Support** - Tanzania, Kenya, Rwanda, Nigeria, Cameroon, and more
- ✅ **Webhook Signature Validation** - Secure callback verification using RSA-PSS SHA-512
- ✅ **Debug Mode** - Built-in request/response logging for easy debugging
- ✅ **Type-Safe** - Comprehensive Go structs for all API models
- ✅ **Error Handling** - Detailed error responses with failure codes and messages
- ✅ **Production Ready** - Automatic fallback to production API URL

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Initialize Client](#initialize-client)
  - [Initiate Deposit](#initiate-deposit)
  - [Handle Callbacks](#handle-callbacks)
  - [Validate Webhook Signatures](#validate-webhook-signatures)
- [Supported Countries & Providers](#supported-countries--providers)
- [Debug Mode](#debug-mode)
- [Error Handling](#error-handling)
- [Examples](#examples)
- [Contributing](#contributing)
- [License](#license)

## Installation

```bash
go get github.com/salticon/pawapay-go-sdk
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/google/uuid"
    pawapay "github.com/salticon/pawapay-go-sdk"
)

func main() {
    // Initialize client
    client := pawapay.NewPawapayClient(&pawapay.ConfigOptions{
        ApiToken: "your-api-token",
        // InstanceURL is optional - defaults to https://api.pawapay.io
        // For sandbox: InstanceURL: "https://api.sandbox.pawapay.io"
    })

    // Create deposit request
    depositRequest := &pawapay.InitiateDepositRequestBody{
        DepositID: uuid.New().String(),
        Amount:    "1000",
        Currency:  pawapay.CURRENCY_CODE_TANZANIA,
        Payer: pawapay.Payer{
            Type: "MSISDN",
            AccountDetails: pawapay.AccountDetails{
                PhoneNumber: "255712345678",
                Provider:    pawapay.VODACOM_TZA,
            },
        },
        ClientReferenceID: "ORDER-12345",
        CustomerMessage:   "Payment for order #12345",
    }

    // Initiate deposit
    response, err := client.InitiateDeposit(depositRequest)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }

    fmt.Printf("Deposit Status: %s\n", response.Status)
}
```

## Configuration

### ConfigOptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ApiToken` | string | Yes | Your Pawapay API token |
| `InstanceURL` | string | No | API base URL (defaults to `https://api.pawapay.io`) |

### Environment Variables

For the example application, you can use a `.env` file:

```env
PAWAPAY_API_TOKEN=your-api-token-here
PAWAPAY_BASE_URL=https://api.sandbox.pawapay.io  # Optional, for sandbox
```

## Usage

### Initialize Client

**Production (default):**
```go
client := pawapay.NewPawapayClient(&pawapay.ConfigOptions{
    ApiToken: "your-api-token",
})
```

**Sandbox:**
```go
client := pawapay.NewPawapayClient(&pawapay.ConfigOptions{
    ApiToken:    "your-sandbox-token",
    InstanceURL: "https://api.sandbox.pawapay.io",
})
```

### Initiate Deposit

```go
depositRequest := &pawapay.InitiateDepositRequestBody{
    DepositID: uuid.New().String(),
    Amount:    "5000",
    Currency:  pawapay.CURRENCY_CODE_KENYA,
    Payer: pawapay.Payer{
        Type: "MSISDN",
        AccountDetails: pawapay.AccountDetails{
            PhoneNumber: "254712345678",
            Provider:    pawapay.MPESA_KEN,
        },
    },
    ClientReferenceID: "ORDER-67890",
    CustomerMessage:   "Payment for services",
    Metadata: []pawapay.MetadataItem{
        {"orderId": "12345"},
        {"customerId": "customer-001"},
    },
}

response, err := client.InitiateDeposit(depositRequest)
if err != nil {
    log.Printf("Error: %v", err)
    return
}

fmt.Printf("Deposit ID: %s\n", response.DepositID)
fmt.Printf("Status: %s\n", response.Status)
```

### Handle Callbacks

Pawapay sends webhook callbacks for deposit status updates:

```go
func handleCallback(w http.ResponseWriter, r *http.Request) {
    var callback pawapay.DepositCallbackRequestBody

    if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Validate signature (see next section)

    switch callback.Status {
    case "COMPLETED":
        fmt.Printf("Deposit %s completed!\n", callback.DepositID)
        // Update your database, fulfill order, etc.
    case "FAILED":
        fmt.Printf("Deposit %s failed: %s\n",
            callback.DepositID,
            callback.FailureReason.FailureMessage)
        // Handle failure
    }

    w.WriteHeader(http.StatusOK)
}
```

### Validate Webhook Signatures

```go
func validateWebhook(r *http.Request, publicKeyPEM string) bool {
    signature := r.Header.Get("X-Pawapay-Signature")

    bodyBytes, _ := io.ReadAll(r.Body)
    r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

    isValid := pawapay.ValidateSignature(publicKeyPEM, bodyBytes, signature)
    return isValid
}
```

## Supported Countries & Providers

The SDK includes constants for all supported mobile money operators:

### Tanzania
- `VODACOM_TZA` - Vodacom M-Pesa
- `AIRTEL_TZA` - Airtel Money
- `TIGO_TZA` - Tigo Pesa
- `HALOPESA_TZA` - Halo Pesa

### Kenya
- `MPESA_KEN` - M-Pesa
- `AIRTEL_KEN` - Airtel Money

### Rwanda
- `MTN_RWA` - MTN Mobile Money
- `AIRTEL_RWA` - Airtel Money

### Nigeria
- `MTN_NGA` - MTN Mobile Money
- `AIRTEL_NGA` - Airtel Money

### Cameroon
- `MTN_CMR` - MTN Mobile Money
- `ORANGE_CMR` - Orange Money

### Other Countries
- Zambia, Uganda, Benin, Ivory Coast, Ghana, Senegal, and more

**Currency Codes:**
```go
pawapay.CURRENCY_CODE_TANZANIA  // TZS
pawapay.CURRENCY_CODE_KENYA     // KES
pawapay.CURRENCY_CODE_RWANDA    // RWF
pawapay.CURRENCY_CODE_NIGERIA   // NGN
pawapay.CURRENCY_CODE_CAMEROON  // XAF
// ... and more
```

## Debug Mode

Enable debug mode to see detailed HTTP request/response logs:

```go
client := pawapay.NewPawapayClient(cfg)
client.Debug = true  // Enable debug logging
```

**Debug Output:**
```
========== DEBUG: REQUEST ==========
URL: https://api.pawapay.io/v2/deposits
Authorization: Bearer 12345678...abcd
Content-Type: application/json; charset=UTF-8
Body:
{"depositId":"...","amount":"1000",...}
====================================

========== DEBUG: RESPONSE ==========
Status: 200 OK
Body:
{"depositId":"...","status":"ACCEPTED",...}
=====================================
```

## Error Handling

The SDK provides detailed error handling for different scenarios:

### HTTP Errors (4xx, 5xx)
```go
response, err := client.InitiateDeposit(request)
if err != nil {
    // Error format: "pawapay API error (status 400): Bad Request - [message]"
    log.Printf("API Error: %v", err)
}
```

### Rejected Deposits
```go
response, err := client.InitiateDeposit(request)
if err != nil {
    // Error format: "deposit rejected: AUTHENTICATION_ERROR - The API token is invalid"
    log.Printf("Deposit Rejected: %v", err)
}
```

### Common Error Codes
- `AUTHENTICATION_ERROR` - Invalid API token
- `INSUFFICIENT_BALANCE` - Customer has insufficient funds
- `INVALID_MSISDN` - Invalid phone number
- `TRANSACTION_LIMIT_EXCEEDED` - Amount exceeds limits
- `DUPLICATE_TRANSACTION` - Duplicate deposit ID

## Examples

See the [example](./example) directory for a complete working example:

```bash
cd example
cp .env.example .env  # Create and configure your .env file
go run main.go
```

## API Reference

### Client Methods

#### `InitiateDeposit(payload *InitiateDepositRequestBody) (*RequestDepositResponse, error)`
Initiates a mobile money deposit request.

### Key Structs

#### `InitiateDepositRequestBody`
- `DepositID` (string) - Unique UUIDv4 identifier
- `Amount` (string) - Amount to collect (e.g., "1000")
- `Currency` (string) - ISO 4217 currency code
- `Payer` (Payer) - Customer details
- `ClientReferenceID` (string) - Your internal reference
- `CustomerMessage` (string) - Message shown to customer
- `Metadata` ([]MetadataItem) - Optional metadata

#### `RequestDepositResponse`
- `DepositID` (string) - The deposit identifier
- `Status` (string) - ACCEPTED, REJECTED, COMPLETED, FAILED
- `Created` (string) - Timestamp
- `FailureReason` (FailureReason) - Error details if rejected

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [Pawapay API Docs](https://docs.pawapay.io/)
- **Issues**: [GitHub Issues](https://github.com/salticon/pawapay-go-sdk/issues)
- **Email**: support@pawapay.io

## Acknowledgments

- Built with ❤️ for the African fintech ecosystem
- Powered by [Pawapay](https://www.pawapay.io/)

---

**Note**: This is an unofficial SDK. For official support, please contact Pawapay directly.
