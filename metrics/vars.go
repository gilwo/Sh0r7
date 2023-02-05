package metrics

import (
	"context"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

var (
	GlobalMeter *MeterCounter
)

type MeterCounter struct {
	meterCtx     metric.Meter
	name         string
	counters     map[string]instrument.Int64ObservableCounter
	meterNumbers map[instrument.Int64ObservableCounter]int64
}

const (
	Created           = "Created"
	CreationFailed    = "CreationFailed" // consider add what failure counter?!?
	ShortExpired      = "ShortExpired"
	ShortRemoved      = "ShortRemoved"
	RemoveFailed      = "RemoveFailed"
	InvalidShort      = "InvalidShort"
	VisitPublic       = "VisitPublic"
	PublicFailed      = "PublicFailed"
	VisitPrivate      = "VisitPrivate"
	PrivateFailed     = "PrivateFailed"
	PublicNotAuth     = "PublicNotAuth"
	PrivateNotAuth    = "PrivateNotAuth"
	RemoveNotAuth     = "RemoveNotAuth"
	ServedPathCreate  = "ServedPathCreate"
	ServedPathPublic  = "ServedPathPublic"
	ServedPathPrivate = "ServedPathPrivate"
	ServedPathRemove  = "ServedPathRemove"
)

func NewMeterCounter(name string, elements ...string) *MeterCounter {
	var err error
	r := &MeterCounter{
		name:         name,
		counters:     map[string]instrument.Int64ObservableCounter{},
		meterNumbers: map[instrument.Int64ObservableCounter]int64{},
	}
	r.meterCtx = global.MeterProvider().Meter(name)
	for _, e := range elements {
		r.counters[e], err = r.meterCtx.Int64ObservableCounter(
			name+"."+e,
			instrument.WithUnit("1"),
			instrument.WithDescription(e),
		)
		if err != nil {
			panic(errors.Wrapf(err, "failed to create counter for %s", e))
		}
		r.meterNumbers[r.counters[e]] = 0
	}
	for k, v := range r.counters {
		counterName := k
		counter := v
		regist, err := r.meterCtx.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
			o.ObserveInt64(counter, r.meterNumbers[counter])
			return nil
		}, counter)
		if err != nil {
			panic(errors.Wrapf(err, "failed to register meter callback for %s (%s)", name, counterName))
		}
		regist = regist // TODO: FIXME ... handle properly
	}
	return r
}

func (mc *MeterCounter) Dump() string {
	s := ""
	for k, e := range mc.counters {
		s += fmt.Sprintf("%s: %d\n", k, mc.meterNumbers[e])
	}
	return s
}

func (mc *MeterCounter) IncMeterCounter(name string) {
	if v, ok := mc.counters[name]; ok {
		if _, ok = mc.meterNumbers[v]; ok {
			mc.meterNumbers[v] = mc.meterNumbers[v] + 1
			return
		}
	}
	log.Printf("error on metric counter %s:%s\n", mc.name, name)
}

func InitGlobalMeter(name string) *MeterCounter {
	return NewMeterCounter(name+".global.counters",
		Created,
		CreationFailed,
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
