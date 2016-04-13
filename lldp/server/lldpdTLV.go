package lldpServer

import (
	"encoding/binary"
	"errors"
	"io"
)

// tlv type def
type lldpTLVType uint8

const (
	// max tlv length of value data which is allowed in TLV
	LLDPTLVLengthMax = 0x01ff

	// Mandatory TLVType values in all LLDPDUs or LLDP Frame
	TLVTypeEnd       lldpTLVType = 0 // indicates end of Frame
	TLVTypeChassisID lldpTLVType = 1 // indicates tlv is for chassis id
	TLVTypePortID    lldpTLVType = 2 // indicates tlv is for port id
	TLVTypeTTL       lldpTLVType = 3 // indicates tlv is for ttl for rx frame

	// Optional TLVType values which may occur in LLDPDUs or LLDP Frame
	TLVTypePortDescription    lldpTLVType = 4
	TLVTypeSystemName         lldpTLVType = 5
	TLVTypeSystemDescription  lldpTLVType = 6
	TLVTypeSystemCapabilities lldpTLVType = 7
	TLVTypeManagementAddress  lldpTLVType = 8

	// TLVType which can be used
	// to carry organization-specific data in a special format.
	TLVTypeOrganizationSpecific lldpTLVType = 127

	// maximum possible value for a TLVType.
	TLVTypeMax lldpTLVType = TLVTypeOrganizationSpecific
)

var (
	// TLV Error type
	LLDP_ERR_INVALID_TLV_LENGTH = errors.New("Invalid TLV length")
	LLDP_ERR_INVALID_TLV_TYPE   = errors.New("Invalid TLV type")
)

// TLV structure used to carry information in an encoded format.
type LLDPTLV struct {
	// Type specifies the type of value carried in TLV.
	Type lldpTLVType

	// Length specifies the length of the value carried in TLV.
	Length uint16

	// Value specifies the raw data carried in TLV.
	Value []byte
}

// Marshall tlv information into binary form
// 1) Check type value
// 2) Check Length
func (c *LLDPTLV) LLDPTLVMarshallInsert(b []byte, offset *int) error {
	// check type
	if c.Type > TLVTypeMax {
		return LLDP_ERR_INVALID_TLV_TYPE
	}

	// check length
	if int(c.Length) != len(c.Value) {
		return LLDP_ERR_INVALID_TLV_LENGTH
	}
	// copy value into b
	// type : 8 bits
	// leng : 16 bits
	// value: N bytes
	// @FIXME: the lenght part
	// No need to make new byte as the caller will create the byte
	// They will also specify offset
	//b := make([]byte, 2+len(c.Value))

	var typeByte uint16
	typeByte |= uint16(c.Type) << 9
	typeByte |= c.Length
	binary.BigEndian.PutUint16(b[(0+*offset):(2+*offset)], typeByte)
	copy(b[(2+*offset):(int(c.Length)+*offset)], c.Value)
	*offset += len(b)
	return nil
}

func (c *LLDPTLV) LLDPTLVMarshallBinary() ([]byte, error) {
	// check type
	if c.Type > TLVTypeMax {
		return nil, LLDP_ERR_INVALID_TLV_TYPE
	}

	// check length
	if int(c.Length) != len(c.Value) {
		return nil, LLDP_ERR_INVALID_TLV_LENGTH
	}
	// copy value into b
	// type : 8 bits
	// leng : 16 bits
	// value: N bytes
	// @FIXME: the lenght part
	b := make([]byte, 2+len(c.Value))

	var typeByte uint16
	typeByte |= uint16(c.Type) << 9
	typeByte |= c.Length
	binary.BigEndian.PutUint16(b[0:2], typeByte)
	copy(b[2:], c.Value)

	return b, nil
}

// UnMarshall tlv information from binary form to LLDPTLV
func (c *LLDPTLV) LLDPTLVUnmarshalBinary(b []byte) error {
	// Must contain type and length values, which are mandatory fields
	if len(b) < 2 {
		return io.ErrUnexpectedEOF
	}

	// type : 8 bits
	// leng : 16 bits
	// value: N bytes
	// @FIXME: the lenght part
	c.Type = lldpTLVType(b[0]) >> 1
	c.Length = binary.BigEndian.Uint16(b[0:2]) & LLDPTLVLengthMax

	// Must contain at least enough bytes as indicated by length
	if len(b[2:]) < int(c.Length) {
		return io.ErrUnexpectedEOF
	}

	// Copy value directly into TLV
	c.Value = make([]byte, len(b[2:2+c.Length]))
	copy(c.Value, b[2:2+c.Length])

	return nil
}