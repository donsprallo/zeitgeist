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

func TestSetGetLeapIndicator(t *testing.T) {
	// Create an test values array; the ntp package leap indicator is compared
	// with this test value. The values are constants from ntp package.
	values := []uint32{
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
		pkg.SetLeap(value)
		leap := pkg.GetLeap()
		if leap != value {
			t.Errorf("ntp get leap indicator value failed: %d != %d",
				leap, value)
		}
	}
}

func TestSetGetVersion(t *testing.T) {
	// Create an test values array; the ntp package version is compared
	// with this test value.
	values := []uint32{
		NTP_VN_V3,
		NTP_VN_V4,
	}

	// Test all data in test values
	for _, value := range values {
		pkg := NtpPackage{}
		// First set version. Next we get and compare the
		// value we set. This must be the same here.
		pkg.SetVersion(value)
		version := pkg.GetVersion()
		if version != value {
			t.Errorf("ntp get version value failed: %d != %d",
				version, value)
		}
	}
}

func TestSetGetMode(t *testing.T) {
	// Create an test values array; the ntp package version is compared
	// with this test value.
	values := []uint32{
		NTP_MODE_RESERVED,
		NTP_MODE_SYM_ACTIVE,
		NTP_MODE_SYM_PASSIVE,
		NTP_MODE_CLIENT,
		NTP_MODE_SERVER,
		NTP_MODE_BROADCAST,
		NTP_MODE_CONTROL,
		NTP_MODE_PRIVATE,
	}

	// Test all data in test values
	for _, value := range values {
		pkg := NtpPackage{}
		// First set mode. Next we get and compare the
		// value we set. This must be the same here.
		pkg.SetMode(value)
		mode := pkg.GetMode()
		if mode != value {
			t.Errorf("ntp get mode value failed: %d != %d",
				mode, value)
		}
	}
}

func TestSetGetStratum(t *testing.T) {
	// Create an test values array; the ntp package version is compared
	// with this test value.
	values := []uint32{
		0,
		1,
		127,
		254,
		255,
	}

	// Test all data in test values
	for _, value := range values {
		pkg := NtpPackage{}
		// First set stratum. Next we get and compare the
		// value we set. This must be the same here.
		pkg.SetStratum(value)
		stratum := pkg.GetStratum()
		if stratum != value {
			t.Errorf("ntp get stratum value failed: %d != %d",
				stratum, value)
		}
	}
}
