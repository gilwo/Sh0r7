package metrics

import (
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestMetrics1(t *testing.T) {
	g1 := NewMetricGlobal().IncFailedShortCreateCounter().IncInvalidShortAccessCounter().IncServedPathCount().
		IncShortAccessVisitDeleteCount()
	g1.IncShortAccessVisitCount().IncShortAccessVisitCount()

	if err := g1.ToMap().Encode().Error(); err != nil {
		t.Logf("encode error: %s\n", err)
		t.FailNow()
	}

	// g2 := NewMetricGlobal()
	g2 := NewMetricGroup(MetricGroupGlobal)
	g2.FromString(g1.EncodedString()).Decode()
	// log.Printf("z: %T - %v\n", g2, g2.Error())
	// log.Printf("z dump:\n%s\n", g2.DumpMap())
	// log.Printf("z dump:\n%s\n", g2.DumpObject())
	if err := g2.ToObject().Error(); err != nil {
		t.Logf("decode error: %s\n", err)
		t.FailNow()
	}

	if !g1.Equal(g2) {
		t.Logf("g1:\n%s\n", g1.DumpObject())
		t.Logf("g2:\n%s\n", g2.DumpObject())
		t.Logf("g1 == g2 : %v\n", g1.Equal(g2))
		t.FailNow()
	}
}

func TestMetrics2(t *testing.T) {
	g1 := NewMetricShortCreationFailure()
	g1.FailedShortCreateShort = "short name"
	g1.FailedShortCreateIP = "1.1.1.1"
	g1.FailedShortCreateInfo = "info ... ++++ "
	g1.FailedShortCreateReferrer = "2.2.2.2 referrer"
	g1.FailedShortCreateTime = time.Now().String()
	g1.FailedShortCreateReason = errors.Errorf("bad something .. ").Error()

	if err := g1.ToMap().Encode().Error(); err != nil {
		t.Logf("encode error: %s\n", err)
		t.FailNow()
	}
	// log.Println("dump: " + g1.DumpMap())
	// log.Println("string: " + g1.EncodedString())
	// log.Printf("g1: %+#v\n", g1)

	// g2 := NewMetricShortCreationFailure()
	g2 := NewMetricGroup(MetricGroupShortCreationFailure)
	g2.FromString(g1.EncodedString()).Decode()
	if err := g2.ToObject().Error(); err != nil {
		t.Logf("decode error: %s\n", err)
		t.Logf("encoded string: %s\n", g1.EncodedString())
		t.FailNow()
	}

	if !g1.Equal(g2) {
		t.Logf("g1:\n%s\n", g1.DumpObject())
		t.Logf("g2:\n%s\n", g2.DumpObject())
		t.Logf("g1 == g2 : %v\n", g1.Equal(g2))
		t.FailNow()
	}

}
func TestMetrics3(t *testing.T) {

	for groupType, metric1 := range metricsParam() {

		t.Logf("testing %s\n", metric1.Name())
		if err := metric1.ToMap().Encode().Error(); err != nil {
			t.Logf("encode error: %s\n", err)
			t.FailNow()
		}
		// log.Println("dump: " + g1.DumpMap())
		// log.Println("string: " + g1.EncodedString())
		// log.Printf("g1: %+#v\n", g1)

		// g2 := NewMetricShortCreationFailure()
		metric2 := NewMetricGroup(groupType)
		if err := metric2.FromString(metric1.EncodedString()).Decode().Error(); err != nil {
			t.Logf("decode error: %s\n", err)
			t.Logf("encoded string: %s\n", metric1.EncodedString())
			t.FailNow()
		}
		if err := metric2.ToObject().Error(); err != nil {
			t.Logf("convert to object error: %s\n", err)
			t.Logf("encoded string: %s\n", metric1.EncodedString())
			t.FailNow()
		}

		if !metric1.Equal(metric2) {
			t.Logf("metric1:\n%s\n", metric1.DumpObject())
			t.Logf("metric2:\n%s\n", metric2.DumpObject())
			t.Logf("metric1 == metric2 : %v\n", metric1.Equal(metric2))
			t.FailNow()
		}
	}
}

func metricsParam() map[MetricGroupType]MetricPacker {
	r := map[MetricGroupType]MetricPacker{}

	g1 := NewMetricGlobal().IncFailedShortCreateCounter().IncInvalidShortAccessCounter().IncServedPathCount().
		IncShortAccessVisitDeleteCount().IncShortAccessVisitCount().IncShortAccessVisitCount()

	r[MetricGroupGlobal] = g1

	g2 := NewMetricShortCreationFailure()
	g2.FailedShortCreateShort = "short name"
	g2.FailedShortCreateIP = "1.1.1.1"
	g2.FailedShortCreateInfo = "info ... ++++ "
	g2.FailedShortCreateReferrer = "2.2.2.2 referrer"
	g2.FailedShortCreateTime = time.Now().String()
	g2.FailedShortCreateReason = errors.Errorf("bad something .. ").Error()

	r[MetricGroupShortCreationFailure] = g2

	g3 := NewMetricShortCreationSuccess()
	g3.ShortCreateName = "hjfgds435h"
	g3.ShortCreateTime = time.Now().String()
	g3.ShortCreateIP = "1.1.1.1"
	g3.ShortCreateInfo = "info ... ++++ "
	g3.ShortCreatePrivate = "true"
	g3.ShortCreateDelete = "false"
	g3.ShortCreatedNamed = "true"
	g3.ShortCreateReferrer = "2.2.2.2 referrer"

	r[MetricGroupShortCreationSuccess] = g3

	return r
}
