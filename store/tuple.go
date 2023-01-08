package store

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	KeyNotFound = errors.New("key not found")
)

type fieldValue struct {
	f, v string
}
type stringTuple struct {
	tuple map[string]string
}

func NewTupleFromString(in string) (*stringTuple, error) {
	t := NewTuple()
	if err := t.FromString(in); err != nil {
		return nil, err
	}
	return t, nil
}

func NewTuple() *stringTuple {
	return &stringTuple{map[string]string{}}
}
func initStringTuple(size int, names ...string) (*stringTuple, error) {
	if len(names) != size {
		return nil, errors.New("mismatch in size and field names count")
	}
	r := &stringTuple{
		tuple: map[string]string{},
	}
	for _, e := range names {
		if _, ok := r.tuple[e]; ok {
			return nil, errors.Errorf("field names have duplicates: %v", names)
		}
		r.tuple[e] = ""
	}
	return r, nil
}
func NewStringTuple(values ...*fieldValue) (*stringTuple, error) {
	f := []string{}
	for _, e := range values {
		f = append(f, e.f)
	}
	r, err := initStringTuple(len(f), f...)
	if err != nil {
		return nil, err
	}
	for _, e := range values {
		err := r.SetCheck(e.f, e.v)
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}
func (t *stringTuple) AtCheck(field string) (string, error) {
	// fmt.Printf("content of tuple %#v\n", t)
	if v, ok := t.tuple[field]; ok {
		return v, nil
	}
	return "", errors.Wrapf(KeyNotFound, " "+field)
}
func (t *stringTuple) Get(field string) string {
	v, err := t.AtCheck(field)
	if err != nil {
		v = ""
	}
	return v
}
func (t *stringTuple) Set(field, value string) {
	t.tuple[field] = value
}
func (t *stringTuple) SetCheck(field, value string) error {
	if _, ok := t.tuple[field]; !ok {
		return errors.Wrapf(KeyNotFound, " "+field)
	}
	t.tuple[field] = value
	return nil
}
func (t *stringTuple) Keys() []string {
	r := []string{}
	for k := range t.tuple {
		if strings.HasSuffix(k, "url") ||
			strings.HasSuffix(k, fCompress) {
			continue
		}
		r = append(r, k)
	}
	sort.Strings(r)
	return r
}

func (t *stringTuple) packStringMsgPack() (string, error) {
	if r, err := t.packMsgPack(); err != nil {
		return "", err
	} else {
		return string(r), nil
	}
}

func (t *stringTuple) packMsgPack() ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.SetSortMapKeys(true)
	err := enc.Encode(t.tuple)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (t *stringTuple) unpackStringMsgPack(data string) error {
	return t.unpackMsgPack([]byte(data))
}
func (t *stringTuple) unpackMsgPack(data []byte) error {
	err := msgpack.Unmarshal(data, &t.tuple)

	if err != nil {
		return err
	}
	return nil
}

func (t *stringTuple) Dump() string {
	s := ""
	for _, k := range t.Keys() {
		v := t.Get(k)
		if len(v) > 300 {
			v = v[:300]
		}
		s += "\t[" + k + "]:\n\t\t<" + v + ">\n"
	}
	return s
}

const (
	fCompress        = ".compress"
	FieldTime        = "created"
	FieldTTL         = "ttl"
	FieldDATA        = "data"
	FieldModCount    = "changes"
	FieldModTime     = "changed_time"
	FieldBlocked     = "blocked"
	FieldDesc        = "description"
	FieldPrvPassSalt = "privatePasswordSalt"
	FieldPrvPassTok  = "privatePasswordToken"
	FieldPubPassSalt = "publicPasswordSalt"
	FieldPubPassTok  = "publicPasswordToken"
	FieldRemPassSalt = "removePasswordSalt"
	FieldRemPassTok  = "removePasswordToken"
	IsBLOCKED        = "true"
	FieldPublic      = "pub"
	FieldPrivate     = "prv"
	FieldRemove      = "rem"
	FieldURL         = "url"
	SuffixPublic     = "#pub"
	SuffixPrivate    = "#prv"
	SuffixRemove     = "#rem"
	SuffixURL        = "#url"
)

func (t *stringTuple) Set2Bytes(field string, value []byte, compress bool) error {
	return t.Set2(field, string(value), compress)
}
func (t *stringTuple) Set2(field, value string, compress bool) error {
	in := bytes.NewBuffer([]byte(value))
	if compress {
		in.Reset()
		s := base64.StdEncoding.EncodeToString([]byte(value))
		b := []byte(s)
		w, err := zlib.NewWriterLevel(in, zlib.BestCompression)
		if err != nil {
			return errors.Wrapf(err, "zlib writer failed")
		}
		w.Write(b)
		w.Close()
		t.tuple[field+fCompress] = "" // indicate this field is compressed
	}
	v := base64.StdEncoding.EncodeToString(in.Bytes())
	t.tuple[field] = v
	return nil
}

func (t *stringTuple) Get2Bytes(field string) ([]byte, error) {
	v, err := t.Get2(field)
	if err != nil {
		return nil, err
	}
	return []byte(v), err
}
func (t *stringTuple) Get2(field string) (string, error) {
	d, ok := t.tuple[field]
	if !ok {
		return "", errors.Wrapf(KeyNotFound, " "+field)
	}
	v := bytes.NewBufferString(d)
	if _, ok := t.tuple[field+fCompress]; ok { // field value is compressed
		res, err := base64.StdEncoding.DecodeString(v.String())
		if err != nil {
			return "", errors.Wrapf(err, "tuple value decode base64 failed for field <%s>", field)
		}
		out := bytes.NewBuffer(res)
		r, err := zlib.NewReader(out)
		if err != nil {
			return "", errors.Wrapf(err, "tuple value reading compressed failed for field <%s>", field)
		}
		io.Copy(out, r)
		v = out
	}
	res, err := base64.StdEncoding.DecodeString(v.String())
	if err != nil {
		return "", errors.Wrapf(err, "tuple value decode base64 failed for field <%s>", field)
	}
	return string(res), nil

}
func (t *stringTuple) MustGet(field string) string {
	r, err := t.Get2(field)
	if err != nil && err != KeyNotFound {
		r = t.tuple[field]
	}
	return r
}

func (t *stringTuple) FromString(asString string) error {
	return json.Unmarshal([]byte(asString), &t.tuple)
}

func (t *stringTuple) ToString() string {
	r, err := json.Marshal(&t.tuple)
	if err != nil {
		return ""
	}
	return string(r)
}
