package sensor

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUptime(t *testing.T) {
	input := "kern.boottime: { sec = 1783072800, usec = 0 } Fri Jul  3 12:00:00 2026"
	expectedState := "2026-07-03T12:00:00+02:00"

	u := NewUptime()

	res, err := u.process(input)
	require.NoError(t, err)
	require.EqualValues(t, expectedState, res.State)

	uptimeSeconds, ok := res.Attributes["uptime_seconds"].(string)
	require.True(t, ok)
	_, err = strconv.ParseFloat(uptimeSeconds, 64)
	require.NoError(t, err)
}
