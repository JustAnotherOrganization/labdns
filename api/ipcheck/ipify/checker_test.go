package ipify_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "justanother.org/labdns/api/ipcheck/ipify"
)

func TestGetIP(t *testing.T) {
	knownIP, ok := os.LookupEnv(`TEST_KNOWN_IP`)
	if !ok {
		t.Skip("known IP not provided")
	}

	ch, _ := New()
	ip, err := ch.GetIP(context.Background())
	require.NoError(t, err)
	assert.Equal(t, knownIP, ip)
}
