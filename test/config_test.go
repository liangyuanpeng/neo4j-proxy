package test

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"neo4j-proxy/pkg/config"
)

var _ = Describe("Configuration Management", func() {
	var tempConfigFile string

	AfterEach(func() {
		if tempConfigFile != "" {
			os.Remove(tempConfigFile)
			tempConfigFile = ""
		}
		os.Unsetenv("CONFIG_FILE")
	})

	Describe("Default Configuration", func() {
		Context("when no config file is specified", func() {
			It("should load default configuration", func() {
				cfg, err := config.Load()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).NotTo(BeNil())
				
				Expect(cfg.ProxyPort).To(Equal(7687))
				Expect(cfg.Tenants).To(HaveKey("tenant1"))
				Expect(cfg.Tenants).To(HaveKey("tenant2"))
			})
		})
	})

	Describe("Configuration File Loading", func() {
		Context("when a valid config file exists", func() {
			It("should load configuration from file", func() {
				// Create temporary config file
				testConfig := &config.Config{
					ProxyPort: 8888,
					Tenants: map[string]config.TenantConfig{
						"test-tenant": {
							Host:     "testhost",
							Port:     9999,
							Username: "testuser",
							Password: "testpass",
						},
					},
				}

				data, err := json.Marshal(testConfig)
				Expect(err).NotTo(HaveOccurred())

				tmpFile, err := os.CreateTemp("", "config-*.json")
				Expect(err).NotTo(HaveOccurred())
				tempConfigFile = tmpFile.Name()

				_, err = tmpFile.Write(data)
				Expect(err).NotTo(HaveOccurred())
				tmpFile.Close()

				os.Setenv("CONFIG_FILE", tempConfigFile)

				cfg, err := config.Load()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.ProxyPort).To(Equal(8888))
				Expect(cfg.Tenants).To(HaveKey("test-tenant"))
				
				tenant := cfg.Tenants["test-tenant"]
				Expect(tenant.Host).To(Equal("testhost"))
				Expect(tenant.Port).To(Equal(9999))
				Expect(tenant.Username).To(Equal("testuser"))
				Expect(tenant.Password).To(Equal("testpass"))
			})
		})

		Context("when config file does not exist", func() {
			It("should return an error", func() {
				os.Setenv("CONFIG_FILE", "/nonexistent/config.json")
				
				cfg, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(cfg).To(BeNil())
			})
		})
	})

	Describe("Tenant Configuration Validation", func() {
		Context("when tenant configuration is valid", func() {
			It("should accept valid tenant configs", func() {
				cfg, err := config.Load()
				Expect(err).NotTo(HaveOccurred())
				
				for tenantID, tenantCfg := range cfg.Tenants {
					Expect(tenantID).NotTo(BeEmpty())
					Expect(tenantCfg.Host).NotTo(BeEmpty())
					Expect(tenantCfg.Port).To(BeNumerically(">", 0))
					Expect(tenantCfg.Port).To(BeNumerically("<", 65536))
				}
			})
		})
	})
})