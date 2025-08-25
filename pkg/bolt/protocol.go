package bolt

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

const (
	// Bolt protocol magic number
	BoltMagicPreamble = 0x6060B017
	
	// Message types
	MsgInit     = 0x01
	MsgRun      = 0x10
	MsgRecord   = 0x71
	MsgSuccess  = 0x70
	MsgFailure  = 0x7F
	MsgIgnored  = 0x7E
	MsgPullAll  = 0x3F
	MsgDiscardAll = 0x2F
	MsgReset    = 0x0F
	MsgBye      = 0x02
	
	// Bolt versions
	Version1 = 1
	Version2 = 2
	Version3 = 3
	Version4 = 4
)

// Message represents a Bolt protocol message
type Message struct {
	Signature byte
	Fields    []interface{}
}

// Connection represents a Bolt protocol connection
type Connection struct {
	conn    net.Conn
	version int
}

// NewConnection creates a new Bolt connection wrapper
func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

// Handshake performs the Bolt protocol handshake
func (c *Connection) Handshake() error {
	// Read the magic preamble
	var magic uint32
	if err := binary.Read(c.conn, binary.BigEndian, &magic); err != nil {
		return err
	}
	
	if magic != BoltMagicPreamble {
		return errors.New("invalid Bolt magic preamble")
	}
	
	// Read supported versions (4 x uint32)
	versions := make([]uint32, 4)
	for i := 0; i < 4; i++ {
		if err := binary.Read(c.conn, binary.BigEndian, &versions[i]); err != nil {
			return err
		}
	}
	
	// Choose the highest supported version (for now, support version 4)
	selectedVersion := uint32(0)
	for _, version := range versions {
		if version <= Version4 && version > selectedVersion {
			selectedVersion = version
		}
	}
	
	if selectedVersion == 0 {
		selectedVersion = Version1
	}
	
	c.version = int(selectedVersion)
	
	// Send back the selected version
	return binary.Write(c.conn, binary.BigEndian, selectedVersion)
}

// ReadMessage reads a message from the connection
func (c *Connection) ReadMessage() (*Message, error) {
	// Read chunk size
	var chunkSize uint16
	if err := binary.Read(c.conn, binary.BigEndian, &chunkSize); err != nil {
		return nil, err
	}
	
	if chunkSize == 0 {
		return nil, errors.New("empty chunk")
	}
	
	// Read chunk data
	data := make([]byte, chunkSize)
	if _, err := io.ReadFull(c.conn, data); err != nil {
		return nil, err
	}
	
	// Parse message (simplified - in production would need full PackStream parser)
	if len(data) < 1 {
		return nil, errors.New("invalid message format")
	}
	
	msg := &Message{
		Signature: data[0],
		Fields:    []interface{}{}, // Simplified for now
	}
	
	return msg, nil
}

// WriteMessage writes a message to the connection
func (c *Connection) WriteMessage(msg *Message) error {
	// Simplified message writing (in production would need full PackStream serializer)
	data := []byte{msg.Signature}
	
	// Write chunk size
	chunkSize := uint16(len(data))
	if err := binary.Write(c.conn, binary.BigEndian, chunkSize); err != nil {
		return err
	}
	
	// Write chunk data
	_, err := c.conn.Write(data)
	return err
}

// ExtractTenantID extracts tenant identifier from the connection
// This is a simplified implementation - in practice you might extract from
// authentication credentials, connection parameters, or other metadata
func (c *Connection) ExtractTenantID(msg *Message) (string, error) {
	// For now, we'll use a simple round-robin approach
	// In production, this would extract from authentication or connection metadata
	if msg.Signature == MsgInit {
		// For demo purposes, assume tenant ID is in connection metadata
		// This would be extracted from the INIT message fields in a real implementation
		return "tenant1", nil // Simplified
	}
	
	return "", errors.New("unable to determine tenant ID")
}

// Close closes the underlying connection
func (c *Connection) Close() error {
	return c.conn.Close()
}

// GetVersion returns the negotiated protocol version
func (c *Connection) GetVersion() int {
	return c.version
}