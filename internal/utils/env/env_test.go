// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

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
	require.Equal(t, "DEFAULT_ENV_VALUE", got)

	os.Setenv("TEST_ENV_KEY", "SET_ENV_VALUE")
	got = Lookup("TEST_ENV_KEY", "DEFAULT_ENV_VALUE")
	require.Equal(t, "SET_ENV_VALUE", got)

	os.Clearenv()
	got = Lookup("TEST_ENV_KEY", "DEFAULT_ENV_VALUE")
	require.Equal(t, "DEFAULT_ENV_VALUE", got)
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
	require.Equal(t, 1000, got)

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
	require.Equal(t, time.Second*10, got)

	os.Clearenv()
	got = Lookup("TEST_ENV_KEY", d)
	require.Equal(t, got, d)

	os.Setenv("TEST_ENV_KEY", "1axx")
	got = Lookup("TEST_ENV_KEY", d)
	require.Equal(t, got, d)
}
