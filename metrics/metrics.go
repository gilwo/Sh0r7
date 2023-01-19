package metrics

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
)

type MetricType int

const (
	InvalidMetrtic MetricType = iota
	// =============================================
	// metrics per invalid short (creation / access)
	// ---------------------------------------------

	// failed short creation - per failure
	FailedShortCreateShort    // string // short name
	FailedShortCreateTime     // string
	FailedShortCreateIP       // string
	FailedShortCreateInfo     // string // useragent / other ...
	FailedShortCreateReferrer // string
	FailedShortCreateReason   // string // if applicable ..?!?

	// global failed creation
	FailedShortCreateCounter //int

	// invalid short access (non existant)
	InvalidShortAccessShort    // string // short name
	InvalidShortAccessTime     // string
	InvalidShortAccessIP       // string
	InvalidShortAccessInfo     // string // useragent / other ...
	InvalidShortAccessReferrer // string

	// global invalid access
	InvalidShortAccessCounter // int

	// =====================================
	// metrics per short (creation / access)
	// -------------------------------------

	// Short Creation - only once on creation
	ShortCreateName     // short name
	ShortCreateTime     // string
	ShortCreateIP       // string // ip
	ShortCreateInfo     // string // useragaent / user id / other ...
	ShortCreatePrivate  // bool
	ShortCreateDelete   // bool
	ShortCreatedNamed   // bool
	ShortCreateReferrer // string

	// global created / expired shorts
	ShortCreatedCount // int
	ShortExpiredCount // int

	// Short Access - per visit metrics
	ShortAccessVisitTime     // string
	ShortAccessVisitIP       // string
	ShortAccessVisitInfo     // string // useragent / other ...
	ShortAccessVisitReferrer // string
	ShortAccessVisitSuccess  // bool
	ShortAccessVisitPrivate  // bool
	ShortAccessVisitDelete   // bool
	ShortAccessVisitIsLocked // bool

	// global Access counter - for all visits
	ShortAccessVisitCount       // int
	ShortAccessVisitFailedCount // int // for failed attempts in locked short
	ShortAccessVisitDeleteCount // int // how many delete occurs

	// ==============================================================================================
	// metrics per served path (only the app specific pages / urls ; not the framework / extra stuff)
	// ----------------------------------------------------------------------------------------------

	// per served path
	ServedPathName     // string
	ServedPathTime     // string
	ServedPathIP       // string
	ServedPathInfo     // string // useragent / other
	ServedPathReferrer // string

	// per failed serve path
	FailedServedPathName     // string
	FailedServedPathTime     // string
	FailedServedPathIP       // string
	FailedServedPathInfo     // string // useragent / other
	FailedServedPathReferrer // string
	FailedServedPathReason   // string // if applicable ..??

	// global served path
	ServedPathCount       // int
	ServedPathFailedCount // int
)

func (mo MetricType) String() string {
	switch mo {
	case InvalidMetrtic:
		return "InvalidMetrtic"

	case FailedShortCreateShort: // string // short name
		return "FailedShortCreateShort"
	case FailedShortCreateTime: // string
		return "FailedShortCreateTime"
	case FailedShortCreateIP: // string
		return "FailedShortCreateIP"
	case FailedShortCreateInfo: // string // useragent / other ...
		return "FailedShortCreateInfo"
	case FailedShortCreateReferrer: // string
		return "FailedShortCreateReferrer"
	case FailedShortCreateReason: // string // if applicable ..?!?
		return "FailedShortCreateReason"

	case FailedShortCreateCounter: //int
		return "FailedShortCreateCounter"

	case InvalidShortAccessShort: // string // short name
		return "InvalidShortAccessShort"
	case InvalidShortAccessTime: // string
		return "InvalidShortAccessTime"
	case InvalidShortAccessIP: // string
		return "InvalidShortAccessIP"
	case InvalidShortAccessInfo: // string // useragent / other ...
		return "InvalidShortAccessInfo"
	case InvalidShortAccessReferrer: // string
		return "InvalidShortAccessReferrer"

	case InvalidShortAccessCounter: // int
		return "InvalidShortAccessCounter"

	case ShortCreateIP: // string // ip
		return "ShortCreateIP"
	case ShortCreateName: // short name
		return "ShortCreateName"
	case ShortCreateTime: // short name
		return "ShortCreateTime"
	case ShortCreateInfo: // string // useragaent / user id / other ...
		return "ShortCreateInfo"
	case ShortCreatePrivate: // bool
		return "ShortCreatePrivate"
	case ShortCreateDelete: // bool
		return "ShortCreateDelete"
	case ShortCreatedNamed: // bool
		return "ShortCreatedNamed"
	case ShortCreateReferrer: // string
		return "ShortCreateReferrer"

	case ShortCreatedCount: // int
		return "ShortCreatedCount"
	case ShortExpiredCount: // int
		return "ShortExpiredCount"

	case ShortAccessVisitTime: // string
		return "ShortAccessVisitTime"
	case ShortAccessVisitIP: // string
		return "ShortAccessVisitIP"
	case ShortAccessVisitInfo: // string // useragent / other ...
		return "ShortAccessVisitInfo"
	case ShortAccessVisitReferrer: // string
		return "ShortAccessVisitReferrer"
	case ShortAccessVisitSuccess: // bool
		return "ShortAccessVisitSuccess"
	case ShortAccessVisitPrivate: // bool
		return "ShortAccessVisitPrivate"
	case ShortAccessVisitDelete: // bool
		return "ShortAccessVisitDelete"
	case ShortAccessVisitIsLocked: // bool
		return "ShortAccessVisitIsLocked"

	case ShortAccessVisitCount: // int
		return "ShortAccessVisitCount"
	case ShortAccessVisitFailedCount: // int // for failed attempts in locked short
		return "ShortAccessVisitFailedCount"
	case ShortAccessVisitDeleteCount: // int // how many delete occurs
		return "ShortAccessVisitDeleteCount"

	case ServedPathName: // string
		return "ServedPathName"
	case ServedPathTime: // string
		return "ServedPathTime"
	case ServedPathIP: // string
		return "ServedPathIP"
	case ServedPathInfo: // string // useragent / other
		return "ServedPathInfo"
	case ServedPathReferrer: // string
		return "ServedPathReferrer"

	case FailedServedPathName: // string
		return "FailedServedPathName"
	case FailedServedPathTime: // string
		return "FailedServedPathTime"
	case FailedServedPathIP: // string
		return "FailedServedPathIP"
	case FailedServedPathInfo: // string // useragent / other
		return "FailedServedPathInfo"
	case FailedServedPathReferrer: // string
		return "FailedServedPathReferrer"
	case FailedServedPathReason: // string // if applicable ..??
		return "FailedServedPathReason"

	case ServedPathCount: // int
		return "ServedPathCount"
	case ServedPathFailedCount: // int
		return "ServedPathFailedCount"
	}
	return "Unknown"
}

type MetricPacker interface {
	Encode() MetricPacker
	EncodedString() string
	Decode() MetricPacker
	ToMap() MetricPacker
	ToObject() MetricPacker
	FromString(string) MetricPacker
	Error() error
	DumpMap() string
	DumpObject() string
	Equal(MetricPacker) bool
	Name() string
}

type MetricGroupType int

const (
	MetricGroupGlobal MetricGroupType = iota
	MetricGroupShortCreationFailure
	MetricGroupShortCreationSuccess
	MetricGroupShortAccessInvalid
	MetricGroupShortAccessSuccess
)

func NewMetricGroup(mg MetricGroupType) MetricPacker {
	switch mg {
	case MetricGroupGlobal:
		return NewMetricGlobal()
	case MetricGroupShortCreationFailure:
		return NewMetricShortCreationFailure()
	case MetricGroupShortCreationSuccess:
		return NewMetricShortCreationSuccess()
	case MetricGroupShortAccessInvalid:
		return NewMetricShortAccessInvalid()
	case MetricGroupShortAccessSuccess:
		return NewMetricShortAccess()
	}
	return nil
}

// func (m *metricObject) DumpObject() string {
// 	if _, ok := m.MetricPacker.(MetricGlobal); ok {

// 	}
// }

type metricObject struct {
	MetricPacker
	encoded string
	mapped  map[interface{}]interface{}
	err     error
	name    string
}

func newMetricObject(name string) *metricObject {
	return &metricObject{
		mapped: map[interface{}]interface{}{},
		name:   name,
	}
}

func (m *metricObject) FromString(in string) MetricPacker {
	m.encoded = in
	return m
}

func (m *metricObject) Name() string {
	return m.name
}

func (m *metricObject) EncodedString() string {
	return m.encoded
}

func (m *metricObject) DumpMap() string {
	return fmt.Sprintf("%s: %#+v", m.name, m.mapped)
}
func (m *metricObject) Error() error {
	return m.err
}

func (m *metricObject) Encode() MetricPacker {
	if m.err != nil {
		return m
	}
	if m.mapped == nil {
		m.err = errors.New("map not ready")
		return m
	}
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	m.err = enc.Encode(m.mapped)
	if m.err != nil {
		return m
	}
	m.encoded = buf.String()
	return m
}

func (m *metricObject) Decode() MetricPacker {
	if m.err != nil {
		return m
	}
	if m.mapped == nil {
		m.err = errors.New("nowhere to decode")
		return m
	}
	if m.encoded == "" {
		m.err = errors.New("nothing to decode")
		return m
	}
	if len(m.mapped) > 0 {
		m.mapped = map[interface{}]interface{}{}
	}
	buf := bytes.NewBufferString(m.encoded)
	dec := msgpack.NewDecoder(buf)
	dec.SetMapDecoder(func(d *msgpack.Decoder) (interface{}, error) {
		return d.DecodeTypedMap()
	})
	m.err = dec.Decode(&m.mapped)
	return m
}

type MetricGlobal struct {
	*metricObject

	// global failed creation
	FailedShortCreateCounter int

	// global invalid access
	InvalidShortAccessCounter int

	// global created / expired shorts
	ShortCreatedCount int
	ShortExpiredCount int

	// global short Access counter - for all visits
	ShortAccessVisitCount       int
	ShortAccessVisitFailedCount int // for failed attempts in locked short
	ShortAccessVisitDeleteCount int // how many delete occurs

	// global served path
	ServedPathCount       int
	ServedPathFailedCount int
}

func NewMetricGlobal() *MetricGlobal {
	return &MetricGlobal{
		metricObject: newMetricObject("global"),
	}
}

func (m *MetricGlobal) IncFailedShortCreateCounter() *MetricGlobal {
	m.FailedShortCreateCounter += 1
	return m
}
func (m *MetricGlobal) IncInvalidShortAccessCounter() *MetricGlobal {
	m.InvalidShortAccessCounter += 1
	return m
}
func (m *MetricGlobal) IncShortCreatedCount() *MetricGlobal {
	m.ShortCreatedCount += 1
	return m
}
func (m *MetricGlobal) IncShortExpiredCount() *MetricGlobal {
	m.ShortExpiredCount += 1
	return m
}
func (m *MetricGlobal) IncShortAccessVisitCount() *MetricGlobal {
	m.ShortAccessVisitCount += 1
	return m
}
func (m *MetricGlobal) IncShortAccessVisitFailedCount() *MetricGlobal {
	m.ShortAccessVisitFailedCount += 1
	return m
}
func (m *MetricGlobal) IncShortAccessVisitDeleteCount() *MetricGlobal {
	m.ShortAccessVisitDeleteCount += 1
	return m
}
func (m *MetricGlobal) IncServedPathCount() *MetricGlobal { m.ServedPathCount += 1; return m }
func (m *MetricGlobal) IncServedPathFailedCount() *MetricGlobal {
	m.ServedPathFailedCount += 1
	return m
}

func (m *MetricGlobal) ToMap() MetricPacker {
	m.mapped = map[interface{}]interface{}{
		FailedShortCreateCounter:    m.FailedShortCreateCounter,
		InvalidShortAccessCounter:   m.InvalidShortAccessCounter,
		ShortCreatedCount:           m.ShortCreatedCount,
		ShortExpiredCount:           m.ShortExpiredCount,
		ShortAccessVisitCount:       m.ShortAccessVisitCount,
		ShortAccessVisitFailedCount: m.ShortAccessVisitFailedCount,
		ShortAccessVisitDeleteCount: m.ShortAccessVisitDeleteCount,
		ServedPathCount:             m.ServedPathCount,
		ServedPathFailedCount:       m.ServedPathFailedCount,
	}
	return m
}
func (m *MetricGlobal) ToObject() MetricPacker {
	if m.err != nil {
		return m
	}
	defer func() {
		if a := recover(); a != nil {
			m.err = errors.Errorf("panic occurred: <%v>", a)
		}
	}()
	mapped := map[MetricType]int{}
	for k, v := range m.mapped {
		mapped[MetricType(k.(int8))] = int(v.(int8))
	}
	m.FailedShortCreateCounter = mapped[FailedShortCreateCounter]
	m.InvalidShortAccessCounter = mapped[InvalidShortAccessCounter]
	m.ShortCreatedCount = mapped[ShortCreatedCount]
	m.ShortExpiredCount = mapped[ShortExpiredCount]
	m.ShortAccessVisitCount = mapped[ShortAccessVisitCount]
	m.ShortAccessVisitFailedCount = mapped[ShortAccessVisitFailedCount]
	m.ShortAccessVisitDeleteCount = mapped[ShortAccessVisitDeleteCount]
	m.ServedPathCount = mapped[ServedPathCount]
	m.ServedPathFailedCount = mapped[ServedPathFailedCount]

	return m
}

func (m *MetricGlobal) DumpObject() string {
	return fmt.Sprintf(
		"%s:\n\t"+
			"%s: %d\n\t"+
			"%s: %d\n\t"+
			"%s: %d\n\t"+
			"%s: %d\n\t"+
			"%s: %d\n\t"+
			"%s: %d\n\t"+
			"%s: %d\n\t"+
			"%s: %d\n\t"+
			"%s: %d\n\t",
		m.name,
		FailedShortCreateCounter, m.FailedShortCreateCounter, // = mapped[FailedShortCreateCounter]
		InvalidShortAccessCounter, m.InvalidShortAccessCounter, // = mapped[InvalidShortAccessCounter]
		ShortCreatedCount, m.ShortCreatedCount, // = mapped[ShortCreatedCount]
		ShortExpiredCount, m.ShortExpiredCount, // = mapped[ShortExpiredCount]
		ShortAccessVisitCount, m.ShortAccessVisitCount, // = mapped[ShortAccessVisitCount]
		ShortAccessVisitFailedCount, m.ShortAccessVisitFailedCount, // = mapped[ShortAccessVisitFailedCount]
		ShortAccessVisitDeleteCount, m.ShortAccessVisitDeleteCount, // = mapped[ShortAccessVisitDeleteCount]
		ServedPathCount, m.ServedPathCount, // = mapped[ServedPathCount]
		ServedPathFailedCount, m.ServedPathFailedCount, // = mapped[ServedPathFailedCount]
	)
}

func (m *MetricGlobal) Equal(om MetricPacker) bool {
	m2, ok := om.(*MetricGlobal)
	return ok &&
		m2.FailedShortCreateCounter == m.FailedShortCreateCounter &&
		m2.InvalidShortAccessCounter == m.InvalidShortAccessCounter &&
		m2.ShortCreatedCount == m.ShortCreatedCount &&
		m2.ShortExpiredCount == m.ShortExpiredCount &&
		m2.ShortAccessVisitCount == m.ShortAccessVisitCount &&
		m2.ShortAccessVisitFailedCount == m.ShortAccessVisitFailedCount &&
		m2.ShortAccessVisitDeleteCount == m.ShortAccessVisitDeleteCount &&
		m2.ServedPathCount == m.ServedPathCount &&
		m2.ServedPathFailedCount == m.ServedPathFailedCount
}

// # Creation Failure Metric
// #######################################

type MetricShortCreationFailure struct {
	*metricObject
	// failed short creation - per failure
	FailedShortCreateShort    string // short name
	FailedShortCreateTime     string
	FailedShortCreateIP       string
	FailedShortCreateInfo     string // useragent / other ...
	FailedShortCreateReferrer string
	FailedShortCreateReason   string // if applicable ..?!?
}

func NewMetricShortCreationFailure() *MetricShortCreationFailure {
	return &MetricShortCreationFailure{
		metricObject: newMetricObject("short creation failure"),
	}
}

func (m *MetricShortCreationFailure) ToMap() MetricPacker {
	m.mapped = map[interface{}]interface{}{
		FailedShortCreateShort:    m.FailedShortCreateShort,
		FailedShortCreateTime:     m.FailedShortCreateTime,
		FailedShortCreateIP:       m.FailedShortCreateIP,
		FailedShortCreateInfo:     m.FailedShortCreateInfo,
		FailedShortCreateReferrer: m.FailedShortCreateReferrer,
		FailedShortCreateReason:   m.FailedShortCreateReason,
	}
	return m
}
func (m *MetricShortCreationFailure) ToObject() MetricPacker {
	defer func() {
		if a := recover(); a != nil {
			m.err = errors.Errorf("panic occurred: <%v>", a)
		}
	}()
	mapped := map[MetricType]string{}
	for k, v := range m.mapped {
		mapped[MetricType(k.(int8))] = v.(string)
	}
	m.FailedShortCreateShort = mapped[FailedShortCreateShort]
	m.FailedShortCreateTime = mapped[FailedShortCreateTime]
	m.FailedShortCreateIP = mapped[FailedShortCreateIP]
	m.FailedShortCreateInfo = mapped[FailedShortCreateInfo]
	m.FailedShortCreateReferrer = mapped[FailedShortCreateReferrer]
	m.FailedShortCreateReason = mapped[FailedShortCreateReason]

	return m
}

func (m *MetricShortCreationFailure) DumpObject() string {
	return fmt.Sprintf(
		"%s:\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t",
		m.name,
		FailedShortCreateShort, m.FailedShortCreateShort,
		FailedShortCreateTime, m.FailedShortCreateTime,
		FailedShortCreateIP, m.FailedShortCreateIP,
		FailedShortCreateInfo, m.FailedShortCreateInfo,
		FailedShortCreateReferrer, m.FailedShortCreateReferrer,
		FailedShortCreateReason, m.FailedShortCreateReason,
	)
}
func (m *MetricShortCreationFailure) Equal(om MetricPacker) bool {
	m2, ok := om.(*MetricShortCreationFailure)
	return ok &&
		m2.FailedShortCreateShort == m.FailedShortCreateShort &&
		m2.FailedShortCreateTime == m.FailedShortCreateTime &&
		m2.FailedShortCreateIP == m.FailedShortCreateIP &&
		m2.FailedShortCreateInfo == m.FailedShortCreateInfo &&
		m2.FailedShortCreateReferrer == m.FailedShortCreateReferrer &&
		m2.FailedShortCreateReason == m.FailedShortCreateReason
}

// # Access Invalid Metric
// #######################################

type MetricShortAccessInvalid struct {
	*metricObject
	// invalid short access (non existant)
	InvalidShortAccessShort    string // short name
	InvalidShortAccessTime     string
	InvalidShortAccessIP       string
	InvalidShortAccessInfo     string // useragent / other ...
	InvalidShortAccessReferrer string
}

func NewMetricShortAccessInvalid() *MetricShortAccessInvalid {
	return &MetricShortAccessInvalid{
		metricObject: newMetricObject("short access invalid"),
	}
}

func (m *MetricShortAccessInvalid) ToMap() MetricPacker {
	m.mapped = map[interface{}]interface{}{
		InvalidShortAccessShort:    m.InvalidShortAccessShort,    //    string // short name
		InvalidShortAccessTime:     m.InvalidShortAccessTime,     //     string
		InvalidShortAccessIP:       m.InvalidShortAccessIP,       //       string
		InvalidShortAccessInfo:     m.InvalidShortAccessInfo,     //     string // useragent / other ...
		InvalidShortAccessReferrer: m.InvalidShortAccessReferrer, // string
	}
	return m
}
func (m *MetricShortAccessInvalid) ToObject() MetricPacker {
	defer func() {
		if a := recover(); a != nil {
			m.err = errors.Errorf("panic occurred: <%v>", a)
		}
	}()
	mapped := map[MetricType]string{}
	for k, v := range m.mapped {
		mapped[MetricType(k.(int8))] = v.(string)
	}
	m.InvalidShortAccessShort = mapped[InvalidShortAccessShort]       //    string // short name
	m.InvalidShortAccessTime = mapped[InvalidShortAccessTime]         //     string
	m.InvalidShortAccessIP = mapped[InvalidShortAccessIP]             //       string
	m.InvalidShortAccessInfo = mapped[InvalidShortAccessInfo]         //     string // useragent / other ...
	m.InvalidShortAccessReferrer = mapped[InvalidShortAccessReferrer] // string
	return m
}

func (m *MetricShortAccessInvalid) DumpObject() string {
	return fmt.Sprintf(
		"%s:\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t",
		m.name,
		InvalidShortAccessShort, m.InvalidShortAccessShort, //    string // short name
		InvalidShortAccessTime, m.InvalidShortAccessTime, //     string
		InvalidShortAccessIP, m.InvalidShortAccessIP, //       string
		InvalidShortAccessInfo, m.InvalidShortAccessInfo, //     string // useragent / other ...
		InvalidShortAccessReferrer, m.InvalidShortAccessReferrer, // string
	)
}
func (m *MetricShortAccessInvalid) Equal(om MetricPacker) bool {
	m2, ok := om.(*MetricShortAccessInvalid)
	return ok &&
		m2.InvalidShortAccessShort == m.InvalidShortAccessShort &&
		m2.InvalidShortAccessTime == m.InvalidShortAccessTime &&
		m2.InvalidShortAccessIP == m.InvalidShortAccessIP &&
		m2.InvalidShortAccessInfo == m.InvalidShortAccessInfo &&
		m2.InvalidShortAccessReferrer == m.InvalidShortAccessReferrer
}

// # Creation Success Metric
// #######################################

type MetricShortCreationSuccess struct {
	*metricObject
	// Short Creation - only once on creation
	ShortCreateName     string // short name
	ShortCreateTime     string // short name
	ShortCreateIP       string // ip
	ShortCreateInfo     string // useragaent / user id / other ...
	ShortCreatePrivate  string // bool
	ShortCreateDelete   string // bool
	ShortCreatedNamed   string // bool
	ShortCreateReferrer string
}

func NewMetricShortCreationSuccess() *MetricShortCreationSuccess {
	return &MetricShortCreationSuccess{
		metricObject: newMetricObject("short creation success"),
	}
}

func (m *MetricShortCreationSuccess) ToMap() MetricPacker {
	m.mapped = map[interface{}]interface{}{
		ShortCreateName:     m.ShortCreateName,     //       short name
		ShortCreateTime:     m.ShortCreateTime,     //       short name
		ShortCreateIP:       m.ShortCreateIP,       //       string // ip
		ShortCreateInfo:     m.ShortCreateInfo,     //     string // useragaent / user id / other ...
		ShortCreatePrivate:  m.ShortCreatePrivate,  //  string // bool
		ShortCreateDelete:   m.ShortCreateDelete,   //   string // bool
		ShortCreatedNamed:   m.ShortCreatedNamed,   //   string // bool
		ShortCreateReferrer: m.ShortCreateReferrer, // string
	}
	return m
}
func (m *MetricShortCreationSuccess) ToObject() MetricPacker {
	defer func() {
		if a := recover(); a != nil {
			m.err = errors.Errorf("panic occurred: <%v>", a)
		}
	}()
	mapped := map[MetricType]string{}
	for k, v := range m.mapped {
		mapped[MetricType(k.(int8))] = v.(string)
	}
	m.ShortCreateName = mapped[ShortCreateName]         //       short name
	m.ShortCreateTime = mapped[ShortCreateTime]         //       short name
	m.ShortCreateIP = mapped[ShortCreateIP]             //       string // ip
	m.ShortCreateInfo = mapped[ShortCreateInfo]         //     string // useragaent / user id / other ...
	m.ShortCreatePrivate = mapped[ShortCreatePrivate]   //  string // bool
	m.ShortCreateDelete = mapped[ShortCreateDelete]     //   string // bool
	m.ShortCreatedNamed = mapped[ShortCreatedNamed]     //   string // bool
	m.ShortCreateReferrer = mapped[ShortCreateReferrer] // string
	return m
}

func (m *MetricShortCreationSuccess) DumpObject() string {
	return fmt.Sprintf(
		"%s:\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t",
		m.name,

		ShortCreateName, m.ShortCreateName, //       string // ip
		ShortCreateTime, m.ShortCreateTime, //       string // ip
		ShortCreateIP, m.ShortCreateIP, //       string // ip
		ShortCreateInfo, m.ShortCreateInfo, //     string // useragaent / user id / other ...
		ShortCreatePrivate, m.ShortCreatePrivate, //  string // bool
		ShortCreateDelete, m.ShortCreateDelete, //   string // bool
		ShortCreatedNamed, m.ShortCreatedNamed, //   string // bool
		ShortCreateReferrer, m.ShortCreateReferrer, // string
	)
}
func (m *MetricShortCreationSuccess) Equal(om MetricPacker) bool {
	m2, ok := om.(*MetricShortCreationSuccess)
	return ok &&
		m2.ShortCreateName == m.ShortCreateName &&
		m2.ShortCreateTime == m.ShortCreateTime &&
		m2.ShortCreateIP == m.ShortCreateIP &&
		m2.ShortCreateInfo == m.ShortCreateInfo &&
		m2.ShortCreatePrivate == m.ShortCreatePrivate &&
		m2.ShortCreateDelete == m.ShortCreateDelete &&
		m2.ShortCreatedNamed == m.ShortCreatedNamed &&
		m2.ShortCreateReferrer == m.ShortCreateReferrer
}

// # Access Success Metric
// #######################################

type MetricShortAccess struct {
	*metricObject

	// Short Access - per visit metrics
	ShortAccessVisitTime     string
	ShortAccessVisitIP       string
	ShortAccessVisitInfo     string // useragent / other ...
	ShortAccessVisitReferrer string
	ShortAccessVisitSuccess  string //bool
	ShortAccessVisitPrivate  string //bool
	ShortAccessVisitDelete   string //bool
	ShortAccessVisitIsLocked string //bool
}

func NewMetricShortAccess() *MetricShortAccess {
	return &MetricShortAccess{
		metricObject: newMetricObject("short access"),
	}
}

func (m *MetricShortAccess) ToMap() MetricPacker {
	m.mapped = map[interface{}]interface{}{
		ShortAccessVisitTime:     m.ShortAccessVisitTime,
		ShortAccessVisitIP:       m.ShortAccessVisitIP,
		ShortAccessVisitInfo:     m.ShortAccessVisitInfo,
		ShortAccessVisitReferrer: m.ShortAccessVisitReferrer,
		ShortAccessVisitSuccess:  m.ShortAccessVisitSuccess,
		ShortAccessVisitPrivate:  m.ShortAccessVisitPrivate,
		ShortAccessVisitDelete:   m.ShortAccessVisitDelete,
		ShortAccessVisitIsLocked: m.ShortAccessVisitIsLocked,
	}
	return m
}
func (m *MetricShortAccess) ToObject() MetricPacker {
	defer func() {
		if a := recover(); a != nil {
			m.err = errors.Errorf("panic occurred: <%v>", a)
		}
	}()
	mapped := map[MetricType]string{}
	for k, v := range m.mapped {
		mapped[MetricType(k.(int8))] = v.(string)
	}
	m.ShortAccessVisitTime = mapped[ShortAccessVisitTime]
	m.ShortAccessVisitIP = mapped[ShortAccessVisitIP]
	m.ShortAccessVisitInfo = mapped[ShortAccessVisitInfo]
	m.ShortAccessVisitReferrer = mapped[ShortAccessVisitReferrer]
	m.ShortAccessVisitSuccess = mapped[ShortAccessVisitSuccess]
	m.ShortAccessVisitPrivate = mapped[ShortAccessVisitPrivate]
	m.ShortAccessVisitDelete = mapped[ShortAccessVisitDelete]
	m.ShortAccessVisitIsLocked = mapped[ShortAccessVisitIsLocked]
	return m
}

func (m *MetricShortAccess) DumpObject() string {
	return fmt.Sprintf(
		"%s:\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t"+
			"%s: %s\n\t",
		m.name,
		ShortAccessVisitTime, m.ShortAccessVisitTime, //     string
		ShortAccessVisitIP, m.ShortAccessVisitIP, //       string
		ShortAccessVisitInfo, m.ShortAccessVisitInfo, //     string // useragent / other ...
		ShortAccessVisitReferrer, m.ShortAccessVisitReferrer, // string
		ShortAccessVisitSuccess, m.ShortAccessVisitSuccess, //  string //bool
		ShortAccessVisitPrivate, m.ShortAccessVisitPrivate, //  string //bool
		ShortAccessVisitDelete, m.ShortAccessVisitDelete, //   string //bool
		ShortAccessVisitIsLocked, m.ShortAccessVisitIsLocked, // string //bool
	)
}

func (m *MetricShortAccess) Equal(om MetricPacker) bool {
	m2, ok := om.(*MetricShortAccess)
	return ok &&
		m2.ShortAccessVisitTime == m.ShortAccessVisitTime &&
		m2.ShortAccessVisitIP == m.ShortAccessVisitIP &&
		m2.ShortAccessVisitInfo == m.ShortAccessVisitInfo &&
		m2.ShortAccessVisitReferrer == m.ShortAccessVisitReferrer &&
		m2.ShortAccessVisitSuccess == m.ShortAccessVisitSuccess &&
		m2.ShortAccessVisitPrivate == m.ShortAccessVisitPrivate &&
		m2.ShortAccessVisitDelete == m.ShortAccessVisitDelete &&
		m2.ShortAccessVisitIsLocked == m.ShortAccessVisitIsLocked
}
