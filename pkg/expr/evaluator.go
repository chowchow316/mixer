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

package expr

import "istio.io/mixer/pkg/attribute"

type (
	// Evaluator evaluates an expression written in the implementation defined
	// expression language. It uses attribute.Bag as variable bindings.
	Evaluator interface {
		// Eval evaluates given expression using the attribute bag
		Eval(expr string, attrs attribute.Bag) (interface{}, error)

		// Eval evaluates given expression using the attribute bag to a string
		EvalString(expr string, attrs attribute.Bag) (string, error)

		PredicateEvaluator

		Validator
	}

	// PredicateEvaluator evaluates a predicate to true or false
	PredicateEvaluator interface {
		// EvalPredicate evaluates given predicate using the attribute bag
		EvalPredicate(expr string, attrs attribute.Bag) (bool, error)
	}

	// Validator validates a given expression
	Validator interface {
		// Validate ensures that the given expression is syntactically correct
		Validate(expr string) error
	}
)
