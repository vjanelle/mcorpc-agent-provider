package tengo

import (
	"fmt"

	"github.com/d5/tengo/compiler/token"
	"github.com/d5/tengo/objects"
)

type Reply struct {
	data map[string]interface{}
}

func (r *Reply) Data() map[string]interface{} {
	return r.data
}

func (r *Reply) TypeName() string {
	return "reply"
}

func (r *Reply) String() string {
	return fmt.Sprintf("%#v", r.data)
}

func (r *Reply) BinaryOp(op token.Token, rhs objects.Object) (res objects.Object, err error) {
	return nil, objects.ErrInvalidOperator
}

func (r *Reply) IsFalsy() bool {
	return len(r.data) == 0
}

func (r *Reply) Equals(o objects.Object) bool {
	return false
}

func (r *Reply) Copy() objects.Object {
	new := &Reply{
		data: make(map[string]interface{}),
	}

	for item, value := range r.data {
		new.data[item] = value
	}

	return new
}

func (r *Reply) IndexGet(index objects.Object) (res objects.Object, err error) {
	strIdx, ok := index.(*objects.String)
	if !ok {
		return nil, objects.ErrInvalidIndexType
	}

	value, ok := r.data[strIdx.Value]
	if !ok {
		return objects.UndefinedValue, nil
	}

	obj, err := objects.FromInterface(value)
	if err != nil {
		return objects.UndefinedValue, nil
	}

	return obj, nil
}

func (r *Reply) IndexSet(index, value objects.Object) (err error) {
	strIdx, ok := index.(*objects.String)
	if !ok {
		return objects.ErrInvalidTypeConversion
	}

	_, ok = r.data[strIdx.Value]
	if !ok {
		return fmt.Errorf("undeclared output item %s", strIdx.Value)
	}

	r.data[strIdx.Value] = value

	return nil
}
