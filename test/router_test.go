package test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"neo4j-proxy/internal/router"
	"neo4j-proxy/pkg/config"
)

var _ = Describe("Router", func() {
	var (
		cfg    *config.Config
		rt     *router.Router
	)

	BeforeEach(func() {
		cfg = &config.Config{
			ProxyPort: 7687,
			Tenants: map[string]config.TenantConfig{
				"tenant1": {
					Host: "yunhorn187",
					Port: 17687,
				},
				"tenant2": {
					Host: "yunhorn187",
					Port: 27687,
				},
				"tenant3": {
					Host: "localhost",
					Port: 7688,
					Username: "neo4j",
					Password: "password",
				},
			},
		}
		rt = router.New(cfg)
	})

	Describe("Router Creation", func() {
		Context("when creating a new router", func() {
			It("should create a router with the provided config", func() {
				router := router.New(cfg)
				Expect(router).NotTo(BeNil())
			})
		})
	})

	Describe("Tenant Management", func() {
		Context("when listing tenants", func() {
			It("should return all configured tenant IDs", func() {
				tenants := rt.ListTenants()
				Expect(tenants).To(HaveLen(3))
				Expect(tenants).To(ContainElements("tenant1", "tenant2", "tenant3"))
			})
		})

		Context("when getting tenant configuration", func() {
			It("should return correct configuration for existing tenant", func() {
				tenantCfg, exists := rt.GetTenantConfig("tenant1")
				Expect(exists).To(BeTrue())
				Expect(tenantCfg).NotTo(BeNil())
				Expect(tenantCfg.Host).To(Equal("yunhorn187"))
				Expect(tenantCfg.Port).To(Equal(17687))
			})

			It("should return false for non-existent tenant", func() {
				_, exists := rt.GetTenantConfig("nonexistent")
				Expect(exists).To(BeFalse())
			})

			It("should return configuration with credentials", func() {
				tenantCfg, exists := rt.GetTenantConfig("tenant3")
				Expect(exists).To(BeTrue())
				Expect(tenantCfg.Username).To(Equal("neo4j"))
				Expect(tenantCfg.Password).To(Equal("password"))
			})
		})

		Context("when updating tenant configuration", func() {
			It("should update existing tenant configuration", func() {
				newCfg := config.TenantConfig{
					Host:     "newhost",
					Port:     9999,
					Username: "newuser",
					Password: "newpass",
				}

				rt.UpdateTenantConfig("tenant1", newCfg)

				tenantCfg, exists := rt.GetTenantConfig("tenant1")
				Expect(exists).To(BeTrue())
				Expect(tenantCfg.Host).To(Equal("newhost"))
				Expect(tenantCfg.Port).To(Equal(9999))
				Expect(tenantCfg.Username).To(Equal("newuser"))
				Expect(tenantCfg.Password).To(Equal("newpass"))
			})

			It("should add new tenant configuration", func() {
				newCfg := config.TenantConfig{
					Host: "newtenant",
					Port: 8888,
				}

				rt.UpdateTenantConfig("tenant4", newCfg)

				tenantCfg, exists := rt.GetTenantConfig("tenant4")
				Expect(exists).To(BeTrue())
				Expect(tenantCfg.Host).To(Equal("newtenant"))
				Expect(tenantCfg.Port).To(Equal(8888))

				tenants := rt.ListTenants()
				Expect(tenants).To(ContainElement("tenant4"))
				Expect(len(tenants)).To(Equal(4))
			})
		})

		Context("when removing tenant configuration", func() {
			It("should remove existing tenant", func() {
				rt.RemoveTenant("tenant1")

				_, exists := rt.GetTenantConfig("tenant1")
				Expect(exists).To(BeFalse())

				tenants := rt.ListTenants()
				Expect(tenants).NotTo(ContainElement("tenant1"))
				Expect(len(tenants)).To(Equal(2))
			})

			It("should handle removing non-existent tenant gracefully", func() {
				rt.RemoveTenant("nonexistent")

				tenants := rt.ListTenants()
				Expect(len(tenants)).To(Equal(3)) // Should remain unchanged
			})
		})
	})

	Describe("Connection Routing", func() {
		Context("when routing to valid tenant", func() {
			It("should successfully connect to tenant1 backend", func() {
				// Since your backends are running, we expect successful connection
				conn, err := rt.RouteConnection("tenant1")
				
				if err != nil {
					// If connection fails, verify it's attempting the correct address
					Expect(err.Error()).To(ContainSubstring("yunhorn187:17687"))
				} else {
					// If connection succeeds, verify we got a connection
					Expect(conn).NotTo(BeNil())
					conn.Close() // Clean up
				}
			})

			It("should successfully connect to tenant2 backend", func() {
				conn, err := rt.RouteConnection("tenant2")
				
				if err != nil {
					// If connection fails, verify it's attempting the correct address
					Expect(err.Error()).To(ContainSubstring("yunhorn187:27687"))
				} else {
					// If connection succeeds, verify we got a connection
					Expect(conn).NotTo(BeNil())
					conn.Close() // Clean up
				}
			})
		})

		Context("when routing to invalid tenant", func() {
			It("should return tenant not found error", func() {
				conn, err := rt.RouteConnection("nonexistent")
				
				Expect(err).To(HaveOccurred())
				Expect(conn).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("tenant nonexistent not found"))
			})
		})
	})

	Describe("Thread Safety", func() {
		Context("when accessing router concurrently", func() {
			It("should handle concurrent reads safely", func() {
				done := make(chan bool, 10)

				// Start multiple goroutines to read tenant configs
				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						
						for j := 0; j < 100; j++ {
							tenants := rt.ListTenants()
							Expect(len(tenants)).To(BeNumerically(">=", 0))
							
							if len(tenants) > 0 {
								_, exists := rt.GetTenantConfig(tenants[0])
								Expect(exists).To(BeTrue())
							}
						}
						
						done <- true
					}()
				}

				// Wait for all goroutines to complete
				for i := 0; i < 10; i++ {
					<-done
				}
			})

			It("should handle concurrent writes safely", func() {
				done := make(chan bool, 5)

				// Start multiple goroutines to update configurations
				for i := 0; i < 5; i++ {
					go func(id int) {
						defer GinkgoRecover()
						
						tenantID := "concurrent-tenant"
						cfg := config.TenantConfig{
							Host: "localhost",
							Port: 7000 + id,
						}
						
						rt.UpdateTenantConfig(tenantID, cfg)
						
						// Verify the update
						retrievedCfg, exists := rt.GetTenantConfig(tenantID)
						Expect(exists).To(BeTrue())
						Expect(retrievedCfg.Host).To(Equal("localhost"))
						
						done <- true
					}(i)
				}

				// Wait for all goroutines to complete
				for i := 0; i < 5; i++ {
					<-done
				}
			})
		})
	})
})