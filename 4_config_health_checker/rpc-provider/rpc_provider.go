package rpc_provider

// RpcProviderAuthType defines the different types of authentication for RPC providers
type RpcProviderAuthType string

const (
	NoAuth    RpcProviderAuthType = "no-auth"    // No authentication
	BasicAuth RpcProviderAuthType = "basic-auth" // HTTP Header "Authorization: Basic base64(username:password)"
	TokenAuth RpcProviderAuthType = "token-auth" // URL Token-based authentication "https://api.example.com/YOUR_TOKEN"
)

// RpcProvider represents an RPC provider configuration with various options
type RpcProvider struct {
	Name    string `json:"name" validate:"required,min=1"` // Provider name for identification
	URL     string `json:"url" validate:"required,url"`    // Current Provider URL
	Enabled bool   `json:"enabled"`                        // Whether the provider is enabled
	// Authentication
	AuthType     RpcProviderAuthType `json:"authType" validate:"required,oneof=no-auth basic-auth token-auth"` // Type of authentication
	AuthLogin    string              `json:"authLogin" validate:"omitempty,min=1"`                             // Login for BasicAuth (empty string if not used)
	AuthPassword string              `json:"authPassword" validate:"omitempty,min=1"`                          // Password for BasicAuth (empty string if not used)
	AuthToken    string              `json:"authToken" validate:"omitempty,min=1"`                             // Token for TokenAuth (empty string if not used)
}
