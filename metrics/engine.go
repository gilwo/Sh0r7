package metrics

import (
	"log"
)

type metricContext struct {
	process        chan (MetricPacker)
	quitProcessing chan (any)
	doneProcessing chan (any)
	dumpMetrics    bool
}

func NewMetricContext() *metricContext {
	return &metricContext{
		process:        make(chan MetricPacker, 100),
		quitProcessing: make(chan any),
		doneProcessing: make(chan any),
	}
}
func (mc *metricContext) EnableDisableDump() {
	mc.dumpMetrics = !mc.dumpMetrics
}
func (mc *metricContext) MetricDump(mt MetricPacker) {
	if mc.dumpMetrics {
		log.Printf("%s\n", mt.DumpObject())
	}
}

func (mc *metricContext) StopProcessing() {
	close(mc.quitProcessing)
	<-mc.doneProcessing
}

func (mc *metricContext) StartProcessing() {
	go func() {
		for {
			select {
			case <-mc.quitProcessing:
				close(mc.doneProcessing)
				return
			case mt := <-mc.process:
				if err := MDBctx.AddMetric(mt); err != nil {
					log.Printf("failed adding metric to db: %s\n", err)
					// TODO: consider save the failed metric in memory and update the db
					//  when it is ready again
				}
			}
		}
	}()
}

func (mc *metricContext) Add(metric MetricPacker) {
	mc.process <- metric
}
