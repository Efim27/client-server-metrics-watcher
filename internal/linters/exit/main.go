package exit

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
	"honnef.co/go/tools/analysis/code"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exit",
	Doc:  "check for os.Exit calls in main func",
	Run:  run,
}

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

func CallFuncName(pass *analysis.Pass, node *ast.CallExpr) (funcName string) {
	funcExpr := astutil.Unparen(node.Fun)
	nodeIdent, ok := funcExpr.(*ast.Ident)
	if !ok {
		return
	}

	typeObj := pass.TypesInfo.ObjectOf(nodeIdent)
	funcType, ok := typeObj.(*types.Func)
	if !ok {
		return
	}

	return funcType.Name()
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			if node == nil || !isMainFunc(node) {
				return true
			}

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
