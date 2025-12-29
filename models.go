package pawapaygo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type ConfigOptions struct {
	InstanceURL string
	ApiToken    string
}

type DepositCallbackRequestBody struct {
	// A UUIDv4 based ID specified by you, that uniquely identifies the deposit.
	DepositID string `json:"depositId"`

	// The final status of the payment.
	// Available options: COMPLETED, FAILED
	Status string `json:"status"`

	/**
	* The amount to be collected (deposit) or disbursed (payout or refund).

	* Amount must follow below requirements or the request will be rejected:

	* Between zero and two decimal places can be supplied, depending on what the specific MMO supports. Learn about all MMO supported decimal places.

	* The minimum and maximum amount depends on the limits of the specific MMO. You can find them from the Active Configuration endpoint.
	* Leading zeroes are not permitted except where the value is less than 1. For any value less than one, one and only one leading zero must be supplied.

	* Trailing zeroes are permitted.

	* Valid examples: 5, 5.0, 5.00, 5.5, 5.55, 5555555, 0.5

	* Not valid examples: 5., 5.555, 5555555555555555555, .5, -5.5, 00.5, 00.00, 00001.32

	* Required string length: 1 - 23

	* Example: "15"

	**/
	RequestedAmount string `json:"requestedAmount"`

	/**
	* The currency in which the amount is specified.

	* Format must be the ISO 4217 three character currency code in upper case. Read more from Wikipedia.

	* You can find all the supported currencies that the specific correspondent supports from here.

	* The active configuration endpoint provides the list of correspondents configured for your account together with the currencies.

	* Example: "ZMW"
	**/
	Currency string `json:"currency"`

	/**
	* The country in which the MMO operates.

	* Format is ISO 3166-1 alpha-3, three character country code in upper case. Read more from Wikipedia.

	* Example: "ZMB"
	**/
	Country string `json:"country"`

	Correspondent string `json:"correspondent"`
	Payer         struct {
		Type    string `json:"type"`
		Address struct {
			Value string `json:"value"`
		} `json:"address"`
	} `json:"payer"`
	CustomerTimestamp time.Time `json:"customerTimestamp"`

	StatementDescription string `json:"statementDescription"`
	Created              string `json:"created"`
	DepositedAmount      string `json:"depositedAmount"`
	RespondedByPayer     string `json:"respondedByPayer"`
	CorrespondentIDs     struct {
		MTNInit  string `json:"MTN_INIT"`
		MTNFinal string `json:"MTN_FINAL"`
	} `json:"correspondentIds"`
	SuspiciousActivityReport []struct {
		ActivityType string `json:"activityType"`
		Comment      string `json:"comment"`
	} `json:"suspiciousActivityReport"`
	FailureReason *struct {
		FailureCode    string `json:"failureCode"`
		FailureMessage string `json:"failureMessage"`
	} `json:"failureReason,omitempty"`
	Metadata struct {
		OrderID    string `json:"orderId"`
		CustomerID string `json:"customerId"`
	} `json:"metadata"`
}

func (i *DepositCallbackRequestBody) ToBytes() (*bytes.Reader, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// Request Deposit request body
type InitiateDepositRequestBody struct {
	DepositID            string         `json:"depositId"`
	Payer                Payer          `json:"payer"`
	PreAuthorisationCode string         `json:"preAuthorisationCode"`
	ClientReferenceID    string         `json:"clientReferenceId"`
	CustomerMessage      string         `json:"customerMessage"`
	Amount               string         `json:"amount"`
	Currency             string         `json:"currency"`
	Metadata             []MetadataItem `json:"metadata"`
}

func (i *InitiateDepositRequestBody) ToBytes() (*bytes.Reader, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

type Payer struct {
	Type           string         `json:"type"`
	AccountDetails AccountDetails `json:"accountDetails"`
}

type AccountDetails struct {
	PhoneNumber string `json:"phoneNumber"`
	Provider    string `json:"provider"`
}

type MetadataItem map[string]any

// Request Deposit response object
type RequestDepositResponse struct {
	DepositID     string        `json:"depositId"`
	Status        string        `json:"status"`
	Created       string        `json:"created"`
	FailureReason FailureReason `json:"failureReason"`
}

// HTTP Error response from Pawapay API (for 4xx, 5xx errors)
type ErrorResponse struct {
	Timestamp string `json:"timestamp"`
	Status    int    `json:"status"`
	Error     string `json:"error"`
	Message   string `json:"message"`
	Path      string `json:"path"`
}

func (e *ErrorResponse) ToError() error {
	return fmt.Errorf("pawapay API error (status %d): %s - %s", e.Status, e.Error, e.Message)
}

func (r *RequestDepositResponse) DecodeBytes(b io.Reader) error {
	resBytes, err := io.ReadAll(b)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(resBytes, r); err != nil {
		return err
	}
	return nil
}

type FailureReason struct {
	FailureCode    string `json:"failureCode"`
	FailureMessage string `json:"failureMessage"`
}

// WalletBalance represents a single wallet balance
type WalletBalance struct {
	Country  string `json:"country"`  // ISO 3166-1 alpha-3 country code (e.g., "ZMB", "UGA")
	Balance  string `json:"balance"`  // Current balance as a string (e.g., "21798.03")
	Currency string `json:"currency"` // ISO 4217 currency code (e.g., "ZMW", "UGX")
	Provider string `json:"provider"` // Mobile money provider (may be empty)
}

// WalletBalancesResponse represents the response from the wallet balances API
type WalletBalancesResponse struct {
	Balances []WalletBalance `json:"balances"`
}

// ActiveConfigurationResponse represents the response from the active configuration API
type ActiveConfigurationResponse struct {
	CompanyName            string                 `json:"companyName"`
	SignatureConfiguration SignatureConfiguration `json:"signatureConfiguration"`
	Countries              []CountryConfig        `json:"countries"`
}

// SignatureConfiguration represents signature settings
type SignatureConfiguration struct {
	SignedRequestsOnly bool `json:"signedRequestsOnly"`
	SignedCallbacks    bool `json:"signedCallbacks"`
}

// CountryConfig represents a country configuration
type CountryConfig struct {
	Country     string            `json:"country"`     // ISO 3166-1 alpha-3 country code
	DisplayName map[string]string `json:"displayName"` // Localized country names
	Prefix      string            `json:"prefix"`      // Phone number prefix
	Flag        string            `json:"flag"`        // URL to flag image
	Providers   []ProviderConfig  `json:"providers"`   // List of providers in this country
}

// ProviderConfig represents a mobile money provider configuration
type ProviderConfig struct {
	Provider                string           `json:"provider"`                // Provider code (e.g., "MTN_MOMO_BEN")
	DisplayName             string           `json:"displayName"`             // Display name (e.g., "MTN")
	Logo                    string           `json:"logo"`                    // URL to provider logo
	NameDisplayedToCustomer string           `json:"nameDisplayedToCustomer"` // Name shown to customer
	Currencies              []CurrencyConfig `json:"currencies"`              // Supported currencies
}

// CurrencyConfig represents a currency configuration for a provider
type CurrencyConfig struct {
	Currency       string                   `json:"currency"`       // ISO 4217 currency code
	DisplayName    string                   `json:"displayName"`    // Display name for currency
	OperationTypes map[string]OperationType `json:"operationTypes"` // Supported operation types
}

// OperationType represents configuration for a specific operation type
type OperationType struct {
	// Common fields for all operation types
	CallbackURL string `json:"callbackUrl,omitempty"`

	// Fields for DEPOSIT operations
	AuthType              string                 `json:"authType,omitempty"`
	PinPrompt             string                 `json:"pinPrompt,omitempty"`
	PinPromptRevivable    bool                   `json:"pinPromptRevivable,omitempty"`
	PinPromptInstructions *PinPromptInstructions `json:"pinPromptInstructions,omitempty"`

	// Fields for transactional operations (DEPOSIT, PAYOUT, REFUND, REMITTANCE)
	MinTransactionLimit string `json:"minTransactionLimit,omitempty"`
	MaxTransactionLimit string `json:"maxTransactionLimit,omitempty"`
	DecimalsInAmount    string `json:"decimalsInAmount,omitempty"`
	Status              string `json:"status,omitempty"` // OPERATIONAL, DELAYED, CLOSED
}

// PinPromptInstructions represents instructions for PIN prompt revival
type PinPromptInstructions struct {
	Channels []Channel `json:"channels"`
}

// Channel represents a communication channel for PIN prompt instructions
type Channel struct {
	Type         string                   `json:"type"`         // e.g., "USSD"
	DisplayName  map[string]string        `json:"displayName"`  // Localized display names
	QuickLink    string                   `json:"quickLink"`    // Quick link for the channel
	Variables    map[string]string        `json:"variables"`    // Variables for instructions
	Instructions map[string][]Instruction `json:"instructions"` // Localized instructions
}

// Instruction represents a single instruction step
type Instruction struct {
	Text     string `json:"text"`
	Template string `json:"template"`
}

// CheckDepositStatusResponse represents the response from checking deposit status
type CheckDepositStatusResponse struct {
	Status string       `json:"status"` // FOUND or NOT_FOUND
	Data   *DepositData `json:"data,omitempty"`
}

// DepositData represents the detailed deposit information
type DepositData struct {
	DepositID             string         `json:"depositId"`
	Status                string         `json:"status"` // SUBMITTED, ACCEPTED, COMPLETED, FAILED, REJECTED, ENQUEUED
	Amount                string         `json:"amount"`
	Currency              string         `json:"currency"`
	Country               string         `json:"country"`
	Payer                 PayerDetails   `json:"payer"`
	CustomerMessage       string         `json:"customerMessage,omitempty"`
	ClientReferenceID     string         `json:"clientReferenceId,omitempty"`
	Created               string         `json:"created"`
	ProviderTransactionID string         `json:"providerTransactionId,omitempty"`
	Metadata              []MetadataItem `json:"metadata,omitempty"`
	FailureReason         *FailureReason `json:"failureReason,omitempty"`
}

// PayerDetails represents payer information in deposit status
type PayerDetails struct {
	Type           string              `json:"type"` // MMO (Mobile Money Operator)
	AccountDetails PayerAccountDetails `json:"accountDetails"`
}

// PayerAccountDetails represents payer account information
type PayerAccountDetails struct {
	PhoneNumber string `json:"phoneNumber"`
	Provider    string `json:"provider"`
}

// PredictProviderRequest represents the request body for predicting a provider
type PredictProviderRequest struct {
	PhoneNumber string `json:"phoneNumber"` // The phone number (MSISDN) to predict the provider of
}

// PredictProviderResponse represents the response from the predict provider API
type PredictProviderResponse struct {
	Country     string `json:"country"`     // ISO 3166-1 alpha-3 country code
	Provider    string `json:"provider"`    // Mobile money provider
	PhoneNumber string `json:"phoneNumber"` // Correctly formatted phone number
}
