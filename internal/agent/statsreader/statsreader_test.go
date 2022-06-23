package statsreader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefresh(t *testing.T) {
	var memStatistics MetricsDump
	memStatistics.Refresh()
	memStatistics.Refresh()
	memStatistics.Refresh()

	assert.Equal(t, 3, int(memStatistics.PollCount))
}
