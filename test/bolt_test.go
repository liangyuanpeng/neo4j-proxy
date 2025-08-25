package test

import (
	"bytes"
	"encoding/binary"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"neo4j-proxy/pkg/bolt"
)

var _ = Describe("Bolt Protocol", func() {
	Describe("Protocol Constants", func() {
		It("should have correct magic preamble", func() {
			Expect(bolt.BoltMagicPreamble).To(Equal(uint32(0x6060B017)))
		})

		It("should have correct message types", func() {
			Expect(bolt.MsgInit).To(Equal(byte(0x01)))
			Expect(bolt.MsgRun).To(Equal(byte(0x10)))
			Expect(bolt.MsgSuccess).To(Equal(byte(0x70)))
			Expect(bolt.MsgFailure).To(Equal(byte(0x7F)))
		})

		It("should have correct version constants", func() {
			Expect(bolt.Version1).To(Equal(1))
			Expect(bolt.Version2).To(Equal(2))
			Expect(bolt.Version3).To(Equal(3))
			Expect(bolt.Version4).To(Equal(4))
		})
	})

	Describe("Connection Management", func() {
		var (
			serverConn, clientConn net.Conn
			boltConn               *bolt.Connection
		)

		BeforeEach(func() {
			var err error
			serverConn, clientConn, err = createConnectedPair()
			Expect(err).NotTo(HaveOccurred())
			
			boltConn = bolt.NewConnection(clientConn)
		})

		AfterEach(func() {
			if serverConn != nil {
				serverConn.Close()
			}
			if clientConn != nil {
				clientConn.Close()
			}
		})

		Context("when creating a new connection", func() {
			It("should create a connection wrapper", func() {
				conn := bolt.NewConnection(clientConn)
				Expect(conn).NotTo(BeNil())
				Expect(conn.GetVersion()).To(Equal(0)) // Not negotiated yet
			})
		})

		Context("when performing handshake", func() {
			It("should handle valid handshake", func() {
				// Simulate server side of handshake
				go func() {
					defer GinkgoRecover()
					
					// Read magic preamble
					var magic uint32
					err := binary.Read(serverConn, binary.BigEndian, &magic)
					Expect(err).NotTo(HaveOccurred())
					Expect(magic).To(Equal(bolt.BoltMagicPreamble))
					
					// Read version requests
					versions := make([]uint32, 4)
					for i := 0; i < 4; i++ {
						err := binary.Read(serverConn, binary.BigEndian, &versions[i])
						Expect(err).NotTo(HaveOccurred())
					}
					
					// Send back agreed version
					err = binary.Write(serverConn, binary.BigEndian, uint32(bolt.Version4))
					Expect(err).NotTo(HaveOccurred())
				}()

				// Perform handshake
				err := performClientHandshake(clientConn)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should reject invalid magic preamble", func() {
				go func() {
					defer GinkgoRecover()
					
					// Send invalid magic
					err := binary.Write(serverConn, binary.BigEndian, uint32(0x12345678))
					Expect(err).NotTo(HaveOccurred())
				}()

				err := boltConn.Handshake()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid Bolt magic preamble"))
			})
		})

		Context("when handling messages", func() {
			It("should read messages correctly", func() {
				go func() {
					defer GinkgoRecover()
					
					// Send a simple message
					chunkSize := uint16(1)
					err := binary.Write(serverConn, binary.BigEndian, chunkSize)
					Expect(err).NotTo(HaveOccurred())
					
					err = binary.Write(serverConn, binary.BigEndian, byte(bolt.MsgInit))
					Expect(err).NotTo(HaveOccurred())
				}()

				msg, err := boltConn.ReadMessage()
				Expect(err).NotTo(HaveOccurred())
				Expect(msg).NotTo(BeNil())
				Expect(msg.Signature).To(Equal(byte(bolt.MsgInit)))
			})

			It("should write messages correctly", func() {
				msg := &bolt.Message{
					Signature: bolt.MsgSuccess,
					Fields:    []interface{}{},
				}

				go func() {
					defer GinkgoRecover()
					
					// Read chunk size
					var chunkSize uint16
					err := binary.Read(serverConn, binary.BigEndian, &chunkSize)
					Expect(err).NotTo(HaveOccurred())
					Expect(chunkSize).To(Equal(uint16(1)))
					
					// Read message signature
					var signature byte
					err = binary.Read(serverConn, binary.BigEndian, &signature)
					Expect(err).NotTo(HaveOccurred())
					Expect(signature).To(Equal(byte(bolt.MsgSuccess)))
				}()

				err := boltConn.WriteMessage(msg)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle empty chunks", func() {
				go func() {
					defer GinkgoRecover()
					
					// Send zero chunk size
					err := binary.Write(serverConn, binary.BigEndian, uint16(0))
					Expect(err).NotTo(HaveOccurred())
				}()

				_, err := boltConn.ReadMessage()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("empty chunk"))
			})
		})

		Context("when extracting tenant information", func() {
			It("should extract tenant from INIT message", func() {
				msg := &bolt.Message{
					Signature: bolt.MsgInit,
					Fields:    []interface{}{},
				}

				tenantID, err := boltConn.ExtractTenantID(msg)
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantID).To(Equal("tenant1")) // Default behavior
			})

			It("should handle non-INIT messages", func() {
				msg := &bolt.Message{
					Signature: bolt.MsgRun,
					Fields:    []interface{}{},
				}

				_, err := boltConn.ExtractTenantID(msg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to determine tenant ID"))
			})
		})
	})

	Describe("Message Structure", func() {
		Context("when working with messages", func() {
			It("should create messages with correct structure", func() {
				msg := &bolt.Message{
					Signature: bolt.MsgInit,
					Fields:    []interface{}{"field1", "field2"},
				}

				Expect(msg.Signature).To(Equal(byte(bolt.MsgInit)))
				Expect(len(msg.Fields)).To(Equal(2))
			})
		})
	})
})

// Helper functions
func createConnectedPair() (net.Conn, net.Conn, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, err
	}
	defer listener.Close()

	// Channel to get server connection
	serverConnChan := make(chan net.Conn, 1)
	errChan := make(chan error, 1)

	// Start server goroutine
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			errChan <- err
			return
		}
		serverConnChan <- conn
	}()

	// Connect as client
	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		return nil, nil, err
	}

	// Get server connection
	select {
	case serverConn := <-serverConnChan:
		return serverConn, clientConn, nil
	case err := <-errChan:
		clientConn.Close()
		return nil, nil, err
	}
}

func performClientHandshake(conn net.Conn) error {
	// Send magic preamble
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, bolt.BoltMagicPreamble)
	
	// Send version requests
	binary.Write(&buf, binary.BigEndian, uint32(bolt.Version4))
	binary.Write(&buf, binary.BigEndian, uint32(bolt.Version3))
	binary.Write(&buf, binary.BigEndian, uint32(bolt.Version2))
	binary.Write(&buf, binary.BigEndian, uint32(bolt.Version1))
	
	_, err := conn.Write(buf.Bytes())
	if err != nil {
		return err
	}
	
	// Read server response
	var agreedVersion uint32
	return binary.Read(conn, binary.BigEndian, &agreedVersion)
}