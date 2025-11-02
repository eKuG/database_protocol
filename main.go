package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("Binary Protocol Implementation")
	fmt.Println("===========================================")
	fmt.Println()

	runTests()
	runBenchmarks()
	demonstrateExtensibility()
}

func runTests() {
	fmt.Println("Running Test Cases:")
	fmt.Println("------------------")

	fmt.Println("\nTest 1: Basic nested structure")
	originalData := NewDataInput("foo", NewDataInput("bar", int32(42)))
	encoded := encode(originalData)
	decoded := decode(encoded)
	
	fmt.Printf("Original: %+v\n", formatDataInput(originalData))
	fmt.Printf("Encoded size: %d bytes\n", len(encoded))
	fmt.Printf("Decoded: %+v\n", formatDataInput(decoded))
	fmt.Printf("Match: %v\n", compareDataInput(originalData, decoded))

	fmt.Println("\nTest 2: Complex nested structure")
	complexData := NewDataInput(
		"user_metrics",
		int32(1234567),
		NewDataInput(
			"events",
			int32(42),
			"click",
			NewDataInput("nested", int32(-999), "deep"),
			"timestamp",
		),
		"end",
	)
	encoded2 := encode(complexData)
	decoded2 := decode(encoded2)
	fmt.Printf("Encoded size: %d bytes\n", len(encoded2))
	fmt.Printf("Match: %v\n", compareDataInput(complexData, decoded2))

	fmt.Println("\nTest 3: Large string handling")
	largeString := make([]byte, 100000)
	for i := range largeString {
		largeString[i] = byte('A' + (i % 26))
	}
	largeData := NewDataInput(string(largeString), int32(999))
	encoded3 := encode(largeData)
	decoded3 := decode(encoded3)
	fmt.Printf("Large string size: %d bytes\n", len(largeString))
	fmt.Printf("Encoded size: %d bytes\n", len(encoded3))
	fmt.Printf("Match: %v\n", compareDataInput(largeData, decoded3))

	fmt.Println("\nTest 4: UTF-8 string support")
	utf8Data := NewDataInput("Hello Ekansh", "🚀 Rocket", int32(2025), "abcdefg")
	encoded4 := encode(utf8Data)
	decoded4 := decode(encoded4)
	fmt.Printf("Original: %+v\n", formatDataInput(utf8Data))
	fmt.Printf("Decoded: %+v\n", formatDataInput(decoded4))
	fmt.Printf("Match: %v\n", compareDataInput(utf8Data, decoded4))

	fmt.Println("\nTest 5: Edge cases")
	edgeData := NewDataInput("", int32(0), NewDataInput(), int32(-2147483648))
	encoded5 := encode(edgeData)
	decoded5 := decode(encoded5)
	fmt.Printf("Match: %v\n", compareDataInput(edgeData, decoded5))
}

func runBenchmarks() {
	fmt.Println("\n\nPerformance Benchmarks:")
	fmt.Println("----------------------")

	fmt.Println("\nBenchmark 1: Small messages (10 elements)")
	smallData := NewDataInput()
	for i := 0; i < 10; i++ {
		smallData.elements = append(smallData.elements, fmt.Sprintf("field_%d", i), int32(i))
	}
	benchmarkEncodeDecode(smallData, 10000)

	fmt.Println("\nBenchmark 2: Medium messages (100 elements)")
	mediumData := NewDataInput()
	for i := 0; i < 100; i++ {
		mediumData.elements = append(mediumData.elements, fmt.Sprintf("field_%d", i), int32(i))
	}
	benchmarkEncodeDecode(mediumData, 1000)

	fmt.Println("\nBenchmark 3: Large nested structure")
	nestedData := NewDataInput()
	for i := 0; i < 10; i++ {
		innerData := NewDataInput()
		for j := 0; j < 10; j++ {
			innerData.elements = append(innerData.elements, fmt.Sprintf("data_%d_%d", i, j), int32(i*10+j))
		}
		nestedData.elements = append(nestedData.elements, innerData)
	}
	benchmarkEncodeDecode(nestedData, 1000)

	fmt.Println("\nBenchmark 4: Maximum size array (1000 elements)")
	maxData := NewDataInput()
	for i := 0; i < 1000; i++ {
		if i%3 == 0 {
			maxData.elements = append(maxData.elements, fmt.Sprintf("element_%d", i))
		} else {
			maxData.elements = append(maxData.elements, int32(i))
		}
	}
	benchmarkEncodeDecode(maxData, 100)
}

func benchmarkEncodeDecode(data *DataInput, iterations int) {
	start := time.Now()
	var encoded string
	for i := 0; i < iterations; i++ {
		encoded = encode(data)
	}
	encodeTime := time.Since(start)

	start = time.Now()
	for i := 0; i < iterations; i++ {
		_ = decode(encoded)
	}
	decodeTime := time.Since(start)

	fmt.Printf("Elements: %d, Encoded size: %d bytes\n", len(data.elements), len(encoded))
	fmt.Printf("Encode: %d iterations in %v (%.2f µs/op)\n", 
		iterations, encodeTime, float64(encodeTime.Microseconds())/float64(iterations))
	fmt.Printf("Decode: %d iterations in %v (%.2f µs/op)\n",
		iterations, decodeTime, float64(decodeTime.Microseconds())/float64(iterations))

	totalBytes := len(encoded) * iterations
	encodeThroughput := float64(totalBytes) / encodeTime.Seconds() / 1024 / 1024
	decodeThroughput := float64(totalBytes) / decodeTime.Seconds() / 1024 / 1024
	fmt.Printf("Throughput - Encode: %.2f MB/s, Decode: %.2f MB/s\n", 
		encodeThroughput, decodeThroughput)
}

func demonstrateExtensibility() {
	fmt.Println("\n\nExtensibility Demonstration:")
	fmt.Println("---------------------------")

	fmt.Print(`
To add support for more types, follow these steps:

1. Define a new type constant:
   const TypeFloat64 byte = 0x04
   const TypeBoolean byte = 0x05
   const TypeBytes   byte = 0x06
   const TypeDate    byte = 0x07

2. Add encoding logic in encodeElement():
   case float64:
       buf.WriteByte(TypeFloat64)
       var bytes [8]byte
       binary.LittleEndian.PutUint64(bytes[:], math.Float64bits(v))
       buf.Write(bytes[:])

3. Add decoding logic in decodeElement():
   case TypeFloat64:
       if offset+8 > len(data) {
           return nil, 0, errors.New("insufficient data for float64")
       }
       bits := binary.LittleEndian.Uint64(data[offset : offset+8])
       return math.Float64frombits(bits), offset + 8, nil

4. For complex types (e.g., maps, timestamps):
   - Maps: Encode as [TypeMap][Count][Key1][Value1][Key2][Value2]...
   - Timestamps: Encode as int64 Unix nanoseconds
   - UUID: Encode as 16-byte fixed array

5. Version compatibility:
   - Reserve type bytes 0xF0-0xFF for protocol extensions
   - Add version header for protocol negotiation
   - Implement schema registry for type evolution
`)

	fmt.Println("\nProtocol Efficiency Analysis:")
	fmt.Println("- Variable-length encoding saves 1-7 bytes per integer")
	fmt.Println("- Type tags use only 1 byte overhead per value")
	fmt.Println("- No padding or alignment requirements")
	fmt.Println("- Zero-copy decoding possible for primitive types")
	fmt.Println("- Compression-friendly due to type clustering")
}

func formatDataInput(v interface{}) string {
	switch val := v.(type) {
	case string:
		if len(val) > 50 {
			return fmt.Sprintf("\"%s...\" (len=%d)", val[:47], len(val))
		}
		return fmt.Sprintf("\"%s\"", val)
	case int32:
		return fmt.Sprintf("%d", val)
	case *DataInput:
		result := "DataInput{"
		for i, elem := range val.elements {
			if i > 0 {
				result += ", "
			}
			result += formatDataInput(elem)
		}
		result += "}"
		return result
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("%v", val)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
