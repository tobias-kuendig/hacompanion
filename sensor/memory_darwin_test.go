package sensor

import (
	"testing"

	"hacompanion/entity"

	"github.com/stretchr/testify/require"
)

func TestMemory(t *testing.T) {
	topOutput := `PhysMem: 6144M used (2048M wired, 1024M compressor), 2048M unused.`
	output := &entity.Payload{
		State: float64(6144),
		Attributes: map[string]interface{}{
			"mem_total": float64(8192),
		},
	}

	m := NewMemory()

	res, err := m.process(topOutput)
	require.NoError(t, err)
	require.EqualValues(t, output, res)
}
