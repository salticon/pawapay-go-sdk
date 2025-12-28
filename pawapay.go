package pawapaygo

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	hs "github.com/thinkgos/http-signature-go"
)

type Client struct {
	instanceURL string
	authToken   string
	Debug       bool
}

var _ PawapayAPIClient = (*Client)(nil)

const defaultBaseURL = "https://api.pawapay.io"

func NewPawapayClient(cfg *ConfigOptions) *Client {
	baseURL := cfg.InstanceURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		instanceURL: baseURL,
		authToken:   cfg.ApiToken,
	}
}

type PawapayAPIClient interface {
	InitiateDeposit(*InitiateDepositRequestBody) (*RequestDepositResponse, error)
}

func (a *Client) InitiateDeposit(payload *InitiateDepositRequestBody) (*RequestDepositResponse, error) {

	// Initialize an http client
	httpc := http.Client{}

	requestBody, err := payload.ToBytes()
	if err != nil {
		fmt.Println("Error converting request body to bytes\n", err)
		return nil, err
	}

	// Build the URL, ensuring no double slashes
	baseURL := strings.TrimSuffix(a.instanceURL, "/")
	url := baseURL + "/v2" + requestDepositRoute

	// Create an http request
	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		fmt.Println("Error creating new request body\n", err)
		return nil, err
	}

	// Add required http headers
	req.Header.Set("Authorization", "Bearer "+a.authToken)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// Debug logging for request body
	if a.Debug {
		fmt.Println("\n========== DEBUG: REQUEST ==========")
		fmt.Printf("URL: %s\n", url)

		// Mask the token for security (show first 8 chars only)
		maskedToken := a.authToken
		if len(maskedToken) > 8 {
			maskedToken = maskedToken[:8] + "..." + maskedToken[len(maskedToken)-4:]
		}
		fmt.Printf("Authorization: Bearer %s\n", maskedToken)
		fmt.Printf("Content-Type: %s\n", req.Header.Get("Content-Type"))

		fmt.Println("Body:")
		// Read the request body for logging
		if req.Body != nil {
			bodyBytes, _ := io.ReadAll(req.Body)
			fmt.Println(string(bodyBytes))
			// Restore the body for the actual request
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		fmt.Println("====================================")
	}

	res, err := httpc.Do(req)
	if err != nil {
		fmt.Println("Error making an http request to pawapay\n", err)
		return nil, err
	}
	// Close request body stream in the end
	defer res.Body.Close()

	// Read response body
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body\n", err)
		return nil, err
	}

	// Debug logging for response body
	if a.Debug {
		fmt.Println("\n========== DEBUG: RESPONSE ==========")
		fmt.Printf("Status: %d %s\n", res.StatusCode, res.Status)
		fmt.Println("Body:")
		fmt.Println(string(resBody))
		fmt.Println("=====================================")
	}

	// Parse the response body
	body := &RequestDepositResponse{}
	if err := body.DecodeBytes(bytes.NewReader(resBody)); err != nil {
		// If we can't parse as RequestDepositResponse, try parsing as HTTP error
		if res.StatusCode >= 400 {
			errResp := &ErrorResponse{}
			if err := json.Unmarshal(resBody, errResp); err != nil {
				// If we can't parse the error response, return a generic error
				return nil, fmt.Errorf("HTTP %d: %s", res.StatusCode, string(resBody))
			}
			return nil, errResp.ToError()
		}
		fmt.Println("Error parsing the response body to go struct\n", err)
		return nil, err
	}

	// Check if the response indicates a rejection with failure reason
	if body.Status == "REJECTED" && body.FailureReason.FailureCode != "" {
		return nil, fmt.Errorf("deposit rejected: %s - %s", body.FailureReason.FailureCode, body.FailureReason.FailureMessage)
	}

	return body, nil
}

func ValidateSignature(r *http.Request, keyId string, privateKey string) bool {

	parser := hs.NewParser(
		hs.WithMinimumRequiredHeaders([]string{
			hs.Date,
			hs.Digest,
			hs.HeaderSignature,
			hs.Host,
		}),
		hs.WithSigningMethods(
			hs.SigningMethodRsaPssSha512.Alg(),
			func() hs.SigningMethod { return hs.SigningMethodRsaPssSha512 },
		),
		hs.WithSigningMethods(
			hs.SigningMethodRsaPssSha256.Alg(),
			func() hs.SigningMethod { return hs.SigningMethodRsaPssSha256 },
		),
		// hs.WithValidators(
		// 	hs.NewDigestUsingSharedValidator(),
		// 	hs.NewDateValidator(),
		// ),
		// hs.WithKeystone(keyStone),
	)
	err := parser.AddMetadata(
		hs.KeyId(keyId),
		hs.Metadata{
			Alg:    hs.SigningMethodRsaPssSha512.Name,
			Key:    []byte(""),
			Scheme: hs.SchemeSignature,
		})
	if err != nil {
		fmt.Println(err)
		return false
	}

	gotParam, err := parser.ParseFromRequest(r)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if err := parser.Verify(r, gotParam); err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

type Component struct {
	Name       string
	Parameters map[string]string
}

type SignatureParams struct {
	Components []Component
	Alg        string
	Created    int64
	KeyID      string
}

// CreateContentDigestHeader generates the SHA-512 content-digest
func CreateContentDigestHeader(body []byte) string {
	sum := sha512.Sum512(body)
	digest := base64.StdEncoding.EncodeToString(sum[:])
	return fmt.Sprintf("sha-512=:%s:", digest)
}

// CreateSignatureBase returns both the signature base string and Signature-Input header value
func CreateSignatureBase(req *http.Request, body []byte, sigParams SignatureParams) (signatureBase, signatureInput string, err error) {
	seen := make(map[string]bool)
	var sb strings.Builder
	var inputNames []string

	// Calculate Content-Digest header and add it if not present
	if req.Header.Get("Content-Digest") == "" {
		digest := CreateContentDigestHeader(body)
		req.Header.Set("Content-Digest", digest)
	}

	for _, comp := range sigParams.Components {
		identifier := serializeComponentIdentifier(comp)
		if seen[identifier] {
			return "", "", fmt.Errorf("duplicate component identifier: %s", identifier)
		}
		seen[identifier] = true
		inputNames = append(inputNames, fmt.Sprintf(`"%s"`, comp.Name))

		// Build signature base line
		sb.WriteString(identifier)
		sb.WriteString(": ")

		value, err := getComponentValue(req, comp)
		if err != nil {
			return "", "", fmt.Errorf("failed to get value for %s: %v", comp.Name, err)
		}
		sb.WriteString(value)
		sb.WriteString("\n")
	}

	// Build final signature-params line
	sigParamsLine := fmt.Sprintf(`"@signature-params": (%s);alg=%s;created=%d;keyid="%s"`,
		strings.Join(inputNames, " "), sigParams.Alg, sigParams.Created, sigParams.KeyID)

	sb.WriteString(sigParamsLine)

	return sb.String(), sigParamsLine, nil
}

func serializeComponentIdentifier(comp Component) string {
	id := fmt.Sprintf(`"%s"`, comp.Name)
	for k, v := range comp.Parameters {
		if v == "" {
			id += ";" + k
		} else {
			id += fmt.Sprintf(`;%s="%s"`, k, v)
		}
	}
	return id
}

func getComponentValue(req *http.Request, comp Component) (string, error) {
	if strings.HasPrefix(comp.Name, "@") {
		switch comp.Name {
		case "@method":
			return req.Method, nil
		case "@authority":
			return req.Host, nil
		case "@path":
			return req.URL.Path, nil
		default:
			return "", fmt.Errorf("unsupported derived component: %s", comp.Name)
		}
	} else {
		values := req.Header[http.CanonicalHeaderKey(comp.Name)]
		if len(values) == 0 {
			return "", fmt.Errorf("header %s not found", comp.Name)
		}
		return strings.Join(values, ", "), nil
	}
}
