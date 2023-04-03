package jq

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"log"
// 	"strings"

// 	"encoding/json"

// 	"github.com/influxdata/flux"
// 	"github.com/influxdata/flux/execute"
// 	"github.com/influxdata/flux/internal/function"
// 	"github.com/influxdata/flux/memory"
// 	"github.com/influxdata/flux/plan"
// 	"github.com/itchyny/gojq"
// )

// const FromKind = "from"

// type FromProcedureSpec struct {
// 	plan.DefaultCost
// 	// Config iox.Config
// 	Data  string
// 	Query string
// }

// func createFromProcedureSpec(args *function.Arguments) (function.Source, error) {
// 	data, err := args.GetRequiredString("data")
// 	if err != nil {
// 		return nil, err
// 	}

// 	query, err := args.GetRequiredString("query")
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &FromProcedureSpec{
// 		Data:  data,
// 		Query: query,
// 	}, nil
// }

// func (s *FromProcedureSpec) Kind() plan.ProcedureKind {
// 	return FromKind
// }

// func (s *FromProcedureSpec) Copy() plan.ProcedureSpec {
// 	ns := *s
// 	return &ns
// }

// func (s *FromProcedureSpec) CreateSource(id execute.DatasetID, a execute.Administration) (execute.Source, error) {
// 	getDataStream := func() (io.ReadCloser, error) {
// 		return io.NopCloser(strings.NewReader(s.Data)), nil
// 	}

// 	iterator := &fromIterator{getDataStream: getDataStream, id: id}
// 	return execute.CreateSourceFromIterator(iterator, id, a.Allocator())
// }

// type fromIterator struct {
// 	getDataStream func() (io.ReadCloser, error)
// 	id            execute.DatasetID
// 	alloc         memory.Allocator
// }

// func (c *fromIterator) Do(ctx context.Context, f func(flux.Table) error) error {
// 	query, err := gojq.Parse(".foo | ..")
// 	r, err := c.getDataStream()
// 	dec := json.NewDecoder(r)
// 	var v any
// 	err = dec.Decode(&v)
// 	iter := query.Run(v)

// 	groupKey := execute.NewGroupKey(nil, nil)
// 	builder := execute.NewColListTableBuilder(groupKey, c.alloc)
// 	builder.AddCol(flux.ColMeta{})

// 	for {
// 		v, ok := iter.Next()
// 		if !ok {
// 			break
// 		}
// 		if err, ok := v.(error); ok {
// 			log.Fatalln(err)
// 		}
// 		fmt.Printf("%#v\n", v)
// 	}

// 	// table, err := c.read(ctx, rows)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// return f(table)
// 	return err
// }

// // type sqlIterator struct {
// // 	execute.ExecutionNode
// // 	d *execute.TransportDataset

// // 	getDataStream func() (io.ReadCloser, error)
// // 	query         string
// // 	mem           memory.Allocator
// // }

// // func (s *fromSource) AddTransformation(t execute.Transformation) {
// // 	s.d.AddTransformation(t)
// // }

// // func (s *fromSource) Run(ctx context.Context) {
// // 	err := s.run(ctx)
// // 	s.d.Finish(err)
// // }

// // func (s *fromSource) createSchema(schema *stdarrow.Schema) ([]flux.ColMeta, error) {
// // 	fields := schema.Fields()
// // 	cols := make([]flux.ColMeta, len(fields))
// // 	for i, f := range fields {
// // 		cols[i].Label = f.Name
// // 		switch id := f.Type.ID(); id {
// // 		case stdarrow.INT64:
// // 			cols[i].Type = flux.TInt
// // 		case stdarrow.UINT64:
// // 			cols[i].Type = flux.TUInt
// // 		case stdarrow.FLOAT64:
// // 			cols[i].Type = flux.TFloat
// // 		case stdarrow.STRING:
// // 			cols[i].Type = flux.TString
// // 		case stdarrow.BOOL:
// // 			cols[i].Type = flux.TBool
// // 		case stdarrow.TIMESTAMP:
// // 			cols[i].Type = flux.TTime
// // 		default:
// // 			return nil, errors.Newf(codes.Internal, "unsupported arrow type %v", id)
// // 		}
// // 	}
// // 	return cols, nil
// // }

// // func (s *fromSource) run(ctx context.Context) error {
// // 	span, ctx := opentracing.StartSpanFromContext(ctx, "fromSource.run")
// // 	defer span.Finish()
// // 	span.LogFields(log.String("query", s.query))

// // 	cols, err := s.createSchema()
// // 	if err != nil {
// // 		return err
// // 	}
// // 	key := execute.NewGroupKey(nil, nil)

// // 	for hasMore && err == nil {
// // 		if err := s.produce(key, cols, rr.Record()); err != nil {
// // 			ext.LogError(span, err)
// // 			return err
// // 		}
// // 		hasMore, err = nextRecordBatch(rr)
// // 	}
// // 	if err != nil {
// // 		ext.LogError(span, err)
// // 		return err
// // 	}

// // 	return nil
// // }

// // func nextRecordBatch(rr iox.RecordReader) (bool, error) {
// // 	n := rr.Next()
// // 	if n {
// // 		return true, nil
// // 	}

// // 	return false, rr.Err()
// // }

// // func (s *fromSource) produce(key flux.GroupKey, cols []flux.ColMeta, record stdarrow.Record) error {
// // 	buffer := arrow.TableBuffer{
// // 		GroupKey: key,
// // 		Columns:  cols,
// // 		Values:   make([]array.Array, len(cols)),
// // 	}
// // 	for i := range buffer.Columns {
// // 		data := record.Column(i)
// // 		switch id := data.DataType().ID(); id {
// // 		case stdarrow.BOOL, stdarrow.INT64, stdarrow.UINT64, stdarrow.FLOAT64:
// // 			// We can just use the data as-is.
// // 			buffer.Values[i] = data
// // 		case stdarrow.TIMESTAMP:
// // 			// IOx returns time columns as Timestamp arrays, but they are really just
// // 			// int64 arrays under the hood, so this is safe.
// // 			// No need to retain here since calling NewInt64Data will bump the reference
// // 			// count on the underlying data.
// // 			rawData := data.(*arrowarray.Timestamp).Data()
// // 			rawData.Reset(stdarrow.PrimitiveTypes.Int64, rawData.Len(), rawData.Buffers(), nil, data.NullN(), rawData.Offset())
// // 			buffer.Values[i] = arrowarray.NewInt64Data(rawData)
// // 		case stdarrow.STRING:
// // 			// IOx returns string columns as String arrays, but Flux uses
// // 			// Binary arrays. The underlying structure of the buffers is the same.
// // 			binaryData := arrowarray.NewBinaryData(data.Data())
// // 			buffer.Values[i] = array.NewStringFromBinaryArray(binaryData)
// // 			binaryData.Release() // The String in data now owns this binary data.
// // 		default:
// // 			return errors.Newf(codes.FailedPrecondition, "unsupported arrow data type %v", id)
// // 		}
// // 		buffer.Values[i].Retain()
// // 	}

// // 	chunk := table.ChunkFromBuffer(buffer)
// // 	return s.d.Process(chunk)
// // }
