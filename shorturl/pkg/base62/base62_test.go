package base62

import (
	"testing"
)

// TestInt2String 测试十进制转62进制
func TestInt2String(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"zero", 0, "0"},
		{"one", 1, "1"},
		{"ten", 10, "a"},
		{"eleven", 11, "b"},
		{"sixty-one", 61, "Z"},
		{"sixty-two", 62, "10"},
		{"sixty-three", 63, "11"},
		{"thirty-five", 35, "z"},
		{"thirty-six", 36, "A"},
		{"max_single_digit", 61, "Z"},
		{"two_digits", 62, "10"},
		{"large_number", 62*62 + 10, "10a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Int2String(tt.input)
			if result != tt.expected {
				t.Errorf("Int2String(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestReverse 测试反转函数
func TestReverse(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{"empty", []byte{}, []byte{}},
		{"single", []byte{'a'}, []byte{'a'}},
		{"two", []byte{'a', 'b'}, []byte{'b', 'a'}},
		{"three", []byte{'a', 'b', 'c'}, []byte{'c', 'b', 'a'}},
		{"base62_result", []byte{'0', '1'}, []byte{'1', '0'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reverse(tt.input)
			if string(result) != string(tt.expected) {
				t.Errorf("reverse(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}