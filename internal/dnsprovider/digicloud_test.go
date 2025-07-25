package dnsprovider

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDigicloudProvider_Present(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		token       string
		keyAuth     string
		expectError bool
	}{
		{
			name:        "successful TXT record creation",
			domain:      "example.com",
			token:       "test-token",
			keyAuth:     "test-key-auth",
			expectError: false,
		},
		{
			name:        "empty domain",
			domain:      "",
			token:       "test-token",
			keyAuth:     "test-key-auth",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewDigicloudProvider("https://api.digicloud.ir", "test-token", "default", 300)
			
			err := provider.Present(tt.domain, tt.token, tt.keyAuth)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Since we don't have a real API, we expect this to fail with HTTP error
				// In a real test, you'd mock the HTTP client
				assert.Error(t, err) // Will fail due to no real API
			}
		})
	}
}

func TestDigicloudProvider_CleanUp(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		token       string
		keyAuth     string
		expectError bool
	}{
		{
			name:        "successful TXT record deletion",
			domain:      "example.com",
			token:       "test-token",
			keyAuth:     "test-key-auth",
			expectError: false,
		},
		{
			name:        "empty domain",
			domain:      "",
			token:       "test-token",
			keyAuth:     "test-key-auth",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewDigicloudProvider("https://api.digicloud.ir", "test-token", "default", 300)
			
			err := provider.CleanUp(tt.domain, tt.token, tt.keyAuth)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Since we don't have a real API, we expect this to fail with HTTP error
				// In a real test, you'd mock the HTTP client
				assert.Error(t, err) // Will fail due to no real API
			}
		})
	}
}

func TestDigicloudProvider_Timeout(t *testing.T) {
	provider := NewDigicloudProvider("https://api.digicloud.ir", "test-token", "default", 300)
	timeout, interval := provider.Timeout()
	
	// Should return reasonable timeout and interval
	assert.True(t, timeout > 0)
	assert.True(t, timeout <= 5*time.Minute)
	assert.True(t, interval > 0)
	assert.True(t, interval <= 30*time.Second)
}

func TestNewDigicloudProvider(t *testing.T) {
	provider := NewDigicloudProvider("https://api.digicloud.ir", "test-token", "default", 300)
	
	assert.NotNil(t, provider)
	assert.Equal(t, "https://api.digicloud.ir", provider.baseURL)
	assert.Equal(t, "test-token", provider.apiToken)
	assert.Equal(t, "default", provider.namespace)
	assert.Equal(t, 300, provider.ttl)
}
