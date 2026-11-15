package secret

import (
	"strings"
	"testing"
)

func TestSealerBindsAssociatedDataAndVersion(t *testing.T) {
	s, err := NewSealer(strings.Repeat("ab", 32), 1)
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := s.Seal([]byte("page-token"), []byte("connection-a"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(sealed), "page-token") {
		t.Fatal("plaintext leaked")
	}
	plain, err := s.Open(sealed, 1, []byte("connection-a"))
	if err != nil || string(plain) != "page-token" {
		t.Fatal("round trip failed")
	}
	if _, err := s.Open(sealed, 1, []byte("connection-b")); err == nil {
		t.Fatal("ciphertext opened for another row")
	}
	if _, err := s.Open(sealed, 2, []byte("connection-a")); err == nil {
		t.Fatal("unknown key version accepted")
	}
	for _, key := range []string{"", "short", strings.Repeat("zz", 32)} {
		if _, err := NewSealer(key, 1); err == nil {
			t.Fatal("invalid key accepted")
		}
	}
}
