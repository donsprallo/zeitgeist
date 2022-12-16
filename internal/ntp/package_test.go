package ntp

import (
	"bytes"
	"testing"
	"time"
)

func TestToTimestamp(t *testing.T) {
	// Create test data table.
	values := []time.Time{
		time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2038, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	// Test all entries in test table.
	for idx, e := range values {
		ts := ToTimestamp(e)

		// Calculate seconds part.
		testS := uint32(e.Unix()) + TimeDelta

		// Calculate fractional part.
		micros := float64(e.UnixMicro())
		factor := (1 << 32) * (1.0e-6)
		testF := uint32(micros * factor)

		// Test calculated results.
		if ts.Seconds != testS {
			t.Errorf("[%d] incorrect secs from TimestampToSeconds", idx)
		}

		if ts.Fraction != testF {
			t.Errorf("[%d] incorrect fracs from TimestampToSeconds", idx)
		}
	}
}

func TestToTime(t *testing.T) {
	// Create test data table.
	table := []struct {
		timestamp Timestamp
		datetime  time.Time
	}{
		{
			Timestamp{
				Seconds:  1671180400 + TimeDelta,
				Fraction: 4096,
			}, time.Date(
				2022, time.December, 16, 8, 46, 40, 4096, time.UTC),
		},
		{
			Timestamp{
				Seconds:  1706742000 + TimeDelta,
				Fraction: 0,
			}, time.Date(
				2024, time.January, 31, 23, 0, 0, 0, time.UTC),
		},
		{
			Timestamp{
				Seconds:  1528596244 + TimeDelta,
				Fraction: 0,
			}, time.Date(
				2018, time.June, 10, 2, 4, 4, 0, time.UTC),
		},
		{
			Timestamp{
				Seconds:  1907287444 + TimeDelta,
				Fraction: 0,
			}, time.Date(
				2030, time.June, 10, 2, 4, 4, 0, time.UTC),
		},
	}

	// Test all entries in test table.
	for idx, e := range table {
		ts := ToTime(e.timestamp)
		// Test conversion result.
		if ts != e.datetime {
			t.Errorf("[%d] incorrect timestamp conversion %s != %s",
				idx, ts.String(), e.datetime.String())
		}
	}
}

func TestTimeConversion(t *testing.T) {
	// Create test data table.
	values := []time.Time{
		// TODO: This case is not handled now.
		// time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2038, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	// Test all entries in test table.
	for idx, e := range values {
		ts := ToTimestamp(e)
		tv := ToTime(ts)

		if tv != e {
			t.Errorf("[%d] incorrect timestamp conversion %s != %s",
				idx, tv.String(), e.String())
		}
	}
}

func TestPackageToBytes(t *testing.T) {
	// Create test e; the ntp package will convert to bytes
	// and check that the result is equal to data.
	table := []struct {
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

	// Test all data in test e
	for _, e := range table {
		b, err := e.pkg.ToBytes()

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
		if !bytes.Equal(b[:idx], e.data) {
			t.Errorf("ntp package to bytes '%s' not equal to '%s'",
				b, e.data)
		}
	}
}

func TestPackageFromBytes(t *testing.T) {
	// Create test e; the ntp package will convert to bytes
	// and check that the result is equal to data.
	table := []struct {
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

	// Test all data in test e
	for _, e := range table {
		// Copy test data to buffer with length of 48 byte length
		buffer := make([]byte, 48)
		copy(buffer[:], e.data)

		// Create ntp package from buffer with 48 byte length
		pkg, err := PackageFromBytes(buffer)

		// Check error value
		if err != nil {
			t.Fatalf("ntp package from bytes failed: %s", err)
		}

		// Check result equal to test value. We do not need to test
		// all fields here. We just test the first three values.
		if pkg.header != e.pkg.header {
			t.Errorf("ntp package from bytes '%X' not equal to '%X'",
				pkg.header, &e.pkg.header)
		}

		if pkg.rootDelay != e.pkg.rootDelay {
			t.Errorf("ntp package from bytes '%X' not equal to '%X'",
				pkg.rootDelay, &e.pkg.rootDelay)
		}

		if pkg.rootDispersion != e.pkg.rootDispersion {
			t.Errorf("ntp package from bytes '%X' not equal to '%X'",
				pkg.rootDispersion, &e.pkg.rootDispersion)
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
