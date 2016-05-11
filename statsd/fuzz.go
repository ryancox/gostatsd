package statsd

import (
	"fmt"

	"github.com/atlassian/gostatsd/types"
)

func shortName(mt types.MetricType) string {
	switch {
	case mt >= types.SET:
		return "s"
	case mt >= types.GAUGE:
		return "g"
	case mt >= types.TIMER:
		return "ms"
	case mt >= types.COUNTER:
		return "c"
	}
	return "?"
}

func renderLine(m *types.Metric) (string, error) {
	if m == nil {
		return "", fmt.Errorf("Invalid input")
	}

	var value interface{}
	if m.Type == types.SET {
		value = m.StringValue
	} else {
		value = m.Value
	}
	ret := fmt.Sprintf("%v:%v|%s", m.Name, value, shortName(m.Type))
	if m.Tags != nil {
		ret = fmt.Sprintf("%v|#%v", ret, m.Tags)
	}
	return ret, nil
}

func Fuzz(data []byte) int {
	mr := &metricReceiver{}
	metric, err := mr.parseLine(data)
	if err != nil {
		if metric != nil {
			panic("metric != nil on error")
		}
		return 0
	}
	renderLine(metric)
	return 1
}
