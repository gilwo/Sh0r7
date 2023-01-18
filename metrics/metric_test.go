package metrics

import (
	"log"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestMetrics1(t *testing.T) {
	g1 := NewMetricGlobal().IncFailedShortCreateCounter().IncInvalidShortAccessCounter().IncServedPathCount().
		IncShortAccessVisitDeleteCount()
	g1.IncShortAccessVisitCount().IncShortAccessVisitCount()
	if err := g1.ToMap().Encode().Error(); err != nil {
		panic(err)
	}
	// log.Println("dump: " + g1.DumpMap())
	// log.Println("string: " + g1.EncodedString())
	// log.Printf("g1: %+#v\n", g1)

	g2 := NewMetricGlobal()
	g2.FromString(g1.EncodedString())
	g2.Decode()
	// log.Printf("decode: %v\n", g2.Error())
	// log.Println("dump: " + g2.DumpMap())
	g2.ToObject()
	if e := g2.Error(); e != nil {
		log.Printf("!!!!! error in conversion: %v\n", e)
	}
	// log.Println("dump: " + g2.DumpMap())
	// log.Println("string: " + g2.EncodedString())
	// log.Printf("g1: %+#v\n", g2)
	if !g1.Equal(g2) {
		t.Logf("g1:\n%s\n", g1.DumpObject())
		t.Logf("g2:\n%s\n", g2.DumpObject())
		t.Logf("g1 == g2 : %v\n", g1.Equal(g2))
		t.FailNow()
	}

	// source := "abcdefghijklmnopqrstuvwxyz0123456789"
	// var ofs, fixed, min, max int
	// var res string

	// ofs, fixed, min, max = 0, 10, -1, -1
	// res = GenericShort(source, ofs, fixed, min, max, nil)
	// fmt.Printf("res: <%s>(%d)\n(ofs, fixed, min, max):(%d, %d, %d, %d) => (<source>,<result>):(<%s>:<%s>(%d))\n",
	// 	res, len(res), ofs, fixed, min, max, source, res, len(res))

	// for _, e := range params() {
	// 	ofs, fixed, min, max = e[0], e[1], e[2], e[3]
	// 	res := GenericShort(source, ofs, fixed, min, max, nil)
	// 	fmt.Printf("res: <%s>(%d)\n(ofs, fixed, min, max):(%d, %d, %d, %d) => (<source>,<result>):(<%s>:<%s>(%d))\n",
	// 		res, len(res), ofs, fixed, min, max, source, res, len(res))
	// }
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
		panic(err)
	}
	// log.Println("dump: " + g1.DumpMap())
	// log.Println("string: " + g1.EncodedString())
	// log.Printf("g1: %+#v\n", g1)

	g2 := NewMetricShortCreationFailure()
	g2.FromString(g1.EncodedString())
	if err := g2.Decode().Error(); err != nil {
		log.Printf("!!!!! error in decoding: %v\n", err)
	}
	log.Printf("decode dump: %v\n", g2.DumpMap())
	if err := g2.ToObject().Error(); err != nil {
		log.Printf("!!!!! error in conversion: %v\n", err)
	}
	// log.Println("toobject dump: " + g2.DumpMap())

	// log.Println("dump: " + g2.DumpMap())
	// log.Println("string: " + g2.EncodedString())
	// log.Printf("g2: %+#v\n", g2)
	log.Printf("g2: %s\n", g2.DumpObject())

	// source := "abcdefghijklmnopqrstuvwxyz0123456789"
	// var ofs, fixed, min, max int
	// var res string

	// ofs, fixed, min, max = 0, 10, -1, -1
	// res = GenericShort(source, ofs, fixed, min, max, nil)
	// fmt.Printf("res: <%s>(%d)\n(ofs, fixed, min, max):(%d, %d, %d, %d) => (<source>,<result>):(<%s>:<%s>(%d))\n",
	// 	res, len(res), ofs, fixed, min, max, source, res, len(res))

	// for _, e := range params() {
	// 	ofs, fixed, min, max = e[0], e[1], e[2], e[3]
	// 	res := GenericShort(source, ofs, fixed, min, max, nil)
	// 	fmt.Printf("res: <%s>(%d)\n(ofs, fixed, min, max):(%d, %d, %d, %d) => (<source>,<result>):(<%s>:<%s>(%d))\n",
	// 		res, len(res), ofs, fixed, min, max, source, res, len(res))
	// }
}

func params() [][]int {
	r := [][]int{
		{0, 10, -1, -1},
		{1, 11, -1, -1},
		{1, 11, -1, -1},
		{2, 10, 5, 7},
		{2, 0, 5, 8},
		{2, 0, 5, 8},
		{2, 0, 5, 8},
		{2, 0, 5, 8},
		{2, 0, 5, 8},
		{2, 0, 5, 8},
		{3, 0, 5, 0},
		{3, 0, 5, 0},
		{3, 0, 5, 0},
		{3, 0, 5, 0},
		{3, 0, 5, 0},
		{3, 0, 5, 0},
		{3, 0, 5, 0},
		{3, 0, 3, 5},
		{3, 0, 3, 5},
		{3, 0, 3, 5},
		{3, 0, 3, 5},
		{3, 0, 3, 5},
		{3, 0, 3, 5},
	}
	return r
}
