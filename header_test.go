package go_stomp_websocket

import (
	"testing"
)

func TestHeader_Contains(t *testing.T) {
	tests := []struct {
		name    string
		header  *Header
		key     string
		wantVal string
		wantOk  bool
	}{
		{
			name:    "empty header",
			header:  testHeader(),
			key:     "test",
			wantVal: "",
			wantOk:  false,
		},
		{
			name:    "key exists",
			header:  testHeader("key1", "value1", "key2", "value2"),
			key:     "key1",
			wantVal: "value1",
			wantOk:  true,
		},
		{
			name:    "key doesn't exist",
			header:  testHeader("key1", "value1", "key2", "value2"),
			key:     "key3",
			wantVal: "",
			wantOk:  false,
		},
		{
			name:    "case sensitive",
			header:  testHeader("Key1", "value1"),
			key:     "key1",
			wantVal: "",
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := tt.header.Contains(tt.key)
			if gotVal != tt.wantVal {
				t.Errorf("Header.Contains() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Header.Contains() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestHeader_Add(t *testing.T) {
	tests := []struct {
		name   string
		header *Header
		key    string
		value  string
		want   []string
	}{
		{
			name:   "add to empty header",
			header: testHeader(),
			key:    "key1",
			value:  "value1",
			want:   []string{"key1", "value1"},
		},
		{
			name:   "add to existing header",
			header: testHeader("key1", "value1"),
			key:    "key2",
			value:  "value2",
			want:   []string{"key1", "value1", "key2", "value2"},
		},
		{
			name:   "add empty key",
			header: testHeader(),
			key:    "",
			value:  "value1",
			want:   []string{"", "value1"},
		},
		{
			name:   "add empty value",
			header: testHeader(),
			key:    "key1",
			value:  "",
			want:   []string{"key1", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.header.Add(tt.key, tt.value)
			if len(tt.header.header) != len(tt.want) {
				t.Errorf("Header.Add() got length = %v, want %v", len(tt.header.header), len(tt.want))
				return
			}
			for i := range tt.want {
				if tt.header.header[i] != tt.want[i] {
					t.Errorf("Header.Add() got[%d] = %v, want %v", i, tt.header.header[i], tt.want[i])
				}
			}
		})
	}
}

func TestHeader_Get(t *testing.T) {
	tests := []struct {
		name   string
		header *Header
		key    string
		want   string
	}{
		{
			name:   "empty header",
			header: testHeader(),
			key:    "test",
			want:   "",
		},
		{
			name:   "key exists",
			header: testHeader("key1", "value1", "key2", "value2"),
			key:    "key1",
			want:   "value1",
		},
		{
			name:   "key doesn't exist",
			header: testHeader("key1", "value1", "key2", "value2"),
			key:    "key3",
			want:   "",
		},
		{
			name:   "case sensitive",
			header: testHeader("Key1", "value1"),
			key:    "key1",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.header.Get(tt.key); got != tt.want {
				t.Errorf("Header.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHeader_index(t *testing.T) {
	tests := []struct {
		name   string
		header *Header
		key    string
		want   int
		wantOk bool
	}{
		{
			name:   "empty header",
			header: testHeader(),
			key:    "test",
			want:   -1,
			wantOk: false,
		},
		{
			name:   "key exists",
			header: testHeader("key1", "value1", "key2", "value2"),
			key:    "key1",
			want:   0,
			wantOk: true,
		},
		{
			name:   "key doesn't exist",
			header: testHeader("key1", "value1", "key2", "value2"),
			key:    "key3",
			want:   -1,
			wantOk: false,
		},
		{
			name:   "case sensitive",
			header: testHeader("Key1", "value1"),
			key:    "key1",
			want:   -1,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := tt.header.index(tt.key)
			if got != tt.want {
				t.Errorf("Header.index() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Header.index() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

// testHeader creates a new Header with the given key-value pairs
func testHeader(pairs ...string) *Header {
	return &Header{header: pairs}
}
