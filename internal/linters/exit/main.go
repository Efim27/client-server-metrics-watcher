package exit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/analysis/code"
)

// ExitCheckAnalyzer prohibits the using os.Exit in main func (package main)
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exit",
	Doc:  "check os.Exit calls in main func",
	Run:  run,
}

// isMainFunc checks if the node is a main func
func isMainFunc(node ast.Node) bool {
	funcNode, ok := node.(*ast.FuncDecl)
	if !ok {
		return false
	}

	if funcNode.Name.Name != "main" {
		return false
	}

	return true
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		// find main func
		ast.Inspect(file, func(node ast.Node) bool {
			if node == nil || !isMainFunc(node) {
				return true
			}

			// check using os.Exit in main func
			ast.Inspect(node, func(node ast.Node) bool {
				funcCallNode, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				if code.CallName(pass, funcCallNode) == "os.Exit" {
					pass.Reportf(funcCallNode.Pos(), "os.Exit is not allowed in main")
				}

				return true
			})

			return true
		})
	}

	return nil, nil
}
