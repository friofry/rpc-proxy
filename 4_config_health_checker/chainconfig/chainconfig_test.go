package chainconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadChains(t *testing.T) {
	// Create temporary test file
	content := `{
		"chains": [
			{
				"name": "ethereum",
				"network": "mainnet",
				"providers": [
					{
						"name": "infura",
						"url": "https://mainnet.infura.io/v3",
						"authType": "token-auth",
						"authToken": "test",
						"enabled": true
					}
				]
			}
		]
	}`

	tmpFile, err := os.CreateTemp("", "test-config-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	tmpFile.Close()

	t.Run("successful load", func(t *testing.T) {
		chains, err := LoadChains(tmpFile.Name())
		assert.NoError(t, err)
		assert.Len(t, chains, 1)
		assert.Equal(t, "ethereum", chains[0].Name)
		assert.Equal(t, "mainnet", chains[0].Network)
		assert.Len(t, chains[0].Providers, 1)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadChains("nonexistent.json")
		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		invalidFile, err := os.CreateTemp("", "invalid-*.json")
		assert.NoError(t, err)
		defer os.Remove(invalidFile.Name())

		_, err = invalidFile.WriteString("{invalid}")
		assert.NoError(t, err)
		invalidFile.Close()

		_, err = LoadChains(invalidFile.Name())
		assert.Error(t, err)
	})
}

func TestGetChainByNameAndNetwork(t *testing.T) {
	chains := []ChainConfig{
		{
			Name:    "ethereum",
			Network: "mainnet",
		},
		{
			Name:    "ethereum",
			Network: "sepolia",
		},
	}

	t.Run("found", func(t *testing.T) {
		chain, err := GetChainByNameAndNetwork(chains, "ethereum", "mainnet")
		assert.NoError(t, err)
		assert.Equal(t, "mainnet", chain.Network)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := GetChainByNameAndNetwork(chains, "unknown", "testnet")
		assert.Error(t, err)
	})
}

func TestGetEnabledProviders(t *testing.T) {
	chain := ChainConfig{
		Providers: []ProviderConfig{
			{Enabled: true},
			{Enabled: false},
			{Enabled: true},
		},
	}

	enabled := chain.GetEnabledProviders()
	assert.Len(t, enabled, 2)
}

func TestValidateChainConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  ChainConfig
		wantErr bool
	}{
		{
			name: "valid",
			config: ChainConfig{
				Name:    "ethereum",
				Network: "mainnet",
				Providers: []ProviderConfig{
					{Enabled: true},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: ChainConfig{
				Network: "mainnet",
				Providers: []ProviderConfig{
					{Enabled: true},
				},
			},
			wantErr: true,
		},
		{
			name: "missing network",
			config: ChainConfig{
				Name: "ethereum",
				Providers: []ProviderConfig{
					{Enabled: true},
				},
			},
			wantErr: true,
		},
		{
			name: "no providers",
			config: ChainConfig{
				Name:    "ethereum",
				Network: "mainnet",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChainConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
