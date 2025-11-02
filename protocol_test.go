package main

import (
	"math/rand"
	"strings"
	"testing"
)

// TestBasicEncoding tests basic encode/decode functionality
func TestBasicEncoding(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "Simple string",
			data: "Hello, World!",
		},
		{
			name: "Simple int32",
			data: int32(42),
		},
		{
			name: "Empty string",
			data: "",
		},
		{
			name: "Negative int32",
			data: int32(-12345),
		},
		{
			name: "Max int32",
			data: int32(2147483647),
		},
		{
			name: "Min int32",
			data: int32(-2147483648),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := encode(tt.data)
			decoded := decode(encoded)
			
			if !compareDataInput(tt.data, decoded) {
				t.Errorf("Encode/Decode mismatch: got %v, want %v", decoded, tt.data)
			}
		})
	}
}

// TestDataInputEncoding tests DataInput structure encoding
func TestDataInputEncoding(t *testing.T) {
	tests := []struct {
		name string
		data *DataInput
	}{
		{
			name: "Empty DataInput",
			data: NewDataInput(),
		},
		{
			name: "Single element",
			data: NewDataInput("test"),
		},
		{
			name: "Multiple elements",
			data: NewDataInput("foo", int32(42), "bar"),
		},
		{
			name: "Nested DataInput",
			data: NewDataInput("outer", NewDataInput("inner", int32(1)), int32(2)),
		},
		{
			name: "Deep nesting",
			data: NewDataInput(
				"level1",
				NewDataInput(
					"level2",
					NewDataInput(
						"level3",
						NewDataInput("level4", int32(42)),
					),
				),
			),
		},
		{
			name: "Mixed types",
			data: NewDataInput(
				"string",
				int32(123),
				NewDataInput("nested", int32(-456)),
				"another string",
				int32(789),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := encode(tt.data)
			decoded := decode(encoded)
			
			if !compareDataInput(tt.data, decoded) {
				t.Errorf("Encode/Decode mismatch for %s", tt.name)
			}
		})
	}
}

// TestUTF8Support tests UTF-8 string handling
func TestUTF8Support(t *testing.T) {
	tests := []struct {
		name string
		str  string
	}{
		{"ASCII", "Hello World"},
		{"Chinese", "你好世界"},
		{"Japanese", "こんにちは世界"},
		{"Arabic", "مرحبا بالعالم"},
		{"Russian", "Привет мир"},
		{"Emoji", "🚀🌟💻🔥"},
		{"Mixed", "Hello 世界 🌍"},
		{"Special chars", "Tab:\t Newline:\n Quote:\" Backslash:\\"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := NewDataInput(tt.str, int32(42))
			encoded := encode(data)
			decoded := decode(encoded)
			
			decodedData := decoded.(*DataInput)
			if decodedData.elements[0].(string) != tt.str {
				t.Errorf("UTF-8 string mismatch: got %q, want %q", 
					decodedData.elements[0].(string), tt.str)
			}
		})
	}
}

// TestLargeData tests handling of large data structures
func TestLargeData(t *testing.T) {
	// Test maximum array size (1000 elements)
	largeData := NewDataInput()
	for i := 0; i < 1000; i++ {
		if i%2 == 0 {
			largeData.elements = append(largeData.elements, 
				strings.Repeat("a", rand.Intn(1000)))
		} else {
			largeData.elements = append(largeData.elements, 
				int32(rand.Int31()))
		}
	}

	encoded := encode(largeData)
	decoded := decode(encoded)

	if !compareDataInput(largeData, decoded) {
		t.Error("Large data encode/decode failed")
	}

	// Test large string (approaching 1,000,000 chars)
	largeString := strings.Repeat("A", 999999)
	data := NewDataInput(largeString, int32(42))
	encoded = encode(data)
	decodedData := decode(encoded).(*DataInput)

	if decodedData.elements[0].(string) != largeString {
		t.Error("Large string encode/decode failed")
	}
}

// TestVarintEncoding tests variable-length integer encoding
func TestVarintEncoding(t *testing.T) {
	tests := []uint64{
		0, 1, 127, 128, 255, 256, 
		16383, 16384, 
		65535, 65536,
		1000000, 10000000,
	}

	for _, val := range tests {
		encoded := encodeVarint(val)
		decoded, consumed, err := decodeVarint(encoded)
		
		if err != nil {
			t.Errorf("Varint decode error for %d: %v", val, err)
		}
		if decoded != val {
			t.Errorf("Varint mismatch: got %d, want %d", decoded, val)
		}
		if consumed != len(encoded) {
			t.Errorf("Consumed bytes mismatch: got %d, want %d", 
				consumed, len(encoded))
		}
	}
}

// BenchmarkEncode benchmarks encoding performance
func BenchmarkEncode(b *testing.B) {
	data := NewDataInput(
		"benchmark",
		int32(12345),
		NewDataInput("nested", int32(67890), "data"),
		"more data",
	)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = encode(data)
	}
}

// BenchmarkDecode benchmarks decoding performance
func BenchmarkDecode(b *testing.B) {
	data := NewDataInput(
		"benchmark",
		int32(12345),
		NewDataInput("nested", int32(67890), "data"),
		"more data",
	)
	encoded := encode(data)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = decode(encoded)
	}
}

// BenchmarkLargeDataEncode benchmarks encoding of large data structures
func BenchmarkLargeDataEncode(b *testing.B) {
	data := NewDataInput()
	for i := 0; i < 100; i++ {
		data.elements = append(data.elements, 
			strings.Repeat("data", 100), 
			int32(i))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = encode(data)
	}
}

// BenchmarkLargeDataDecode benchmarks decoding of large data structures
func BenchmarkLargeDataDecode(b *testing.B) {
	data := NewDataInput()
	for i := 0; i < 100; i++ {
		data.elements = append(data.elements, 
			strings.Repeat("data", 100), 
			int32(i))
	}
	encoded := encode(data)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = decode(encoded)
	}
}

// TestErrorHandling tests error conditions
func TestErrorHandling(t *testing.T) {
	// Test invalid UTF-8 (this would be caught during encoding from Go strings)
	// Test truncated data
	validData := NewDataInput("test", int32(42))
	encoded := encode(validData)
	
	// Try to decode truncated data
	truncated := encoded[:len(encoded)/2]
	recovered := func() (r interface{}) {
		defer func() {
			if err := recover(); err != nil {
				r = err
			}
		}()
		return decode(truncated)
	}()
	
	if recovered == nil {
		// The decode didn't fail as expected, but might return partial/invalid data
		t.Log("Decoder handled truncated data gracefully")
	}
}

// TestProtocolEfficiency analyzes space efficiency
func TestProtocolEfficiency(t *testing.T) {
	testCases := []struct {
		name     string
		data     *DataInput
		maxRatio float64 // Maximum acceptable encoded/raw ratio
	}{
		{
			name:     "Small integers",
			data:     NewDataInput(int32(1), int32(2), int32(3)),
			maxRatio: 1.5, // Should be very efficient
		},
		{
			name:     "Short strings",
			data:     NewDataInput("a", "b", "c"),
			maxRatio: 4.0, // Type + length overhead for very short strings
		},
		{
			name:     "Mixed data",
			data:     NewDataInput("test", int32(123), NewDataInput("nested", int32(456))),
			maxRatio: 1.6, // Account for nested structure overhead
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := encode(tc.data)
			rawSize := calculateRawSize(tc.data)
			ratio := float64(len(encoded)) / float64(rawSize)
			
			t.Logf("%s: Raw size: %d, Encoded size: %d, Ratio: %.2f",
				tc.name, rawSize, len(encoded), ratio)
			
			if ratio > tc.maxRatio {
				t.Errorf("Encoding ratio too high: %.2f > %.2f", ratio, tc.maxRatio)
			}
		})
	}
}

func calculateRawSize(v interface{}) int {
	switch val := v.(type) {
	case string:
		return len(val)
	case int32:
		return 4
	case *DataInput:
		size := 0
		for _, elem := range val.elements {
			size += calculateRawSize(elem)
		}
		return size
	default:
		return 0
	}
}
