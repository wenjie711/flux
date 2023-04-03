package jq_test

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	_ "github.com/influxdata/flux/fluxinit/static" // We need to init flux for the tests to work.
	"github.com/influxdata/flux/internal/operation"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/contrib/wuhan/jq"
	"github.com/influxdata/flux/stdlib/csv"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
)

func TestFromCSV_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "from no args",
			Raw:     `import "csv" csv.from()`,
			WantErr: true,
		},
		{
			Name:    "from conflicting args",
			Raw:     `import "csv" csv.from(csv:"d", file:"b")`,
			WantErr: true,
		},
		{
			Name:    "from repeat arg",
			Raw:     `import "csv" csv.from(csv:"telegraf", csv:"oops")`,
			WantErr: true,
		},
		{
			Name:    "from",
			Raw:     `import "csv" csv.from(csv:"telegraf", chicken:"what is this?")`,
			WantErr: true,
		},
		{
			Name: "fromCSV text",
			Raw:  `import "csv" csv.from(csv: "1,2") |> range(start:-4h, stop:-2h) |> sum()`,
			Want: &operation.Spec{
				Operations: []*operation.Node{
					{
						ID: "fromCSV0",
						Spec: &csv.FromCSVOpSpec{
							CSV:  "1,2",
							Mode: "annotations",
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							SimpleAggregateConfig: execute.DefaultSimpleAggregateConfig,
						},
					},
				},
				Edges: []operation.Edge{
					{Parent: "fromCSV0", Child: "range1"},
					{Parent: "range1", Child: "sum2"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestFrom(t *testing.T) {
	proxyUrl, err := url.Parse("http://127.0.0.1:8111")
	http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(proxyUrl)
	http.DefaultClient.Transport = http.DefaultTransport
	resp, err := http.Get("http://bvc-ost-oss.bilibili.co/api/bban2/v1/schedule")
	if err != nil {
		t.Fatal(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return
	}
	//  log.Printf("%v===9==\n", body)
	// query := ".data[1]|{\"country\": .country, \"node\": (.nodes[]|.id)}|[.country, .node]|@csv"
	//query := ".data[1]|{\"country\": .country, \"node\": (.nodes[]|.id)}|[.country, .node]|@csv"
	query := ".data[3]"
	obj := values.NewObjectWithValues(map[string]values.Value{
		"data":  values.NewString(string(body)),
		"query": values.NewString(query),
	})
	args := interpreter.NewArguments(obj)
	jq.From(args)
}
