package contract

import "testing"

func TestContext(t *testing.T) {
	if RequestUrlKey == IpKey {
		t.Fatal()
	}
}
