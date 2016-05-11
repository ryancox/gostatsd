package statsd

import (
	//	"fmt"
	"testing"
)

func TestRoundtrip(t *testing.T) {
	//metric := &types.Metric{Name: "foo.bar.baz", Value: 2, Type: types.COUNTER}
	line := []byte("foo.bar.baz:2|c")
	mr := &metricReceiver{}
	parsedMetric, err := mr.parseLine(line)
	if err != nil {
		t.Errorf("error: %s", err)
		return
	}
	renderedLine, err := renderLine(parsedMetric)
	if err != nil {
		t.Errorf("error: %s", err)
	}
	if string(line) != renderedLine {
		t.Errorf("expected %s error but got %s", line, renderedLine)
		return
	}
	// TODO: parse parsedLine and compare via deepcopy to original
}
