/// Copyright 2017 the Istio Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aspect

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	dpb "istio.io/api/mixer/v1/config/descriptor"
	"istio.io/mixer/pkg/adapter"
	atest "istio.io/mixer/pkg/adapter/test"
	aconfig "istio.io/mixer/pkg/aspect/config"
	"istio.io/mixer/pkg/aspect/test"
	"istio.io/mixer/pkg/attribute"
	cpb "istio.io/mixer/pkg/config/proto"
	"istio.io/mixer/pkg/expr"
)

type fakeaspect struct {
	adapter.Aspect
	closed bool
	body   func([]adapter.Value) error
}

func (a *fakeaspect) Close() error {
	a.closed = true
	return nil
}

func (a *fakeaspect) Record(v []adapter.Value) error {
	return a.body(v)
}

type fakeBuilder struct {
	adapter.Builder
	name string

	body func() (adapter.MetricsAspect, error)
}

func (b *fakeBuilder) Name() string {
	return b.name
}

func (b *fakeBuilder) NewMetricsAspect(env adapter.Env, config adapter.Config,
	metrics map[string]*adapter.MetricDefinition) (adapter.MetricsAspect, error) {
	return b.body()
}

func TestNewMetricsManager(t *testing.T) {
	m := newMetricsManager()
	if m.Kind() != MetricsKind {
		t.Errorf("m.Kind() = %s wanted %s", m.Kind(), MetricsKind)
	}
	if err := m.ValidateConfig(m.DefaultConfig()); err != nil {
		t.Errorf("m.ValidateConfig(m.DefaultConfig()) = %v; wanted no err", err)
	}
}

func TestMetricsManager_NewAspect(t *testing.T) {
	conf := &cpb.Combined{
		Aspect: &cpb.Aspect{
			Params: &aconfig.MetricsParams{
				Metrics: []*aconfig.MetricsParams_Metric{
					{
						DescriptorName: "request_count",
						Value:          "",
						Labels:         map[string]string{"source": "", "target": ""},
					},
				},
			},
		},
		// the params we use here don't matter because we're faking the aspect
		Builder: &cpb.Adapter{Params: &aconfig.MetricsParams{}},
	}
	builder := &fakeBuilder{name: "test", body: func() (adapter.MetricsAspect, error) {
		return &fakeaspect{body: func([]adapter.Value) error { return nil }}, nil
	}}
	if _, err := newMetricsManager().NewAspect(conf, builder, atest.NewEnv(t)); err != nil {
		t.Errorf("NewAspect(conf, builder, test.NewEnv(t)) = _, %v; wanted no err", err)
	}
}

func TestMetricsManager_NewAspect_PropagatesError(t *testing.T) {
	conf := &cpb.Combined{
		Aspect: &cpb.Aspect{Params: &aconfig.MetricsParams{}},
		// the params we use here don't matter because we're faking the aspect
		Builder: &cpb.Adapter{Params: &aconfig.MetricsParams{}},
	}
	errString := "expected"
	builder := &fakeBuilder{
		body: func() (adapter.MetricsAspect, error) {
			return nil, errors.New(errString)
		}}
	_, err := newMetricsManager().NewAspect(conf, builder, atest.NewEnv(t))
	if err == nil {
		t.Error("newMetricsManager().NewAspect(conf, builder, test.NewEnv(t)) = _, nil; wanted err")
	}
	if !strings.Contains(err.Error(), errString) {
		t.Errorf("NewAspect(conf, builder, test.NewEnv(t)) = _, %v; wanted err %s", err, errString)
	}
}

func TestMetricsWrapper_Execute(t *testing.T) {
	// TODO: all of these test values are hardcoded to match the metric definitions hardcoded in metricsManager
	// (since things have to line up for us to test them), they can be made dynamic when we get the ability to set the definitions
	goodEval := test.NewFakeEval(func(exp string, _ attribute.Bag) (interface{}, error) {
		switch exp {
		case "value":
			return 1, nil
		case "source":
			return "me", nil
		case "target":
			return "you", nil
		case "service":
			return "echo", nil
		default:
			return nil, fmt.Errorf("default case for exp = %s", exp)
		}
	})
	errEval := test.NewFakeEval(func(_ string, _ attribute.Bag) (interface{}, error) {
		return nil, errors.New("expected")
	})
	labelErrEval := test.NewFakeEval(func(exp string, _ attribute.Bag) (interface{}, error) {
		switch exp {
		case "value":
			return 1, nil
		default:
			return nil, errors.New("expected")
		}
	})

	goodMd := map[string]*metricInfo{
		"request_count": {
			definition: &adapter.MetricDefinition{Kind: adapter.Counter, Name: "request_count"},
			value:      "value",
			labels: map[string]string{
				"source":  "source",
				"target":  "target",
				"service": "service",
			},
		},
	}
	badGoodMd := map[string]*metricInfo{
		"bad": {
			definition: &adapter.MetricDefinition{Kind: adapter.Counter, Name: "bad"},
			value:      "bad",
			labels: map[string]string{
				"bad": "bad",
			},
		},
		"request_count": {
			definition: &adapter.MetricDefinition{Kind: adapter.Counter, Name: "request_count"},
			value:      "value",
			labels: map[string]string{
				"source":  "source",
				"target":  "target",
				"service": "service",
			},
		},
	}

	type o struct {
		value  interface{}
		labels []string
	}
	cases := []struct {
		mdin      map[string]*metricInfo
		recordErr error
		eval      expr.Evaluator
		out       map[string]o
		errString string
	}{
		{make(map[string]*metricInfo), nil, test.NewIDEval(), make(map[string]o), ""},
		{goodMd, nil, errEval, make(map[string]o), "expected"},
		{goodMd, nil, labelErrEval, make(map[string]o), "expected"},
		{goodMd, nil, goodEval, map[string]o{"request_count": {1, []string{"source", "target"}}}, ""},
		{goodMd, errors.New("record"), goodEval, map[string]o{"request_count": {1, []string{"source", "target"}}}, "record"},
		{badGoodMd, nil, goodEval, map[string]o{"request_count": {1, []string{"source", "target"}}}, "default case"},
	}
	for idx, c := range cases {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			var receivedValues []adapter.Value
			wrapper := &metricsWrapper{
				aspect: &fakeaspect{body: func(v []adapter.Value) error {
					receivedValues = v
					return c.recordErr
				}},
				metadata: c.mdin,
			}
			out := wrapper.Execute(test.NewBag(), c.eval, &ReportMethodArgs{})

			errString := out.Message()
			if !strings.Contains(errString, c.errString) {
				t.Errorf("wrapper.Execute(&fakeBag{}, eval) = _, %v; wanted error containing %s", out.Message(), c.errString)
			}

			if len(receivedValues) != len(c.out) {
				t.Errorf("wrapper.Execute(&fakeBag{}, eval) got vals %v, wanted at least %d", receivedValues, len(c.out))
			}
			for _, v := range receivedValues {
				o, found := c.out[v.Definition.Name]
				if !found {
					t.Errorf("Got unexpected value %v, wanted only %v", v, c.out)
				}
				if v.MetricValue != o.value {
					t.Errorf("v.MetricValue = %v; wanted %v", v.MetricValue, o.value)
				}
				for _, l := range o.labels {
					if _, found := v.Labels[l]; !found {
						t.Errorf("value.Labels = %v; wanted label named %s", v.Labels, l)
					}
				}
			}
		})
	}
}

func TestMetricsWrapper_Close(t *testing.T) {
	inner := &fakeaspect{closed: false}
	wrapper := &metricsWrapper{aspect: inner}
	if err := wrapper.Close(); err != nil {
		t.Errorf("wrapper.Close() = %v; wanted no err", err)
	}
	if !inner.closed {
		t.Error("metricsWrapper.Close() didn't close the aspect inside")
	}
}

func TestMetrics_DescToDef(t *testing.T) {
	cases := []struct {
		in        *dpb.MetricDescriptor
		out       *adapter.MetricDefinition
		errString string
	}{
		{&dpb.MetricDescriptor{}, nil, "METRIC_KIND_UNSPECIFIED"},
		{
			&dpb.MetricDescriptor{
				Name:   "bad label",
				Labels: []*dpb.LabelDescriptor{{ValueType: dpb.VALUE_TYPE_UNSPECIFIED}},
			},
			nil,
			"VALUE_TYPE_UNSPECIFIED",
		},
		{
			&dpb.MetricDescriptor{
				Name:   "bad metric kind",
				Kind:   dpb.METRIC_KIND_UNSPECIFIED,
				Labels: []*dpb.LabelDescriptor{{Name: "string", ValueType: dpb.STRING}},
			},
			nil,
			"METRIC_KIND_UNSPECIFIED",
		},
		{
			&dpb.MetricDescriptor{
				Name:   "good",
				Kind:   dpb.COUNTER,
				Value:  dpb.STRING,
				Labels: []*dpb.LabelDescriptor{{Name: "string", ValueType: dpb.STRING}},
			},
			&adapter.MetricDefinition{
				Name:   "good",
				Kind:   adapter.Counter,
				Labels: map[string]adapter.LabelType{"string": adapter.String},
			},
			""},
	}
	for idx, c := range cases {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			result, err := metricDefinitionFromProto(c.in)

			errString := ""
			if err != nil {
				errString = err.Error()
			}
			if !strings.Contains(errString, c.errString) {
				t.Errorf("metricsDescToDef(%v) = _, %v; wanted err containing %s", c.in, err, c.errString)
			}
			if !reflect.DeepEqual(result, c.out) {
				t.Errorf("metricsDescToDef(%v) = %v, %v; wanted %v", c.in, result, err, c.out)
			}
		})
	}
}

func TestMetrics_Find(t *testing.T) {
	cases := []struct {
		in   []*aconfig.MetricsParams_Metric
		find string
		out  bool
	}{
		{[]*aconfig.MetricsParams_Metric{}, "", false},
		{[]*aconfig.MetricsParams_Metric{{DescriptorName: "foo"}}, "foo", true},
		{[]*aconfig.MetricsParams_Metric{{DescriptorName: "bar"}}, "foo", false},
	}
	for _, c := range cases {
		t.Run(c.find, func(t *testing.T) {
			if _, found := findMetric(c.in, c.find); found != c.out {
				t.Errorf("find(%v, %s) = _, %t; wanted %t", c.in, c.find, found, c.out)
			}
		})
	}
}

func init() {
	// bump up the log level so log-only logic runs during the tests, for correctness and coverage.
	_ = flag.Lookup("v").Value.Set("99")
}
