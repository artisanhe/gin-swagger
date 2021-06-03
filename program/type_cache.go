package program

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/loader"
)

type tp struct {
	Expr ast.Expr
	Str  string
}

func newTypeCache(pkgs map[*types.Package]*loader.PackageInfo) map[types.Type]tp {
	tMap := make(map[types.Type]tp, len(pkgs))
	for _, pkgInfo := range pkgs {
		for e, t := range pkgInfo.Types {
			tMap[t.Type] = tp{
				Expr: e,
				Str:  t.Type.String(),
			}
		}
	}
	return tMap
}
