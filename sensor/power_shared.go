package sensor

import (
	"fmt"
	"strconv"
	"strings"
)

type Power struct {
	Battery string
}

func (pwr Power) resolveIcon(state any, charging bool) string {
	num, err := strconv.Atoi(strings.TrimSpace(state.(string)))
	if err != nil {
		return "mdi:battery-unknown"
	}

	var level int
	switch {
	case num >= 90:
		level = 100
	case num >= 80:
		level = 80
	case num >= 70:
		level = 70
	case num >= 60:
		level = 60
	case num >= 50:
		level = 50
	case num >= 40:
		level = 40
	case num >= 30:
		level = 30
	case num >= 20:
		level = 20
	default:
		level = 10
	}

	if level == 100 {
		return "mdi:battery"
	}

	if charging {
		return fmt.Sprintf("mdi:battery-charging-%d", level)
	}

	return fmt.Sprintf("mdi:battery-%d", level)
}
