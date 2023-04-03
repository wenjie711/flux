package jq

// import (
// 	"github.com/influxdata/flux/internal/function"
// )

// const pkgpath = "contrib/wuhan/jq"

// func init() {
// 	b := function.ForPackage(pkgpath)
// 	b.RegisterSource("from", FromKind, createFromProcedureSpec)
// }

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
	"github.com/itchyny/gojq"
)

const (
	dataArg  = "data"
	queryArg = "query"
	pkgName  = "contrib/wuhan/jq"
)

// function is a function definition
type function func(args interpreter.Arguments) (values.Value, error)

// makeFunction constructs a values.Function from a function definition.
func makeFunction(name string, fn function) values.Function {
	mt := runtime.MustLookupBuiltinType(pkgName, name)
	return values.NewFunction(name, mt, func(ctx context.Context, args values.Object) (values.Value, error) {
		return interpreter.DoFunctionCall(fn, args)
	}, false)
}

func init() {
	runtime.RegisterPackageValue(pkgName, "from", makeFunction("from", From))
}

func From(args interpreter.Arguments) (values.Value, error) {

	data, err := args.GetRequiredString(dataArg)
	if err != nil {
		return nil, err
	}

	query, err := args.GetRequiredString(queryArg)
	if err != nil {
		return nil, err
	}

	q, err := gojq.Parse(query)
	var v any

	err = json.Unmarshal([]byte(data), &v)
	if err != nil {
		return nil, err
	}

	iter := q.Run(v)

	var ret strings.Builder
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, err
		}
		fmt.Fprintf(&ret, "%v\n", v)
	}
	return values.NewString(ret.String()), nil
}
