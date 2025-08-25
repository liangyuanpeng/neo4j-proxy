package test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"neo4j-proxy/internal/auth"
)

var _ = Describe("Authentication and Tenant Extraction", func() {
	var authenticator *auth.Authenticator

	Describe("Username-based Tenant Extraction", func() {
		BeforeEach(func() {
			extractor := auth.NewUsernameBasedExtractor()
			authenticator = auth.New(extractor)
		})

		Context("when username contains tenant prefix", func() {
			It("should extract tenant ID from username with @ format", func() {
				tenantID, err := authenticator.AuthenticateAndRoute("tenant1@user", "password", map[string]interface{}{})
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("tenant1"))
			})

			It("should extract tenant ID from username with different tenant", func() {
				tenantID, err := authenticator.AuthenticateAndRoute("tenant2@admin", "password", map[string]interface{}{})
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("tenant2"))
			})

			It("should handle complex tenant IDs", func() {
				tenantID, err := authenticator.AuthenticateAndRoute("my-org-prod@developer", "password", map[string]interface{}{})
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("my-org-prod"))
			})
		})

		Context("when username does not contain tenant prefix", func() {
			It("should use tenant_id from metadata if available", func() {
				metadata := map[string]interface{}{
					"tenant_id": "metadata-tenant",
				}
				tenantID, err := authenticator.AuthenticateAndRoute("plainuser", "password", metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("metadata-tenant"))
			})

			It("should default to tenant1 when no tenant information available", func() {
				tenantID, err := authenticator.AuthenticateAndRoute("plainuser", "password", map[string]interface{}{})
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("tenant1"))
			})
		})

		Context("when username is empty", func() {
			It("should return an error", func() {
				_, err := authenticator.AuthenticateAndRoute("", "password", map[string]interface{}{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("username is required"))
			})
		})

		Context("when username has invalid format", func() {
			It("should handle username with @ but no tenant", func() {
				tenantID, err := authenticator.AuthenticateAndRoute("@user", "password", map[string]interface{}{})
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("tenant1")) // Should default
			})

			It("should handle username with multiple @ symbols", func() {
				tenantID, err := authenticator.AuthenticateAndRoute("tenant@user@domain", "password", map[string]interface{}{})
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("tenant"))
			})
		})
	})

	Describe("Database-based Tenant Extraction", func() {
		BeforeEach(func() {
			extractor := auth.NewDatabaseBasedExtractor()
			authenticator = auth.New(extractor)
		})

		Context("when database name is provided in metadata", func() {
			It("should map database names to tenant IDs", func() {
				metadata := map[string]interface{}{
					"database": "db1",
				}
				tenantID, err := authenticator.AuthenticateAndRoute("user", "password", metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("tenant1"))

				metadata["database"] = "db2"
				tenantID, err = authenticator.AuthenticateAndRoute("user", "password", metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("tenant2"))
			})

			It("should handle alternative database names", func() {
				metadata := map[string]interface{}{
					"database": "database1",
				}
				tenantID, err := authenticator.AuthenticateAndRoute("user", "password", metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("tenant1"))
			})

			It("should use database name directly if no mapping exists", func() {
				metadata := map[string]interface{}{
					"database": "custom-database",
				}
				tenantID, err := authenticator.AuthenticateAndRoute("user", "password", metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("custom-database"))
			})
		})

		Context("when no database name is provided", func() {
			It("should fall back to username-based extraction", func() {
				tenantID, err := authenticator.AuthenticateAndRoute("tenant2@user", "password", map[string]interface{}{})
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("tenant2"))
			})
		})
	})

	Describe("Authenticator Configuration", func() {
		Context("when creating authenticator with nil extractor", func() {
			It("should use default username-based extractor", func() {
				auth := auth.New(nil)
				Expect(auth).NotTo(BeNil())
				
				tenantID, err := auth.AuthenticateAndRoute("tenant1@user", "password", map[string]interface{}{})
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("tenant1"))
			})
		})
	})
})