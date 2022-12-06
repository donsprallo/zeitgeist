package ntp

import (
	"bytes"
	"testing"
)

func TestPackageToBytes(t *testing.T) {
	// Create test table; the ntp package will convert to bytes
	// and check that the result is euqal to data.
	tables := []struct {
		pkg  NtpPackage
		data []byte
	}{
		// Encode "HelloWorld" to hex codes
		{NtpPackage{
			Header:         0x4865_6C6C,
			RootDelay:      0x6F57_6F72,
			RootDispersion: 0x6C64_0000,
		}, []byte("HelloWorld")},
	}

	// Test all data in test table
	for _, table := range tables {
		b, err := table.pkg.ToBytes()

		// Check error value
		if err != nil {
			t.Fatalf("ntp package to bytes failed: %s", err)
		}

		// Check length of bytes
		if len(b) != NTP_PACKAGE_SIZE {
			t.Errorf("ntp package to bytes invalid size: %d", len(b))
		}

		// Check result equal to test value
		idx := bytes.IndexByte(b, 0x00)
		if !bytes.Equal(b[:idx], table.data) {
			t.Errorf("ntp package to bytes '%s' not equal to '%s'",
				b, table.data)
		}
	}
}

func TestPackageFromBytes(t *testing.T) {
	// Create test table; the ntp package will convert to bytes
	// and check that the result is euqal to data.
	tables := []struct {
		pkg  NtpPackage
		data []byte
	}{
		// Encode "HelloWorld" to hex codes
		{NtpPackage{
			Header:         0x4865_6C6C,
			RootDelay:      0x6F57_6F72,
			RootDispersion: 0x6C64_0000,
		}, []byte("HelloWorld")},
	}

	// Test all data in test table
	for _, table := range tables {
		// Copy test data to buffer with length of 48 byte length
		buffer := make([]byte, 48)
		copy(buffer[:], table.data)

		// Create ntp package from buffer with 48 byte length
		pkg, err := PackageFromBytes(buffer)

		// Check error value
		if err != nil {
			t.Fatalf("ntp package from bytes failed: %s", err)
		}

		// Check result equal to test value. We do not need to test
		// all fields here. We just test the first three values.
		if pkg.Header != table.pkg.Header {
			t.Errorf("ntp package from bytes '%X' not equal to '%X'",
				pkg.Header, &table.pkg.Header)
		}

		if pkg.RootDelay != table.pkg.RootDelay {
			t.Errorf("ntp package from bytes '%X' not equal to '%X'",
				pkg.RootDelay, &table.pkg.RootDelay)
		}

		if pkg.RootDispersion != table.pkg.RootDispersion {
			t.Errorf("ntp package from bytes '%X' not equal to '%X'",
				pkg.RootDispersion, &table.pkg.RootDispersion)
		}
	}
}

func TestGetLeapIndicator(t *testing.T) {
	// Create an test values array; the ntp package leap indicator is compared
	// with this test value. The values are constants from ntp package.
	values := []int{
		NTP_LI_NOT_SET,
		NTP_LI_SUB_SEC,
		NTP_LI_ADD_SEC,
		NTP_LI_NOT_SYN,
	}

	// Test all data in test values
	for _, value := range values {
		pkg := NtpPackage{}
		// First set leap indicator. Next we get and compare the
		// value we set. This must be the same here.
		pkg.SetLeap(uint32(value))
		leap := pkg.GetLeap()
		if leap != uint32(value) {
			t.Errorf("ntp get leap indicator value failed: %d != %d",
				leap, value)
		}
	}
}
