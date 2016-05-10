package statsd

import (
	//	"fmt"
	"testing"
)

/*
func TestEmpty(t *testing.T) {
	mr := &metricReceiver{}
	m, err := mr.parseLine([]byte{})
	fmt.Printf("Parsing empty byte array got back %v\n", err)
	if err == nil {
		t.Errorf("Attempting to parse empty byte slice and did not get error back. Result [%v]", m)
	}
}

func TestNil(t *testing.T) {
	mr := &metricReceiver{}
	m, err := mr.parseLine(nil)
	if err == nil {
		t.Errorf("Attempting to parse nil slice and did not get error back. Result [%v]", m)
	}
}
*/
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
