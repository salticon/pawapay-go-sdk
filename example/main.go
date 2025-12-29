package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	pawapay "github.com/salticon/pawapay-go-sdk"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Uncomment the example you want to run:
	pawapayDepositExample()
	// getWalletBalancesExample()
	// getActiveConfigurationExample()
	// getDepositStatusExample()
	// predictProviderExample()
}

// Uncomment below to run the web server example instead
/*
func mainWebServer() {
	godotenv.Load()

	cfg := &pawapay.ConfigOptions{
		InstanceURL: os.Getenv("BASE_URL"),
		ApiToken:    os.Getenv("AUTH_TOKEN"),
	}
	fmt.Println("CONFIG Variables\n", cfg)
	client := pawapay.NewPawapayClient(cfg)

	router := gin.Default()

	router.POST("/initiate-deposit", func(c *gin.Context) {
		reqBody := &pawapay.InitiateDepositRequestBody{
			DepositID:            uuid.New().String(),
			Amount:               "100",
			Currency:             pawapay.CURRENCY_CODE_CAMEROON,
			PreAuthorisationCode: "54366",
			ClientReferenceID:    "REF-45343",
			CustomerMessage:      "Testing the api",
			Payer: pawapay.Payer{
				Type: "MMO",
				AccountDetails: pawapay.AccountDetails{
					PhoneNumber: "237653456789",
					Provider:    pawapay.MTN_MOMO_CMR,
				},
			},
		}
		res, err := client.InitiateDeposit(reqBody)
		fmt.Println(err)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, res)
	})
	router.POST("/deposit-callback", func(c *gin.Context) {
		// fmt.Println(c.Request)
		body := &pawapay.DepositCallbackRequestBody{}
		if err := c.ShouldBindBodyWith(body, binding.JSON); err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		validated := pawapay.ValidateSignature(c.Request, "keyId", os.Getenv("PrivateKey"))
		fmt.Println(validated)
		c.JSON(http.StatusOK, gin.H{
			"validated": validated,
		})

		// Parse signature input parameters
		for _, part := range strings.Split(c.Request.Header.Get("Signature-Input"), ";") {
			kv := strings.Split(part, "=")
			if len(kv) >= 2 {
				fmt.Printf("Key: %s, Value: %s\n", kv[0], kv[1])
			}
		}

		sigParams := pawapay.SignatureParams{
			Components: []pawapay.Component{
				{Name: "@method"},
				{Name: "@authority"},
				{Name: "@path"},
				{Name: "content-digest"},
				{Name: "content-type"},
			},
			Created: time.Now().Unix(),
			KeyID:   os.Getenv("KEY_ID"),
			Alg:     "sha-512",
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			fmt.Println("Fail to marshal struct")
		}
		signatureBase, _, err := pawapay.CreateSignatureBase(c.Request, bodyBytes, sigParams)

		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": err,
			})
		}
		fmt.Println(signatureBase)
		fmt.Println("")
	})
	router.Run()
}
*/

func pawapayDepositExample() {
	fmt.Println("=== Pawapay-Go Deposit Example ===")

	// Initialize the Pawapay client with configuration
	cfg := &pawapay.ConfigOptions{
		InstanceURL: os.Getenv("PAWAPAY_BASE_URL"),  // e.g., "https://api.sandbox.pawapay.io" for sandbox
		ApiToken:    os.Getenv("PAWAPAY_API_TOKEN"), // Your Pawapay API token
	}

	client := pawapay.NewPawapayClient(cfg)

	// Enable debug mode to log HTTP requests and responses
	client.Debug = true

	// Generate a unique deposit ID (UUIDv4)
	depositID := uuid.New().String()

	// Create the deposit request
	depositRequest := &pawapay.InitiateDepositRequestBody{
		DepositID: depositID,
		Amount:    "1000",                       // Amount in the currency's smallest unit or as string
		Currency:  pawapay.CURRENCY_CODE_RWANDA, // "TZS"
		Payer: pawapay.Payer{
			Type: "MMO",
			AccountDetails: pawapay.AccountDetails{
				PhoneNumber: "250722345678",     // Customer's phone number in international format
				Provider:    pawapay.AIRTEL_RWA, // Mobile money provider (e.g., VODACOM_TZA, AIRTEL_TZA, TIGO_TZA)
			},
		},
		ClientReferenceID:    fmt.Sprintf("ORDER-%d", time.Now().Unix()), // Your internal reference
		CustomerMessage:      "",                                         // Message shown to customer
		PreAuthorisationCode: "random1",                                  // Optional: if using pre-authorization
		Metadata:             []pawapay.MetadataItem{},
	}

	// Initiate the deposit
	response, err := client.InitiateDeposit(depositRequest)
	if err != nil {
		log.Printf("Error initiating deposit: %v", err)
		return
	}

	// Handle the response
	fmt.Printf("Deposit initiated successfully!\n")
	fmt.Printf("  Deposit ID: %s\n", response.DepositID)
	fmt.Printf("  Status: %s\n", response.Status)
	fmt.Printf("  Created: %s\n", response.Created)

	if response.FailureReason.FailureCode != "" {
		fmt.Printf("  Failure Code: %s\n", response.FailureReason.FailureCode)
		fmt.Printf("  Failure Message: %s\n", response.FailureReason.FailureMessage)
	}

	// The deposit status will be one of:
	// - ACCEPTED: The deposit request was accepted and is being processed
	// - REJECTED: The deposit request was rejected
	// - DUPLICATE_IGNORED: A deposit with the same ID already exists
	switch response.Status {
	case "ACCEPTED":
		fmt.Println("✅ Deposit accepted! Waiting for customer to confirm on their phone.")
	case "REJECTED":
		fmt.Printf("❌ Deposit rejected: %s\n", response.FailureReason.FailureMessage)
	case "DUPLICATE_IGNORED":
		fmt.Println("⚠️ Duplicate deposit ignored - this deposit ID was already used.")
	}
}

// Example function to get wallet balances
func getWalletBalancesExample() {
	fmt.Println("\n=== Pawapay-Go Wallet Balances Example ===")

	// Initialize the Pawapay client with configuration
	cfg := &pawapay.ConfigOptions{
		InstanceURL: os.Getenv("PAWAPAY_BASE_URL"),  // e.g., "https://api.sandbox.pawapay.io" for sandbox
		ApiToken:    os.Getenv("PAWAPAY_API_TOKEN"), // Your Pawapay API token
	}

	client := pawapay.NewPawapayClient(cfg)

	// Enable debug mode to log HTTP requests and responses
	client.Debug = true

	// Get wallet balances
	response, err := client.GetWalletBalances()
	if err != nil {
		log.Printf("Error getting wallet balances: %v", err)
		return
	}

	// Display the balances
	fmt.Printf("\n📊 Wallet Balances:\n")
	fmt.Printf("Found %d wallet(s)\n\n", len(response.Balances))

	for i, balance := range response.Balances {
		fmt.Printf("Wallet %d:\n", i+1)
		fmt.Printf("  Country:  %s\n", balance.Country)
		fmt.Printf("  Currency: %s\n", balance.Currency)
		fmt.Printf("  Balance:  %s\n", balance.Balance)
		if balance.Provider != "" {
			fmt.Printf("  Provider: %s\n", balance.Provider)
		}
		fmt.Println()
	}
}

// Example function to get active configuration
func getActiveConfigurationExample() {
	fmt.Println("\n=== Pawapay-Go Active Configuration Example ===")

	// Initialize the Pawapay client with configuration
	cfg := &pawapay.ConfigOptions{
		InstanceURL: os.Getenv("PAWAPAY_BASE_URL"),  // e.g., "https://api.sandbox.pawapay.io" for sandbox
		ApiToken:    os.Getenv("PAWAPAY_API_TOKEN"), // Your Pawapay API token
	}

	client := pawapay.NewPawapayClient(cfg)

	// Enable debug mode to log HTTP requests and responses
	client.Debug = true

	// Get active configuration
	response, err := client.GetActiveConfiguration()
	if err != nil {
		log.Printf("Error getting active configuration: %v", err)
		return
	}

	// Display the configuration
	fmt.Printf("\n📋 Active Configuration:\n")
	fmt.Printf("Company Name: %s\n", response.CompanyName)
	fmt.Printf("Signed Requests Only: %v\n", response.SignatureConfiguration.SignedRequestsOnly)
	fmt.Printf("Signed Callbacks: %v\n\n", response.SignatureConfiguration.SignedCallbacks)

	fmt.Printf("Countries: %d\n\n", len(response.Countries))

	// Display each country and its providers
	for _, country := range response.Countries {
		fmt.Printf("🌍 Country: %s (%s)\n", country.DisplayName["en"], country.Country)
		fmt.Printf("   Phone Prefix: +%s\n", country.Prefix)
		fmt.Printf("   Providers: %d\n", len(country.Providers))

		for _, provider := range country.Providers {
			fmt.Printf("\n   📱 Provider: %s (%s)\n", provider.DisplayName, provider.Provider)
			fmt.Printf("      Name shown to customer: %s\n", provider.NameDisplayedToCustomer)

			for _, currency := range provider.Currencies {
				fmt.Printf("\n      💰 Currency: %s\n", currency.Currency)

				// Display operation types
				for opType, opConfig := range currency.OperationTypes {
					fmt.Printf("         • %s:\n", opType)
					if opConfig.Status != "" {
						fmt.Printf("           Status: %s\n", opConfig.Status)
					}
					if opConfig.MinTransactionLimit != "" {
						fmt.Printf("           Min: %s, Max: %s %s\n",
							opConfig.MinTransactionLimit,
							opConfig.MaxTransactionLimit,
							currency.Currency)
					}
					if opConfig.AuthType != "" {
						fmt.Printf("           Auth Type: %s\n", opConfig.AuthType)
					}
					if opConfig.CallbackURL != "" {
						fmt.Printf("           Callback URL: %s\n", opConfig.CallbackURL)
					}
				}
			}
		}
		fmt.Println()
	}
}

// Example function to check deposit status
func getDepositStatusExample() {
	fmt.Println("\n=== Pawapay-Go Check Deposit Status Example ===")

	// Initialize the Pawapay client with configuration
	cfg := &pawapay.ConfigOptions{
		InstanceURL: os.Getenv("PAWAPAY_BASE_URL"),  // e.g., "https://api.sandbox.pawapay.io" for sandbox
		ApiToken:    os.Getenv("PAWAPAY_API_TOKEN"), // Your Pawapay API token
	}

	client := pawapay.NewPawapayClient(cfg)

	// Enable debug mode to log HTTP requests and responses
	client.Debug = true

	// The depositId from a previous deposit request
	// Replace this with an actual depositId from your system
	depositID := "8917c345-4791-4285-a416-62f24b6982db"

	// Check deposit status
	response, err := client.GetDepositStatus(depositID)
	if err != nil {
		log.Printf("Error checking deposit status: %v", err)
		return
	}

	// Display the status
	fmt.Printf("\n📊 Deposit Status Check:\n")
	fmt.Printf("Status: %s\n", response.Status)

	if response.Status == "FOUND" && response.Data != nil {
		fmt.Printf("\n✅ Deposit Found:\n")
		fmt.Printf("Deposit ID: %s\n", response.Data.DepositID)
		fmt.Printf("Status: %s\n", response.Data.Status)
		fmt.Printf("Amount: %s %s\n", response.Data.Amount, response.Data.Currency)
		fmt.Printf("Country: %s\n", response.Data.Country)
		fmt.Printf("Created: %s\n", response.Data.Created)

		// Display payer information
		fmt.Printf("\n👤 Payer Information:\n")
		fmt.Printf("Type: %s\n", response.Data.Payer.Type)
		fmt.Printf("Phone Number: %s\n", response.Data.Payer.AccountDetails.PhoneNumber)
		fmt.Printf("Provider: %s\n", response.Data.Payer.AccountDetails.Provider)

		// Display optional fields if present
		if response.Data.CustomerMessage != "" {
			fmt.Printf("\n💬 Customer Message: %s\n", response.Data.CustomerMessage)
		}

		if response.Data.ClientReferenceID != "" {
			fmt.Printf("📝 Client Reference ID: %s\n", response.Data.ClientReferenceID)
		}

		if response.Data.ProviderTransactionID != "" {
			fmt.Printf("🔖 Provider Transaction ID: %s\n", response.Data.ProviderTransactionID)
		}

		// Display metadata if present
		if len(response.Data.Metadata) > 0 {
			fmt.Printf("\n📋 Metadata:\n")
			for _, meta := range response.Data.Metadata {
				for key, value := range meta {
					fmt.Printf("  %s: %v\n", key, value)
				}
			}
		}

		// Display failure reason if present
		if response.Data.FailureReason != nil {
			fmt.Printf("\n❌ Failure Reason:\n")
			fmt.Printf("Code: %s\n", response.Data.FailureReason.FailureCode)
			fmt.Printf("Message: %s\n", response.Data.FailureReason.FailureMessage)
		}

		// Status interpretation
		fmt.Printf("\n📌 Status Interpretation:\n")
		switch response.Data.Status {
		case "SUBMITTED":
			fmt.Println("The deposit has been submitted to the provider")
		case "ACCEPTED":
			fmt.Println("The deposit has been accepted by the provider")
		case "COMPLETED":
			fmt.Println("✅ The deposit has been completed successfully")
		case "FAILED":
			fmt.Println("❌ The deposit has failed")
		case "REJECTED":
			fmt.Println("❌ The deposit was rejected")
		case "ENQUEUED":
			fmt.Println("The deposit is enqueued and waiting to be processed")
		default:
			fmt.Printf("Unknown status: %s\n", response.Data.Status)
		}
	} else if response.Status == "NOT_FOUND" {
		fmt.Printf("\n❌ Deposit Not Found\n")
		fmt.Printf("The deposit with ID %s was not found in the system.\n", depositID)
	}
}

// Example function to predict mobile money provider from phone number
func predictProviderExample() {
	fmt.Println("\n=== Pawapay-Go Predict Provider Example ===")

	// Initialize the Pawapay client with configuration
	cfg := &pawapay.ConfigOptions{
		InstanceURL: os.Getenv("PAWAPAY_BASE_URL"),
		ApiToken:    os.Getenv("PAWAPAY_API_TOKEN"),
	}

	client := pawapay.NewPawapayClient(cfg)

	// Enable debug mode to see request/response details
	client.Debug = true

	// Test phone numbers from different countries
	testPhoneNumbers := []struct {
		number      string
		description string
	}{
		{"+260 763-456789", "Zambia MTN"},
		{"+254 712-345678", "Kenya M-Pesa"},
		{"+233 244-123456", "Ghana MTN"},
		{"+256 700-123456", "Uganda MTN"},
		{"+234 803-123456", "Nigeria MTN"},
	}

	fmt.Println("\n📱 Testing Provider Prediction for Multiple Phone Numbers:")
	fmt.Println("=" + string(make([]byte, 60)))

	for i, test := range testPhoneNumbers {
		fmt.Printf("\n[%d] Testing: %s (%s)\n", i+1, test.number, test.description)
		fmt.Println("-" + string(make([]byte, 60)))

		response, err := client.PredictProvider(test.number)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		// Display the prediction results
		fmt.Printf("\n✅ Provider Prediction Successful!\n")
		fmt.Printf("  📍 Country:      %s\n", response.Country)
		fmt.Printf("  🏢 Provider:     %s\n", response.Provider)
		fmt.Printf("  📞 Phone Number: %s (formatted)\n", response.PhoneNumber)

		// Provide additional context based on provider
		fmt.Printf("\n  ℹ️  Information:\n")
		switch response.Provider {
		case "MTN_MOMO_ZMB":
			fmt.Printf("     - Mobile Money Operator: MTN Mobile Money Zambia\n")
			fmt.Printf("     - Currency: ZMW (Zambian Kwacha)\n")
		case "MPESA_KEN":
			fmt.Printf("     - Mobile Money Operator: M-Pesa Kenya\n")
			fmt.Printf("     - Currency: KES (Kenyan Shilling)\n")
		case "MTN_MOMO_GHA":
			fmt.Printf("     - Mobile Money Operator: MTN Mobile Money Ghana\n")
			fmt.Printf("     - Currency: GHS (Ghanaian Cedi)\n")
		case "MTN_MOMO_UGA":
			fmt.Printf("     - Mobile Money Operator: MTN Mobile Money Uganda\n")
			fmt.Printf("     - Currency: UGX (Ugandan Shilling)\n")
		case "MTN_MOMO_NGA":
			fmt.Printf("     - Mobile Money Operator: MTN Mobile Money Nigeria\n")
			fmt.Printf("     - Currency: NGN (Nigerian Naira)\n")
		default:
			fmt.Printf("     - Provider: %s\n", response.Provider)
		}
	}

	// Example: Using prediction result to initiate a deposit
	fmt.Println("\n\n💡 Use Case: Using Prediction for Deposit Initiation")
	fmt.Println("=" + string(make([]byte, 60)))

	phoneNumber := "+260763456789"
	fmt.Printf("\nStep 1: Predict provider for phone number: %s\n", phoneNumber)

	prediction, err := client.PredictProvider(phoneNumber)
	if err != nil {
		log.Fatalf("Failed to predict provider: %v", err)
	}

	fmt.Printf("  ✓ Predicted Provider: %s\n", prediction.Provider)
	fmt.Printf("  ✓ Country: %s\n", prediction.Country)

	fmt.Println("\nStep 2: Use predicted provider to initiate deposit")
	fmt.Printf("  → You can now use '%s' as the provider in your deposit request\n", prediction.Provider)
	fmt.Printf("  → This ensures the correct mobile money operator is used\n")

	// Example deposit payload (not executed)
	fmt.Println("\nExample Deposit Payload:")
	fmt.Printf(`  {
    "depositId": "%s",
    "amount": "100.00",
    "currency": "ZMW",
    "country": "%s",
    "payer": {
      "type": "MMO",
      "accountDetails": {
        "phoneNumber": "%s",
        "provider": "%s"
      }
    }
  }
`, uuid.New().String(), prediction.Country, prediction.PhoneNumber, prediction.Provider)

	fmt.Println("\n✅ Provider Prediction Example Complete!")
	fmt.Println("\nNote: Average misprediction rate is 0.12%")
	fmt.Println("      (Benin has higher rate of 6% due to number portability)")
}
