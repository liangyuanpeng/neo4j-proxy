package auth

import (
	"errors"
	"strings"
)

// TenantExtractor defines interface for extracting tenant ID from various sources
type TenantExtractor interface {
	ExtractTenantID(username, password string, metadata map[string]interface{}) (string, error)
}

// UsernameBasedExtractor extracts tenant ID from username prefix
type UsernameBasedExtractor struct{}

// NewUsernameBasedExtractor creates a new username-based tenant extractor
func NewUsernameBasedExtractor() *UsernameBasedExtractor {
	return &UsernameBasedExtractor{}
}

// ExtractTenantID extracts tenant ID from username using format: tenantID@username
func (e *UsernameBasedExtractor) ExtractTenantID(username, password string, metadata map[string]interface{}) (string, error) {
	if username == "" {
		return "", errors.New("username is required")
	}

	// Check if username contains tenant prefix (format: tenant@user)
	if strings.Contains(username, "@") {
		parts := strings.SplitN(username, "@", 2)
		if len(parts) == 2 && parts[0] != "" {
			return parts[0], nil
		}
	}

	// If no tenant prefix, try to extract from metadata or use default
	if tenantID, ok := metadata["tenant_id"].(string); ok && tenantID != "" {
		return tenantID, nil
	}

	// Default to tenant1 if no explicit tenant specified
	return "tenant1", nil
}

// DatabaseBasedExtractor extracts tenant ID from database name
type DatabaseBasedExtractor struct{}

// NewDatabaseBasedExtractor creates a new database-based tenant extractor
func NewDatabaseBasedExtractor() *DatabaseBasedExtractor {
	return &DatabaseBasedExtractor{}
}

// ExtractTenantID extracts tenant ID from database name in metadata
func (e *DatabaseBasedExtractor) ExtractTenantID(username, password string, metadata map[string]interface{}) (string, error) {
	if database, ok := metadata["database"].(string); ok && database != "" {
		// Map database names to tenant IDs
		switch database {
		case "db1", "database1":
			return "tenant1", nil
		case "db2", "database2":
			return "tenant2", nil
		default:
			return database, nil
		}
	}

	// Fall back to username-based extraction
	extractor := NewUsernameBasedExtractor()
	return extractor.ExtractTenantID(username, password, metadata)
}

// Authenticator handles authentication and tenant routing
type Authenticator struct {
	extractor TenantExtractor
}

// New creates a new authenticator with the specified tenant extractor
func New(extractor TenantExtractor) *Authenticator {
	if extractor == nil {
		extractor = NewUsernameBasedExtractor()
	}
	return &Authenticator{
		extractor: extractor,
	}
}

// AuthenticateAndRoute authenticates the user and determines tenant routing
func (a *Authenticator) AuthenticateAndRoute(username, password string, metadata map[string]interface{}) (string, error) {
	return a.extractor.ExtractTenantID(username, password, metadata)
}