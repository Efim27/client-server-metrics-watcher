package statsreader

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleMetricsDump() {
	metricsDump, err := NewMetricsDump()
	if err != nil {
		log.Fatal(err)
	}

	metricsDump.Refresh()

	fmt.Println(metricsDump.MetricsGauge)
	fmt.Println(metricsDump.MetricsCounter)
}

func TestRefresh(t *testing.T) {
	metricsDump, err := NewMetricsDump()
	assert.NoError(t, err)

	metricsDump.Refresh()
	metricsDump.Refresh()
	metricsDump.Refresh()

	assert.Equal(t, 3, int(metricsDump.MetricsCounter["PollCount"]))
}
