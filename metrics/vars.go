package metrics

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

var (
	// MetricGlobalCounter *MetricGlobal
	// MetricProcessor     *metricContext

	GlobalMeter *MeterCounter
)

// using uptrace :
// func init() {
// 	GlobalMeter = NewMeterCounter("sh0r7.global.counters",
// 		CreationSuccess,
// 		CreationFailure,
// 		DeletionCounter,
// 		DeletionExpired,
// 		DeletionRemoved,
// 		InvalidShort,
// 		VisitPublic,
// 		VisitPrivate,
// 		ForbiddenPublic,
// 		ForbiddenPrivate,
// 		ForbiddenRemove,
// 		ServedPathCreate,
// 		ServedPathPublic,
// 		ServedPathPrivate,
// 		ServedPathRemove,
// 	)

// 	panic("asd")
// }

type MeterCounter struct {
	meterCtx       metric.Meter
	name           string
	counters       *sync.Map // - map[string]instrument.Int64ObservableCounter
	countersRegist *sync.Map // - map[string]instrument.Int64ObservableCounter
	meterNumbers   *sync.Map // - map[instrument.Int64ObservableCounter]int64
}

const (
	Created             = "Created"
	CreationFailed      = "CreationFailed"      // consider add what failure counter?!?
	CreationFailedToken = "CreationFailedToken" // consider add what failure counter?!?
	ShortExpired        = "ShortExpired"
	ShortRemoved        = "ShortRemoved"
	RemoveFailed        = "RemoveFailed"
	InvalidShort        = "InvalidShort"
	VisitPublic         = "VisitPublic"
	PublicFailed        = "PublicFailed"
	VisitPrivate        = "VisitPrivate"
	PrivateFailed       = "PrivateFailed"
	PublicNotAuth       = "PublicNotAuth"
	PrivateNotAuth      = "PrivateNotAuth"
	RemoveNotAuth       = "RemoveNotAuth"
	ServedPathCreate    = "ServedPathCreate"
	ServedPathPublic    = "ServedPathPublic"
	ServedPathPrivate   = "ServedPathPrivate"
	ServedPathRemove    = "ServedPathRemove"
)

func NewMeterCounter(name string, elements ...string) *MeterCounter {
	var err error
	r := &MeterCounter{
		name: name,
		// counters:     map[string]instrument.Int64ObservableCounter{},
		// meterNumbers: map[instrument.Int64ObservableCounter]int64{},
		counters:       &sync.Map{},
		countersRegist: &sync.Map{},
		meterNumbers:   &sync.Map{},
	}
	r.meterCtx = global.MeterProvider().Meter(name)
	for _, e := range elements {
		var x instrument.Int64Observable
		x, err = r.meterCtx.Int64ObservableCounter(
			name+"."+e,
			instrument.WithUnit("1"),
			instrument.WithDescription(e),
		)
		// r.counters[e] = x
		r.counters.Store(e, x)
		if err != nil {
			panic(errors.Wrapf(err, "failed to create counter for %s", e))
		}
		// r.meterNumbers[r.counters[e]] = 0
		r.meterNumbers.Store(x, int64(0))
	}
	// for k, v := range r.counters {
	r.counters.Range(func(key, value any) bool {

		counterName := key.(string)
		counter := value.(instrument.Int64ObservableCounter)
		regist, err := r.meterCtx.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
			v, ok := r.meterNumbers.Load(counter)
			if !ok {
				panic("key not exists in map")
			}
			o.ObserveInt64(counter, v.(int64))
			// v, ok := r.meterNumbers2.Load(counter)
			// if !ok {
			// 	panic("key not exits in map")
			// }
			// o.ObserveInt64(counter, v.(int64))
			return nil
		}, counter)
		if err != nil {
			panic(errors.Wrapf(err, "failed to register meter callback for %s (%s)", name, counterName))
		}
		r.countersRegist.Store(counterName, regist)
		// if err = r.meterCtx.RegisterCallback([]instrument.Asynchronous{counter},
		// 	func(ctx context.Context) {
		// 		counter.Observe(ctx, r.meterNumbers[counter])
		// 		// log.Printf("callback invoked for %s\n", counterName)
		// 	},
		// ); err != nil {
		// 	panic(errors.Wrapf(err, "failed to register meter callback for %s (%s)", name, counterName))
		// }
		// }
		return true
	})
	return r
}

func (mc *MeterCounter) Dump() string {
	s := ""
	// for k, e := range mc.counters {
	// 	s += fmt.Sprintf("%s: %d\n", k, mc.meterNumbers[e])
	// }
	mc.counters.Range(func(key, value any) bool {
		counterName := key.(string)
		counter := value.(instrument.Int64ObservableCounter)
		v, ok := mc.meterNumbers.Load(counter)
		if !ok {
			panic("key not found in map")
		}
		s += fmt.Sprintf("%s: %d\n", counterName, v)
		return true
	})
	// x, ok := mc.counters.Load("")
	// y := x.(instrument.Int64ObservableCounter)
	return s
}

func (mc *MeterCounter) IncMeterCounter(name string) {
	// if v, ok := mc.counters[name]; ok {
	// 	if _, ok = mc.meterNumbers[v]; ok {
	// 		mc.meterNumbers[v] = mc.meterNumbers[v] + 1
	// 		return
	// 	}
	// }
	v, ok := mc.counters.Load(name)
	if ok {
		vCounter := v.(instrument.Int64ObservableCounter)
		v, ok = mc.meterNumbers.Load(vCounter)
		if ok {
			vNum := v.(int64)
			mc.meterNumbers.Store(vCounter, vNum+1)
		}
	}
	// log.Printf("error on metric counter %s:%s\n", mc.name, name)
}

func InitGlobalMeter(name string) *MeterCounter {
	return NewMeterCounter(name+".global.counters",
		Created,
		CreationFailed,
		CreationFailedToken,
		ShortExpired,
		ShortRemoved,
		RemoveFailed,
		InvalidShort,
		VisitPublic,
		PublicFailed,
		VisitPrivate,
		PrivateFailed,
		PublicNotAuth,
		PrivateNotAuth,
		RemoveNotAuth,
		ServedPathCreate,
		ServedPathPublic,
		ServedPathPrivate,
		ServedPathRemove,
	)
}
