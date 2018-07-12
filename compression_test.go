package kafka_test

import (
	"math/rand"
	"testing"

	kafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/gzip"
	"github.com/segmentio/kafka-go/lz4"
	"github.com/segmentio/kafka-go/snappy"
)

func TestCompression(t *testing.T) {
	msg := kafka.Message{
		Value: []byte("message"),
	}

	testEncodeDecode(t, msg, kafka.CompressionNoneCode)
	testEncodeDecode(t, msg, gzip.Code)
	testEncodeDecode(t, msg, snappy.Code)
	testEncodeDecode(t, msg, lz4.Code)
	testUnknownCodec(t, msg, 42)
}

func testEncodeDecode(t *testing.T, m kafka.Message, codec int8) {
	var r1, r2 kafka.Message
	var err error

	t.Run("encode with "+codecToStr(codec), func(t *testing.T) {
		m.CompressionCodec = codec
		r1, err = m.Encode()
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("encode with "+codecToStr(codec), func(t *testing.T) {
		r2, err = r1.Decode()
		if err != nil {
			t.Error(err)
		}
		if string(r2.Value) != "message" {
			t.Error("bad message")
			t.Log("got: ", r2.Value)
			t.Log("expected: ", []byte("message"))
		}
	})
}

func testUnknownCodec(t *testing.T, m kafka.Message, codec int8) {
	t.Run("unknown codec", func(t *testing.T) {
		expectedErr := "codec 42 not imported."
		m.CompressionCodec = codec
		_, err := m.Encode()
		if err.Error() != expectedErr {
			t.Error("wrong error")
			t.Log("got: ", err)
			t.Error("expected: ", expectedErr)
		}
	})
}

func codecToStr(codec int8) string {
	switch codec {
	case kafka.CompressionNoneCode:
		return "none"
	case gzip.Code:
		return "gzip"
	case snappy.Code:
		return "snappy"
	case lz4.Code:
		return "lz4"
	default:
		return "unknown"
	}
}

func BenchmarkCompression(b *testing.B) {
	benchmarks := []struct {
		scenario string
		codec    int8
		function func(*testing.B, int8, int, map[int][]byte)
	}{
		{
			scenario: "None",
			codec:    kafka.CompressionNoneCode,
			function: benchmarkCompression,
		},
		{
			scenario: "GZIP",
			codec:    gzip.Code,
			function: benchmarkCompression,
		},
		{
			scenario: "Snappy",
			codec:    snappy.Code,
			function: benchmarkCompression,
		},
		{
			scenario: "LZ4",
			codec:    lz4.Code,
			function: benchmarkCompression,
		},
	}

	payload := map[int][]byte{
		1024:  randomPayload(1024),
		4096:  randomPayload(4096),
		8192:  randomPayload(8192),
		16384: randomPayload(16384),
	}

	for _, benchmark := range benchmarks {
		b.Run(benchmark.scenario+"1024", func(b *testing.B) {
			benchmark.function(b, benchmark.codec, 1024, payload)
		})
		b.Run(benchmark.scenario+"4096", func(b *testing.B) {
			benchmark.function(b, benchmark.codec, 4096, payload)
		})
		b.Run(benchmark.scenario+"8192", func(b *testing.B) {
			benchmark.function(b, benchmark.codec, 8192, payload)
		})
		b.Run(benchmark.scenario+"16384", func(b *testing.B) {
			benchmark.function(b, benchmark.codec, 16384, payload)
		})
	}

}

func benchmarkCompression(b *testing.B, codec int8, payloadSize int, payload map[int][]byte) {
	msg := kafka.Message{
		Value:            payload[payloadSize],
		CompressionCodec: codec,
	}

	for i := 0; i < b.N; i++ {
		m1, err := msg.Encode()
		if err != nil {
			b.Fatal(err)
		}

		_, err = m1.Decode()
		if err != nil {
			b.Fatal(err)
		}

	}
}

const dataset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func randomPayload(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = dataset[rand.Intn(len(dataset))]
	}
	return b
}
