package dnsprovider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"k8s.io/klog/v2"
)

// DigicloudProvider implements the DNS provider for Digicloud Edge DNS API
type DigicloudProvider struct {
	client      *http.Client
	baseURL     string
	apiToken    string
	namespace   string
	ttl         int
	httpTimeout time.Duration
}

// NewDigicloudProvider creates a new Digicloud DNS provider
func NewDigicloudProvider(baseURL, apiToken, namespace string, ttl int) *DigicloudProvider {
	if baseURL == "" {
		baseURL = "https://api.digicloud.ir"
	}
	if ttl == 0 {
		ttl = 300 // Default TTL of 5 minutes
	}

	return &DigicloudProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		apiToken:    apiToken,
		namespace:   namespace,
		ttl:         ttl,
		httpTimeout: 30 * time.Second,
	}
}

// DNSTXTRecord represents a TXT record for the Digicloud API
type DNSTXTRecord struct {
	Name    string `json:"name"`
	TTL     string `json:"ttl"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Note    string `json:"note,omitempty"`
}

// DNSTXTRecordDetails represents a TXT record with ID returned from the API
type DNSTXTRecordDetails struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	TTL     string `json:"ttl"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Note    string `json:"note,omitempty"`
}

// DNSRecordListResponse represents the response when listing DNS records
type DNSRecordListResponse struct {
	Records []DNSTXTRecordDetails `json:"records"`
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (p *DigicloudProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	klog.V(2).Infof("Creating TXT record for domain %s with value %s", info.EffectiveFQDN, info.Value)

	// Extract the domain name from the FQDN
	domainName := p.extractDomainName(info.EffectiveFQDN)
	if domainName == "" {
		return fmt.Errorf("could not extract domain name from %s", info.EffectiveFQDN)
	}

	// Get domain ID
	domainID, err := p.getDomainID(domainName)
	if err != nil {
		return fmt.Errorf("failed to get domain ID for %s: %w", domainName, err)
	}

	// Extract record name (subdomain part)
	recordName := p.extractRecordName(info.EffectiveFQDN, domainName)

	// Create the TXT record
	record := DNSTXTRecord{
		Name:    recordName,
		TTL:     fmt.Sprintf("%ds", p.ttl),
		Type:    "TXT",
		Content: info.Value,
		Note:    "Created by cert-manager digicloud issuer",
	}

	err = p.createTXTRecord(domainID, record)
	if err != nil {
		return fmt.Errorf("failed to create TXT record: %w", err)
	}

	klog.V(2).Infof("Successfully created TXT record for %s", info.EffectiveFQDN)
	return nil
}

// CleanUp removes the TXT record after the challenge is complete
func (p *DigicloudProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	klog.V(2).Infof("Cleaning up TXT record for domain %s", info.EffectiveFQDN)

	// Extract the domain name from the FQDN
	domainName := p.extractDomainName(info.EffectiveFQDN)
	if domainName == "" {
		return fmt.Errorf("could not extract domain name from %s", info.EffectiveFQDN)
	}

	// Get domain ID
	domainID, err := p.getDomainID(domainName)
	if err != nil {
		return fmt.Errorf("failed to get domain ID for %s: %w", domainName, err)
	}

	// Extract record name (subdomain part)
	recordName := p.extractRecordName(info.EffectiveFQDN, domainName)

	// Find and delete the TXT record
	recordID, err := p.findTXTRecord(domainID, recordName, info.Value)
	if err != nil {
		return fmt.Errorf("failed to find TXT record: %w", err)
	}

	if recordID != "" {
		err = p.deleteTXTRecord(domainID, recordID)
		if err != nil {
			return fmt.Errorf("failed to delete TXT record: %w", err)
		}
		klog.V(2).Infof("Successfully deleted TXT record for %s", info.EffectiveFQDN)
	} else {
		klog.V(2).Infof("TXT record not found for %s, may have been already deleted", info.EffectiveFQDN)
	}

	return nil
}

// Timeout returns the timeout for DNS propagation
func (p *DigicloudProvider) Timeout() (timeout, interval time.Duration) {
	return 5 * time.Minute, 10 * time.Second
}

// extractDomainName extracts the domain name from the FQDN
// For example: _acme-challenge.sub.example.com -> example.com
func (p *DigicloudProvider) extractDomainName(fqdn string) string {
	fqdn = strings.TrimSuffix(fqdn, ".")
	parts := strings.Split(fqdn, ".")

	// We need to find the actual domain (not subdomain)
	// This is a simplified approach - for production, you might want to use
	// a more sophisticated domain detection algorithm or maintain a list of known domains
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return ""
}

// extractRecordName extracts the record name from the FQDN
// For example: _acme-challenge.sub.example.com with domain example.com -> _acme-challenge.sub
func (p *DigicloudProvider) extractRecordName(fqdn, domain string) string {
	fqdn = strings.TrimSuffix(fqdn, ".")
	domain = strings.TrimSuffix(domain, ".")

	if fqdn == domain {
		return "@"
	}

	if strings.HasSuffix(fqdn, "."+domain) {
		return strings.TrimSuffix(fqdn, "."+domain)
	}

	return fqdn
}

// getDomainID gets the domain ID from the domain name
// For now, we'll assume the domain name is the ID - this might need adjustment based on the actual API
func (p *DigicloudProvider) getDomainID(domainName string) (string, error) {
	// In the Digicloud API, it appears the domain_name_id is used in the path
	// This might be the domain name itself or an actual ID
	// For now, we'll use the domain name as the ID
	return domainName, nil
}

// createTXTRecord creates a TXT record via the Digicloud API
func (p *DigicloudProvider) createTXTRecord(domainID string, record DNSTXTRecord) error {
	url := fmt.Sprintf("%s/v1/edge/domains/%s/records", p.baseURL, domainID)

	jsonData, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	req.Header.Set("Digicloud-Namespace", p.namespace)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// findTXTRecord finds a TXT record by name and content
func (p *DigicloudProvider) findTXTRecord(domainID, recordName, content string) (string, error) {
	url := fmt.Sprintf("%s/v1/edge/domains/%s/records", p.baseURL, domainID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	req.Header.Set("Digicloud-Namespace", p.namespace)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var recordList DNSRecordListResponse
	if err := json.NewDecoder(resp.Body).Decode(&recordList); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Find the TXT record with matching name and content
	for _, record := range recordList.Records {
		if record.Type == "TXT" && record.Name == recordName && record.Content == content {
			return record.ID, nil
		}
	}

	return "", nil // Record not found
}

// deleteTXTRecord deletes a TXT record by ID
func (p *DigicloudProvider) deleteTXTRecord(domainID, recordID string) error {
	url := fmt.Sprintf("%s/v1/edge/domains/%s/records/%s", p.baseURL, domainID, recordID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	req.Header.Set("Digicloud-Namespace", p.namespace)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// NewDigicloudDNSProviderFromIssuerAndSecretData creates a DNS provider from issuer spec and secret data
// This function signature is expected by the cert-manager issuer-lib
func NewDigicloudDNSProviderFromIssuerAndSecretData(issuerSpec interface{}, secretData map[string][]byte) (interface{}, error) {
	// TODO: Implement proper type assertion and DNS provider creation
	// For now, return a basic provider
	apiToken := string(secretData["token"])
	namespace := string(secretData["namespace"])
	if namespace == "" {
		namespace = "default"
	}

	provider := NewDigicloudProvider("", apiToken, namespace, 300)
	return provider, nil
}
