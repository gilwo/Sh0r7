package handler

type resTri interface {
	String() string
	True() resTri
	False() resTri
	Nil() resTri
	IsTrue() bool
	IsFalse() bool
	IsNil() bool
}
type _resTri struct {
	v    *bool
	t, f bool
}

func ResTri() resTri {
	return &_resTri{t: true, f: false}
}
func (r *_resTri) String() string {
	if r.v == nil {
		return "invalid"
	} else if *r.v {
		return "true"
	} else {
		return "false"
	}
}

func (r *_resTri) True() resTri {
	r.v = &r.t
	return r
}

func (r *_resTri) False() resTri {
	r.v = &r.f
	return r
}

func (r *_resTri) Nil() resTri {
	r.v = nil
	return r
}

func (r *_resTri) IsTrue() bool {
	return r.v != nil && *r.v
}

func (r *_resTri) IsFalse() bool {
	return r.v != nil && !*r.v
}

func (r *_resTri) IsNil() bool {
	return r.v == nil
}
