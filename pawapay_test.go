package pawapaygo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetActiveConfiguration tests the GetActiveConfiguration method
func TestGetActiveConfiguration(t *testing.T) {
	// Create a mock response
	mockResponse := ActiveConfigurationResponse{
		CompanyName: "Test Merchant Inc.",
		SignatureConfiguration: SignatureConfiguration{
			SignedRequestsOnly: true,
			SignedCallbacks:    true,
		},
		Countries: []CountryConfig{
			{
				Country: "ZMB",
				DisplayName: map[string]string{
					"en": "Zambia",
				},
				Prefix: "260",
				Flag:   "https://cdn.example.com/zmb_flag.png",
				Providers: []ProviderConfig{
					{
						Provider:                "MTN_MOMO_ZMB",
						DisplayName:             "MTN",
						Logo:                    "https://cdn.example.com/mtn_logo.png",
						NameDisplayedToCustomer: "Test Merchant Inc.",
						Currencies: []CurrencyConfig{
							{
								Currency:    "ZMW",
								DisplayName: "ZMW",
								OperationTypes: map[string]OperationType{
									"DEPOSIT": {
										AuthType:            "PROVIDER_AUTH",
										PinPrompt:           "AUTOMATIC",
										PinPromptRevivable:  true,
										MinTransactionLimit: "1",
										MaxTransactionLimit: "100000",
										DecimalsInAmount:    "NONE",
										Status:              "OPERATIONAL",
										CallbackURL:         "https://merchant.com/depositCallback",
									},
									"PAYOUT": {
										MinTransactionLimit: "1",
										MaxTransactionLimit: "100000",
										DecimalsInAmount:    "NONE",
										Status:              "OPERATIONAL",
										CallbackURL:         "https://merchant.com/payoutCallback",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/v2/active-conf" {
			t.Errorf("Expected path /v2/active-conf, got %s", r.URL.Path)
		}

		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("Authorization header is missing")
		}

		// Send mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewPawapayClient(&ConfigOptions{
		InstanceURL: server.URL,
		ApiToken:    "test-token-12345678",
	})

	// Call the method
	response, err := client.GetActiveConfiguration()
	if err != nil {
		t.Fatalf("GetActiveConfiguration failed: %v", err)
	}

	// Verify response
	if response.CompanyName != mockResponse.CompanyName {
		t.Errorf("Expected company name %s, got %s", mockResponse.CompanyName, response.CompanyName)
	}

	if response.SignatureConfiguration.SignedRequestsOnly != mockResponse.SignatureConfiguration.SignedRequestsOnly {
		t.Error("SignedRequestsOnly mismatch")
	}

	if len(response.Countries) != 1 {
		t.Errorf("Expected 1 country, got %d", len(response.Countries))
	}

	if response.Countries[0].Country != "ZMB" {
		t.Errorf("Expected country ZMB, got %s", response.Countries[0].Country)
	}

	if len(response.Countries[0].Providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(response.Countries[0].Providers))
	}

	provider := response.Countries[0].Providers[0]
	if provider.Provider != "MTN_MOMO_ZMB" {
		t.Errorf("Expected provider MTN_MOMO_ZMB, got %s", provider.Provider)
	}

	if len(provider.Currencies) != 1 {
		t.Errorf("Expected 1 currency, got %d", len(provider.Currencies))
	}

	currency := provider.Currencies[0]
	if currency.Currency != "ZMW" {
		t.Errorf("Expected currency ZMW, got %s", currency.Currency)
	}

	depositOp, exists := currency.OperationTypes["DEPOSIT"]
	if !exists {
		t.Error("DEPOSIT operation type not found")
	}

	if depositOp.Status != "OPERATIONAL" {
		t.Errorf("Expected DEPOSIT status OPERATIONAL, got %s", depositOp.Status)
	}

	if depositOp.AuthType != "PROVIDER_AUTH" {
		t.Errorf("Expected AuthType PROVIDER_AUTH, got %s", depositOp.AuthType)
	}
}

// TestGetActiveConfiguration_ErrorResponse tests error handling
func TestGetActiveConfiguration_ErrorResponse(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Timestamp: "2025-12-28T10:00:00Z",
			Status:    401,
			Error:     "Unauthorized",
			Message:   "Invalid API token",
			Path:      "/v2/active-conf",
		})
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewPawapayClient(&ConfigOptions{
		InstanceURL: server.URL,
		ApiToken:    "invalid-token",
	})

	// Call the method
	_, err := client.GetActiveConfiguration()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify error message contains expected information
	expectedMsg := "pawapay API error"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

// TestGetActiveConfiguration_NetworkError tests network error handling
func TestGetActiveConfiguration_NetworkError(t *testing.T) {
	// Create client with invalid URL
	client := NewPawapayClient(&ConfigOptions{
		InstanceURL: "http://invalid-url-that-does-not-exist.local",
		ApiToken:    "test-token",
	})

	// Call the method
	_, err := client.GetActiveConfiguration()
	if err == nil {
		t.Fatal("Expected network error, got nil")
	}
}

// TestGetActiveConfiguration_WithDebug tests debug mode
func TestGetActiveConfiguration_WithDebug(t *testing.T) {
	// Create a simple test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ActiveConfigurationResponse{
			CompanyName: "Debug Test",
			SignatureConfiguration: SignatureConfiguration{
				SignedRequestsOnly: false,
				SignedCallbacks:    false,
			},
			Countries: []CountryConfig{},
		})
	}))
	defer server.Close()

	// Create client with debug enabled
	client := NewPawapayClient(&ConfigOptions{
		InstanceURL: server.URL,
		ApiToken:    "test-token-12345678",
	})
	client.Debug = true

	// Call the method (debug output will be printed to console)
	response, err := client.GetActiveConfiguration()
	if err != nil {
		t.Fatalf("GetActiveConfiguration failed: %v", err)
	}

	if response.CompanyName != "Debug Test" {
		t.Errorf("Expected company name 'Debug Test', got %s", response.CompanyName)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestGetDepositStatus tests the GetDepositStatus method with FOUND status
func TestGetDepositStatus(t *testing.T) {
	depositID := "8917c345-4791-4285-a416-62f24b6982db"

	// Create a mock response
	mockResponse := CheckDepositStatusResponse{
		Status: "FOUND",
		Data: &DepositData{
			DepositID: depositID,
			Status:    "COMPLETED",
			Amount:    "123.00",
			Currency:  "ZMW",
			Country:   "ZMB",
			Payer: PayerDetails{
				Type: "MMO",
				AccountDetails: PayerAccountDetails{
					PhoneNumber: "260763456789",
					Provider:    "MTN_MOMO_ZMB",
				},
			},
			CustomerMessage:       "To ACME company",
			ClientReferenceID:     "REF-987654321",
			Created:               "2020-10-19T08:17:01Z",
			ProviderTransactionID: "12356789",
			Metadata: []MetadataItem{
				{"orderId": "ORD-123456789"},
				{"customerId": "[email protected]"},
			},
		},
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		expectedPath := "/v2/deposits/" + depositID
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("Authorization header is missing")
		}

		// Send mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewPawapayClient(&ConfigOptions{
		InstanceURL: server.URL,
		ApiToken:    "test-token-12345678",
	})

	// Call the method
	response, err := client.GetDepositStatus(depositID)
	if err != nil {
		t.Fatalf("GetDepositStatus failed: %v", err)
	}

	// Verify response
	if response.Status != "FOUND" {
		t.Errorf("Expected status FOUND, got %s", response.Status)
	}

	if response.Data == nil {
		t.Fatal("Expected data to be present")
	}

	if response.Data.DepositID != depositID {
		t.Errorf("Expected depositId %s, got %s", depositID, response.Data.DepositID)
	}

	if response.Data.Status != "COMPLETED" {
		t.Errorf("Expected status COMPLETED, got %s", response.Data.Status)
	}

	if response.Data.Amount != "123.00" {
		t.Errorf("Expected amount 123.00, got %s", response.Data.Amount)
	}

	if response.Data.Currency != "ZMW" {
		t.Errorf("Expected currency ZMW, got %s", response.Data.Currency)
	}

	if response.Data.Payer.AccountDetails.PhoneNumber != "260763456789" {
		t.Errorf("Expected phone number 260763456789, got %s", response.Data.Payer.AccountDetails.PhoneNumber)
	}
}

// TestGetDepositStatus_NotFound tests the GetDepositStatus method with NOT_FOUND status
func TestGetDepositStatus_NotFound(t *testing.T) {
	depositID := "00000000-0000-0000-0000-000000000000"

	// Create a mock response
	mockResponse := CheckDepositStatusResponse{
		Status: "NOT_FOUND",
		Data:   nil,
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewPawapayClient(&ConfigOptions{
		InstanceURL: server.URL,
		ApiToken:    "test-token-12345678",
	})

	// Call the method
	response, err := client.GetDepositStatus(depositID)
	if err != nil {
		t.Fatalf("GetDepositStatus failed: %v", err)
	}

	// Verify response
	if response.Status != "NOT_FOUND" {
		t.Errorf("Expected status NOT_FOUND, got %s", response.Status)
	}

	if response.Data != nil {
		t.Error("Expected data to be nil for NOT_FOUND status")
	}
}

// TestGetDepositStatus_EmptyDepositID tests validation for empty depositID
func TestGetDepositStatus_EmptyDepositID(t *testing.T) {
	client := NewPawapayClient(&ConfigOptions{
		InstanceURL: "http://localhost",
		ApiToken:    "test-token",
	})

	// Call the method with empty depositID
	_, err := client.GetDepositStatus("")
	if err == nil {
		t.Fatal("Expected error for empty depositID, got nil")
	}

	expectedMsg := "depositID is required"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
	}
}

// TestGetDepositStatus_ErrorResponse tests error handling
func TestGetDepositStatus_ErrorResponse(t *testing.T) {
	depositID := "8917c345-4791-4285-a416-62f24b6982db"

	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Timestamp: "2025-12-28T10:00:00Z",
			Status:    404,
			Error:     "Not Found",
			Message:   "Deposit not found",
			Path:      "/v2/deposits/" + depositID,
		})
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewPawapayClient(&ConfigOptions{
		InstanceURL: server.URL,
		ApiToken:    "test-token",
	})

	// Call the method
	_, err := client.GetDepositStatus(depositID)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify error message contains expected information
	expectedMsg := "pawapay API error"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

// TestGetDepositStatus_WithDebug tests debug mode
func TestGetDepositStatus_WithDebug(t *testing.T) {
	depositID := "8917c345-4791-4285-a416-62f24b6982db"

	// Create a simple test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CheckDepositStatusResponse{
			Status: "FOUND",
			Data: &DepositData{
				DepositID: depositID,
				Status:    "COMPLETED",
				Amount:    "100.00",
				Currency:  "ZMW",
				Country:   "ZMB",
				Payer: PayerDetails{
					Type: "MMO",
					AccountDetails: PayerAccountDetails{
						PhoneNumber: "260123456789",
						Provider:    "MTN_MOMO_ZMB",
					},
				},
				Created: "2025-12-28T10:00:00Z",
			},
		})
	}))
	defer server.Close()

	// Create client with debug enabled
	client := NewPawapayClient(&ConfigOptions{
		InstanceURL: server.URL,
		ApiToken:    "test-token-12345678",
	})
	client.Debug = true

	// Call the method (debug output will be printed to console)
	response, err := client.GetDepositStatus(depositID)
	if err != nil {
		t.Fatalf("GetDepositStatus failed: %v", err)
	}

	if response.Status != "FOUND" {
		t.Errorf("Expected status FOUND, got %s", response.Status)
	}

	if response.Data.DepositID != depositID {
		t.Errorf("Expected depositId %s, got %s", depositID, response.Data.DepositID)
	}
}

// TestPredictProvider tests the PredictProvider method with a successful response
func TestPredictProvider(t *testing.T) {
	// Create a mock response
	mockResponse := PredictProviderResponse{
		Country:     "ZMB",
		Provider:    "MTN_MOMO_ZMB",
		PhoneNumber: "260763456789",
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify request path
		if r.URL.Path != "/v2/predict-provider" {
			t.Errorf("Expected path /v2/predict-provider, got %s", r.URL.Path)
		}

		// Verify Content-Type header
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Verify Authorization header
		if r.Header.Get("Authorization") == "" {
			t.Error("Expected Authorization header to be set")
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewPawapayClient(&ConfigOptions{
		InstanceURL: server.URL,
		ApiToken:    "test-token",
	})

	// Call PredictProvider
	phoneNumber := "+260 763-456789"
	response, err := client.PredictProvider(phoneNumber)
	if err != nil {
		t.Fatalf("PredictProvider failed: %v", err)
	}

	// Verify response
	if response.Country != "ZMB" {
		t.Errorf("Expected country ZMB, got %s", response.Country)
	}

	if response.Provider != "MTN_MOMO_ZMB" {
		t.Errorf("Expected provider MTN_MOMO_ZMB, got %s", response.Provider)
	}

	if response.PhoneNumber != "260763456789" {
		t.Errorf("Expected phoneNumber 260763456789, got %s", response.PhoneNumber)
	}
}

// TestPredictProvider_EmptyPhoneNumber tests validation for empty phone number
func TestPredictProvider_EmptyPhoneNumber(t *testing.T) {
	client := NewPawapayClient(&ConfigOptions{
		InstanceURL: "https://api.sandbox.pawapay.io",
		ApiToken:    "test-token",
	})

	_, err := client.PredictProvider("")
	if err == nil {
		t.Error("Expected error for empty phone number, got nil")
	}

	if err.Error() != "phoneNumber is required" {
		t.Errorf("Expected error 'phoneNumber is required', got '%s'", err.Error())
	}
}

// TestPredictProvider_ErrorResponse tests handling of HTTP error responses
func TestPredictProvider_ErrorResponse(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  400,
			Error:   "Bad Request",
			Message: "Invalid phone number format",
		})
	}))
	defer server.Close()

	client := NewPawapayClient(&ConfigOptions{
		InstanceURL: server.URL,
		ApiToken:    "test-token",
	})

	_, err := client.PredictProvider("+123456789")
	if err == nil {
		t.Error("Expected error for bad request, got nil")
	}

	// Check that error message contains status code
	if !contains(err.Error(), "400") {
		t.Errorf("Expected error to contain status code 400, got: %s", err.Error())
	}
}

// TestPredictProvider_WithDebug tests debug mode output
func TestPredictProvider_WithDebug(t *testing.T) {
	// Create a mock response
	mockResponse := PredictProviderResponse{
		Country:     "KEN",
		Provider:    "MPESA_KEN",
		PhoneNumber: "254712345678",
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with debug enabled
	client := NewPawapayClient(&ConfigOptions{
		InstanceURL: server.URL,
		ApiToken:    "test-token-1234567890",
	})
	client.Debug = true

	// Call PredictProvider
	response, err := client.PredictProvider("+254 712-345678")
	if err != nil {
		t.Fatalf("PredictProvider failed: %v", err)
	}

	if response.Country != "KEN" {
		t.Errorf("Expected country KEN, got %s", response.Country)
	}
}

// TestPredictProvider_DifferentCountries tests provider prediction for different countries
func TestPredictProvider_DifferentCountries(t *testing.T) {
	testCases := []struct {
		name        string
		phoneNumber string
		country     string
		provider    string
	}{
		{
			name:        "Zambia MTN",
			phoneNumber: "+260763456789",
			country:     "ZMB",
			provider:    "MTN_MOMO_ZMB",
		},
		{
			name:        "Kenya M-Pesa",
			phoneNumber: "+254712345678",
			country:     "KEN",
			provider:    "MPESA_KEN",
		},
		{
			name:        "Ghana MTN",
			phoneNumber: "+233244123456",
			country:     "GHA",
			provider:    "MTN_MOMO_GHA",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mockResponse := PredictProviderResponse{
					Country:     tc.country,
					Provider:    tc.provider,
					PhoneNumber: tc.phoneNumber,
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(mockResponse)
			}))
			defer server.Close()

			client := NewPawapayClient(&ConfigOptions{
				InstanceURL: server.URL,
				ApiToken:    "test-token",
			})

			response, err := client.PredictProvider(tc.phoneNumber)
			if err != nil {
				t.Fatalf("PredictProvider failed: %v", err)
			}

			if response.Country != tc.country {
				t.Errorf("Expected country %s, got %s", tc.country, response.Country)
			}

			if response.Provider != tc.provider {
				t.Errorf("Expected provider %s, got %s", tc.provider, response.Provider)
			}
		})
	}
}
