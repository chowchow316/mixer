// Copyright 2016 Istio Authors
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
	"istio.io/mixer/pkg/adapter"
	aconfig "istio.io/mixer/pkg/aspect/config"
	"istio.io/mixer/pkg/attribute"
	"istio.io/mixer/pkg/config"
	cpb "istio.io/mixer/pkg/config/proto"
	"istio.io/mixer/pkg/expr"
)

type (
	denialsManager struct{}

	denialsWrapper struct {
		aspect adapter.DenialsAspect
	}
)

// newDenialsManager returns a manager for the denials aspect.
func newDenialsManager() Manager {
	return denialsManager{}
}

// NewAspect creates a denyChecker aspect.
func (denialsManager) NewAspect(cfg *cpb.Combined, ga adapter.Builder, env adapter.Env) (Wrapper, error) {
	aa := ga.(adapter.DenialsBuilder)
	var asp adapter.DenialsAspect
	var err error

	if asp, err = aa.NewDenialsAspect(env, cfg.Builder.Params.(config.AspectParams)); err != nil {
		return nil, err
	}

	return &denialsWrapper{
		aspect: asp,
	}, nil
}

func (denialsManager) Kind() Kind                                                      { return DenialsKind }
func (denialsManager) DefaultConfig() config.AspectParams                              { return &aconfig.DenialsParams{} }
func (denialsManager) ValidateConfig(c config.AspectParams) (ce *adapter.ConfigErrors) { return }

func (a *denialsWrapper) Execute(attrs attribute.Bag, mapper expr.Evaluator, ma APIMethodArgs) Output {
	return Output{Status: a.aspect.Deny()}
}

func (a *denialsWrapper) Close() error { return a.aspect.Close() }
