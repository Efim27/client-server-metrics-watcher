package main

import (
	"fmt"
	"strings"

	"github.com/gostaticanalysis/sqlrows/passes/sqlrows"
	"github.com/nishanths/predeclared/passes/predeclared"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/staticcheck"
)

type MultiCheckerRules []*analysis.Analyzer

func (m *MultiCheckerRules) printCount() {
	fmt.Printf("Loaded %d checkers \n", len(*m))
}

func (m *MultiCheckerRules) addPassesRules() {
	*m = append(*m,
		unmarshal.Analyzer,
		stringintconv.Analyzer,
		unreachable.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer)
}

func (m *MultiCheckerRules) addStaticCheckRulesSA() {
	for _, v := range staticcheck.Analyzers {
		if strings.Contains(v.Analyzer.Name, "SA") {
			*m = append(*m, v.Analyzer)
		}
	}
}

func (m *MultiCheckerRules) addStaticCheckRulesQT() {
	for _, v := range quickfix.Analyzers {
		*m = append(*m, v.Analyzer)
	}
}

func main() {
	var checkerRules MultiCheckerRules

	checkerRules.addPassesRules()
	checkerRules.addStaticCheckRulesSA()
	checkerRules.addStaticCheckRulesQT()

	checkerRules = append(checkerRules, sqlrows.Analyzer)
	checkerRules = append(checkerRules, predeclared.Analyzer)

	checkerRules.printCount()
	multichecker.Main(
		checkerRules...,
	)
}
