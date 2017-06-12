/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2016 Load Impact
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package stats

import (
	"encoding/json"

	"github.com/dop251/goja"
	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/parser"
	"github.com/pkg/errors"
)

const jsEnvSrc = `function p(pct) { return __sink__.P(pct/100.0); };`

var jsEnv *goja.Program

func init() {
	pgm, err := goja.Compile("__env__", jsEnvSrc, true)
	if err != nil {
		panic(err)
	}
	jsEnv = pgm
}

type refParser struct {
	seen map[string]bool
}

func parseRefs(apgm *ast.Program) []string {
	p := refParser{make(map[string]bool)}
	p.crawlStatements(apgm.Body...)
	refs := make([]string, 0, len(p.seen))
	for ref, _ := range p.seen {
		refs = append(refs, ref)
	}
	return refs
}

func (p *refParser) identify(id *ast.Identifier) {
	p.seen[id.Name] = true
}

func (p *refParser) crawlStatements(stmts ...ast.Statement) {
	for _, rawStmt := range stmts {
		switch stmt := rawStmt.(type) {
		case *ast.IfStatement:
			p.crawlExpressions(stmt.Test)
			p.crawlStatements(stmt.Consequent, stmt.Alternate)
		case *ast.BlockStatement:
			p.crawlStatements(stmt.List...)
		case *ast.BranchStatement:
			p.identify(stmt.Label)
		case *ast.CaseStatement:
			p.crawlExpressions(stmt.Test)
			p.crawlStatements(stmt.Consequent...)
		case *ast.CatchStatement:
			p.identify(stmt.Parameter)
			p.crawlStatements(stmt.Body)
		case *ast.DoWhileStatement:
			p.crawlExpressions(stmt.Test)
			p.crawlStatements(stmt.Body)
		case *ast.ExpressionStatement:
			p.crawlExpressions(stmt.Expression)
		case *ast.ForInStatement:
			p.crawlExpressions(stmt.Source, stmt.Into)
			p.crawlStatements(stmt.Body)
		case *ast.ForStatement:
			p.crawlExpressions(stmt.Initializer, stmt.Test, stmt.Update)
			p.crawlStatements(stmt.Body)
		case *ast.SwitchStatement:
			p.crawlExpressions(stmt.Discriminant)
			for _, cas := range stmt.Body {
				p.crawlExpressions(cas.Test)
				p.crawlStatements(cas.Consequent...)
			}
		case *ast.ThrowStatement:
			p.crawlExpressions(stmt.Argument)
		case *ast.TryStatement:
			p.crawlStatements(stmt.Body, stmt.Catch.Body, stmt.Finally)
		case *ast.VariableStatement:
			p.crawlExpressions(stmt.List...)
		case *ast.WhileStatement:
			p.crawlExpressions(stmt.Test)
			p.crawlStatements(stmt.Body)
		case *ast.WithStatement:
			p.crawlExpressions(stmt.Object)
			p.crawlStatements(stmt.Body)
		}
	}
}

func (p *refParser) crawlExpressions(exprs ...ast.Expression) {
	for _, rawExpr := range exprs {
		switch expr := rawExpr.(type) {
		case *ast.Identifier:
			p.identify(expr)
		case *ast.AssignExpression:
			p.crawlExpressions(expr.Left, expr.Right)
		case *ast.BinaryExpression:
			p.crawlExpressions(expr.Left, expr.Right)
		case *ast.BracketExpression:
			p.crawlExpressions(expr.Left, expr.Member)
		case *ast.CallExpression:
			p.crawlExpressions(append(expr.ArgumentList, expr.Callee)...)
		case *ast.ConditionalExpression:
			p.crawlExpressions(expr.Test, expr.Consequent, expr.Alternate)
		case *ast.DotExpression:
			// expr.Identifier is eg. `obj.IDENTIFIER` and isn't what we're looking for.
			p.crawlExpressions(expr.Left)
		case *ast.NewExpression:
			p.crawlExpressions(append(expr.ArgumentList, expr.Callee)...)
		case *ast.SequenceExpression:
			p.crawlExpressions(expr.Sequence...)
		case *ast.ThisExpression:
		case *ast.UnaryExpression:
			p.crawlExpressions(expr.Operand)
		case *ast.VariableExpression:
			// expr.Name declares a variable, we're looking for ones that *aren't* declared.
			p.crawlExpressions(expr.Initializer)
		}
	}
}

type Threshold struct {
	Source string
	Failed bool
	Refs   []string

	pgm *goja.Program
	rt  *goja.Runtime
}

func NewThreshold(src string, rt *goja.Runtime) (*Threshold, error) {
	apgm, err := parser.ParseFile(nil, "__threshold__", src, 0)
	if err != nil {
		return nil, err
	}
	refs := parseRefs(apgm)

	pgm, err := goja.Compile("__threshold__", src, true)
	if err != nil {
		return nil, err
	}

	return &Threshold{
		Source: src,
		Refs:   refs,
		pgm:    pgm,
		rt:     rt,
	}, nil
}

func (t Threshold) RunNoTaint() (bool, error) {
	v, err := t.rt.RunProgram(t.pgm)
	if err != nil {
		return false, err
	}
	return v.ToBoolean(), nil
}

func (t *Threshold) Run() (bool, error) {
	b, err := t.RunNoTaint()
	if !b {
		t.Failed = true
	}
	return b, err
}

type Thresholds struct {
	Runtime    *goja.Runtime
	Thresholds []*Threshold
}

func NewThresholds(sources []string) (Thresholds, error) {
	rt := goja.New()
	if _, err := rt.RunProgram(jsEnv); err != nil {
		return Thresholds{}, errors.Wrap(err, "builtin")
	}

	ts := make([]*Threshold, len(sources))
	for i, src := range sources {
		t, err := NewThreshold(src, rt)
		if err != nil {
			return Thresholds{}, errors.Wrapf(err, "%d", i)
		}
		ts[i] = t
	}
	return Thresholds{rt, ts}, nil
}

func (ts *Thresholds) UpdateVM(sink Sink) error {
	ts.Runtime.Set("__sink__", sink)
	for k, v := range sink.Format() {
		ts.Runtime.Set(k, v)
	}
	return nil
}

func (ts *Thresholds) RunAll() (bool, error) {
	succ := true
	for i, th := range ts.Thresholds {
		b, err := th.Run()
		if err != nil {
			return false, errors.Wrapf(err, "%d", i)
		}
		if !b {
			succ = false
		}
	}
	return succ, nil
}

func (ts *Thresholds) Run(sink Sink) (bool, error) {
	if err := ts.UpdateVM(sink); err != nil {
		return false, err
	}
	return ts.RunAll()
}

func (ts *Thresholds) UnmarshalJSON(data []byte) error {
	var sources []string
	if err := json.Unmarshal(data, &sources); err != nil {
		return err
	}

	newts, err := NewThresholds(sources)
	if err != nil {
		return err
	}
	*ts = newts
	return nil
}

func (ts Thresholds) MarshalJSON() ([]byte, error) {
	sources := make([]string, len(ts.Thresholds))
	for i, t := range ts.Thresholds {
		sources[i] = t.Source
	}
	return json.Marshal(sources)
}
