package ntp

import (
	"bytes"
	"testing"
)

func TestPackageToBytes(t *testing.T) {
	// Create test table; the ntp package will convert to bytes
	// and check that the result is equal to data.
	tables := []struct {
		pkg  Package
		data []byte
	}{
		// Encode "HelloWorld" to hex codes
		{Package{
			header:         0x4865_6C6C,
			rootDelay:      0x6F57_6F72,
			rootDispersion: 0x6C64_0000,
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
		if len(b) != PackageSize {
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
	// and check that the result is equal to data.
	tables := []struct {
		pkg  Package
		data []byte
	}{
		// Encode "HelloWorld" to hex codes
		{Package{
			header:         0x4865_6C6C,
			rootDelay:      0x6F57_6F72,
			rootDispersion: 0x6C64_0000,
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
		if pkg.header != table.pkg.header {
			t.Errorf("ntp package from bytes '%X' not equal to '%X'",
				pkg.header, &table.pkg.header)
		}

		if pkg.rootDelay != table.pkg.rootDelay {
			t.Errorf("ntp package from bytes '%X' not equal to '%X'",
				pkg.rootDelay, &table.pkg.rootDelay)
		}

		if pkg.rootDispersion != table.pkg.rootDispersion {
			t.Errorf("ntp package from bytes '%X' not equal to '%X'",
				pkg.rootDispersion, &table.pkg.rootDispersion)
		}
	}
}

func TestSetGetLeapIndicator(t *testing.T) {
	// Create a test values array; the ntp package leap indicator is compared
	// with this test value. The values are constants from ntp package.
	values := []uint32{
		LeapNotSet,
		LeapSubSec,
		LeapAddSec,
		LeapNotSyn,
	}

	// Test all data in test values
	for _, value := range values {
		pkg := Package{}
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
	// Create a test values array; the ntp package version is compared
	// with this test value.
	values := []uint32{
		VersionV3,
		VersionV4,
	}

	// Test all data in test values
	for _, value := range values {
		pkg := Package{}
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
	// Create a test values array; the ntp package version is compared
	// with this test value.
	values := []uint32{
		ModeReserved,
		ModeSymActive,
		ModeSymPassive,
		ModeClient,
		ModeServer,
		ModeBroadcast,
		ModeControl,
		ModePrivate,
	}

	// Test all data in test values
	for _, value := range values {
		pkg := Package{}
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
	// Create a test values array; the ntp package version is compared
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
		pkg := Package{}
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
