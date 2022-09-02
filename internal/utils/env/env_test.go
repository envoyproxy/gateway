package env

import (
	"crypto/rand"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLookupString(t *testing.T) {
	defer os.Clearenv()

	got := Lookup("TEST_ENV_KEY", "DEFAULT_ENV_VALUE")
	require.Equal(t, got, "DEFAULT_ENV_VALUE")

	os.Setenv("TEST_ENV_KEY", "SET_ENV_VALUE")
	got = Lookup("TEST_ENV_KEY", "DEFAULT_ENV_VALUE")
	require.Equal(t, got, "SET_ENV_VALUE")

	os.Clearenv()
	got = Lookup("TEST_ENV_KEY", "DEFAULT_ENV_VALUE")
	require.Equal(t, got, "DEFAULT_ENV_VALUE")
}

func TestLookupInt(t *testing.T) {
	defer os.Clearenv()

	intVal, err := rand.Int(rand.Reader, big.NewInt(1000))
	require.NoError(t, err)

	i := int(intVal.Int64())
	got := Lookup("TEST_ENV_KEY", i)
	require.Equal(t, got, i)

	os.Setenv("TEST_ENV_KEY", "1000")
	got = Lookup("TEST_ENV_KEY", i)
	require.Equal(t, got, 1000)

	os.Clearenv()
	got = Lookup("TEST_ENV_KEY", i)
	require.Equal(t, got, i)

	os.Clearenv()
	os.Setenv("TEST_ENV_KEY", "10s")
	got = Lookup("TEST_ENV_KEY", i)
	require.Equal(t, got, i)
}

func TestLookupDuration(t *testing.T) {
	defer os.Clearenv()

	d := time.Second * 1000
	got := Lookup("TEST_ENV_KEY", d)
	require.Equal(t, got, d)

	os.Setenv("TEST_ENV_KEY", "10s")
	got = Lookup("TEST_ENV_KEY", d)
	require.Equal(t, got, time.Second*10)

	os.Clearenv()
	got = Lookup("TEST_ENV_KEY", d)
	require.Equal(t, got, d)

	os.Setenv("TEST_ENV_KEY", "1axx")
	got = Lookup("TEST_ENV_KEY", d)
	require.Equal(t, got, d)
}
