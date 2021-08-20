package filter

import "fmt"

const (
	Eq  Operator = "="
	Neq Operator = "<>"
	Gt  Operator = ">"
	Lt  Operator = "<"
	Ge  Operator = ">="
	Le  Operator = "<="
)

type Operator string

type KV struct {
	k string
	v interface{}
	o Operator
}

func (f *KV) Build() (string, []interface{}, error) {
	return fmt.Sprintf("%s %s ?", f.k, f.o), []interface{}{f.v}, nil
}

func NewKV(k string, v interface{}, o Operator) *KV {
	return &KV{
		k: k,
		v: v,
		o: o,
	}
}

