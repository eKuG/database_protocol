package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf8"
)

const (
	TypeString    byte = 0x01
	TypeInt32     byte = 0x02
	TypeDataInput byte = 0x03
	TypeNull      byte = 0x00
)

type DataInput struct {
	elements []interface{}
}

func NewDataInput(elements ...interface{}) *DataInput {
	return &DataInput{elements: elements}
}

func (d *DataInput) Elements() []interface{} {
	return d.elements
}

func encodeVarint(n uint64) []byte {
	var buf [10]byte
	var i int
	for n >= 0x80 {
		buf[i] = byte(n) | 0x80
		n >>= 7
		i++
	}
	buf[i] = byte(n)
	return buf[:i+1]
}

func decodeVarint(data []byte) (uint64, int, error) {
	var n uint64
	var shift uint
	for i := 0; i < len(data); i++ {
		if i > 9 {
			return 0, 0, errors.New("varint too long")
		}
		b := data[i]
		n |= uint64(b&0x7F) << shift
		if b < 0x80 {
			return n, i + 1, nil
		}
		shift += 7
	}
	return 0, 0, errors.New("incomplete varint")
}

// encode converts DataInput to a binary string
// Time Complexity: O(n) where n is the total number of elements including nested ones
// Space Complexity: O(m) where m is the total size of all data
func encode(toSend interface{}) string {
	buf := &buffer{data: make([]byte, 0, 1024)} // Pre-allocate for efficiency
	encodeElement(buf, toSend)
	return string(buf.data)
}

// encodeElement recursively encodes a single element
// Time Complexity: O(1) for primitives, O(k) for strings where k is string length,
//                  O(n) for DataInput where n is number of elements
func encodeElement(buf *buffer, elem interface{}) error {
	switch v := elem.(type) {
	case string:
		// Encode string: [TypeString][Length as varint][UTF-8 bytes]
		buf.WriteByte(TypeString)
		bytes := []byte(v)
		buf.Write(encodeVarint(uint64(len(bytes))))
		buf.Write(bytes)
		
	case int32:
		// Encode int32: [TypeInt32][4 bytes little-endian]
		buf.WriteByte(TypeInt32)
		var bytes [4]byte
		binary.LittleEndian.PutUint32(bytes[:], uint32(v))
		buf.Write(bytes[:])
		
	case *DataInput:
		// Encode DataInput: [TypeDataInput][Count as varint][Elements...]
		buf.WriteByte(TypeDataInput)
		buf.Write(encodeVarint(uint64(len(v.elements))))
		for _, subElem := range v.elements {
			if err := encodeElement(buf, subElem); err != nil {
				return err
			}
		}
		
	case nil:
		buf.WriteByte(TypeNull)
		
	default:
		return fmt.Errorf("unsupported type: %T", elem)
	}
	return nil
}

// decode converts a binary string back to DataInput
// Time Complexity: O(n) where n is the total number of elements
// Space Complexity: O(m) where m is the total size of decoded data
func decode(received string) interface{} {
	data := []byte(received)
	result, _, _ := decodeElement(data, 0)
	return result
}

// decodeElement recursively decodes a single element
// Returns: decoded element, bytes consumed, error
// Time Complexity: O(1) for primitives, O(k) for strings, O(n) for DataInput
func decodeElement(data []byte, offset int) (interface{}, int, error) {
	if offset >= len(data) {
		return nil, 0, errors.New("unexpected end of data")
	}
	
	typeTag := data[offset]
	offset++
	
	switch typeTag {
	case TypeString:
		// Decode string length
		length, consumed, err := decodeVarint(data[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += consumed
		
		// Read string bytes
		if offset+int(length) > len(data) {
			return nil, 0, errors.New("string length exceeds data")
		}
		str := string(data[offset : offset+int(length)])
		
		// Validate UTF-8
		if !utf8.ValidString(str) {
			return nil, 0, errors.New("invalid UTF-8 string")
		}
		
		return str, offset + int(length), nil
		
	case TypeInt32:
		// Read 4 bytes for int32
		if offset+4 > len(data) {
			return nil, 0, errors.New("insufficient data for int32")
		}
		val := binary.LittleEndian.Uint32(data[offset : offset+4])
		return int32(val), offset + 4, nil
		
	case TypeDataInput:
		// Decode element count
		count, consumed, err := decodeVarint(data[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += consumed
		
		// Decode each element
		elements := make([]interface{}, 0, count)
		for i := 0; i < int(count); i++ {
			elem, bytesRead, err := decodeElement(data, offset)
			if err != nil {
				return nil, 0, err
			}
			elements = append(elements, elem)
			offset = bytesRead
		}
		
		return &DataInput{elements: elements}, offset, nil
		
	case TypeNull:
		return nil, offset, nil
		
	default:
		return nil, 0, fmt.Errorf("unknown type tag: %02x", typeTag)
	}
}

// buffer is a simple byte buffer for efficient encoding
type buffer struct {
	data []byte
}

func (b *buffer) Write(p []byte) {
	b.data = append(b.data, p...)
}

func (b *buffer) WriteByte(c byte) {
	b.data = append(b.data, c)
}

// Helper function to compare DataInput structures (for testing)
func compareDataInput(a, b interface{}) bool {
	switch va := a.(type) {
	case string:
		vb, ok := b.(string)
		return ok && va == vb
	case int32:
		vb, ok := b.(int32)
		return ok && va == vb
	case *DataInput:
		vb, ok := b.(*DataInput)
		if !ok || len(va.elements) != len(vb.elements) {
			return false
		}
		for i := range va.elements {
			if !compareDataInput(va.elements[i], vb.elements[i]) {
				return false
			}
		}
		return true
	case nil:
		return b == nil
	default:
		return false
	}
}
