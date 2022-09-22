package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCertgenCommand(t *testing.T) {
	got := getCertGenCommand()
	assert.Equal(t, "certgen", got.Use)
}
