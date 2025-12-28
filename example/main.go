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
	pawapayDepositExample()
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
