package metrics

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

type metricPacker interface {
	Pack() map[interface{}]interface{}
}

type MetricGlobal struct {
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

func (m *MetricGlobal) Pack() map[interface{}]interface{} {
	return map[interface{}]interface{}{
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
}

type MetricShortCreationFailure struct {
	// failed short creation - per failure
	FailedShortCreateShort    string // short name
	FailedShortCreateTime     string
	FailedShortCreateIP       string
	FailedShortCreateInfo     string // useragent / other ...
	FailedShortCreateReferrer string
	FailedShortCreateReason   string // if applicable ..?!?
}

func (m *MetricShortCreationFailure) Pack() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		FailedShortCreateShort:    m.FailedShortCreateShort,
		FailedShortCreateTime:     m.FailedShortCreateTime,
		FailedShortCreateIP:       m.FailedShortCreateIP,
		FailedShortCreateInfo:     m.FailedShortCreateInfo,
		FailedShortCreateReferrer: m.FailedShortCreateReferrer,
		FailedShortCreateReason:   m.FailedShortCreateReason,
	}
}

type MetricShortAccessInvalid struct {
	// invalid short access (non existant)
	InvalidShortAccessShort    string // short name
	InvalidShortAccessTime     string
	InvalidShortAccessIP       string
	InvalidShortAccessInfo     string // useragent / other ...
	InvalidShortAccessReferrer string
}

func (m *MetricShortAccessInvalid) Pack() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		InvalidShortAccessShort:    m.InvalidShortAccessShort,
		InvalidShortAccessTime:     m.InvalidShortAccessTime,
		InvalidShortAccessIP:       m.InvalidShortAccessIP,
		InvalidShortAccessInfo:     m.InvalidShortAccessInfo,
		InvalidShortAccessReferrer: m.InvalidShortAccessReferrer,
	}
}

type MetricShortCreationSuccess struct {
	// Short Creation - only once on creation
	ShortCreateIP       string // ip
	ShortCreateInfo     string // useragaent / user id / other ...
	ShortCreatePrivate  bool
	ShortCreateDelete   bool
	ShortCreatedNamed   bool
	ShortCreateReferrer string
}

func (m *MetricShortCreationSuccess) Pack() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		ShortCreateIP:       m.ShortCreateIP,
		ShortCreateInfo:     m.ShortCreateInfo,
		ShortCreatePrivate:  m.ShortCreatePrivate,
		ShortCreateDelete:   m.ShortCreateDelete,
		ShortCreatedNamed:   m.ShortCreatedNamed,
		ShortCreateReferrer: m.ShortCreateReferrer,
	}
}

type MetricShortAccess struct {
	// Short Access - per visit metrics
	ShortAccessVisitTime     string
	ShortAccessVisitIP       string
	ShortAccessVisitInfo     string // useragent / other ...
	ShortAccessVisitReferrer string
	ShortAccessVisitSuccess  bool
	ShortAccessVisitPrivate  bool
	ShortAccessVisitDelete   bool
	ShortAccessVisitIsLocked bool
}

func (m *MetricShortAccess) Pack() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		ShortAccessVisitTime:     m.ShortAccessVisitTime,
		ShortAccessVisitIP:       m.ShortAccessVisitIP,
		ShortAccessVisitInfo:     m.ShortAccessVisitInfo,
		ShortAccessVisitReferrer: m.ShortAccessVisitReferrer,
		ShortAccessVisitSuccess:  m.ShortAccessVisitSuccess,
		ShortAccessVisitPrivate:  m.ShortAccessVisitPrivate,
		ShortAccessVisitDelete:   m.ShortAccessVisitDelete,
		ShortAccessVisitIsLocked: m.ShortAccessVisitIsLocked,
	}
}
