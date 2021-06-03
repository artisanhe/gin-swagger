package program

import (
	"go/token"
	"go/types"
	"sort"

	"golang.org/x/tools/go/loader"
)

type fn struct {
	pkg     *types.Package
	pkgInfo *loader.PackageInfo
	tfn     *types.Func
	pos     token.Pos
}

type fns []*fn

func (f fns) Len() int           { return len(f) }
func (f fns) Less(i, j int) bool { return f[i].pos < f[j].pos }
func (f fns) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func newFns(pkgs map[*types.Package]*loader.PackageInfo) fns {
	fs := make(fns, 0, len(pkgs))
	for pkg, pkgInfo := range pkgs {
		for _, obj := range pkgInfo.Defs {
			if tpeFunc, ok := obj.(*types.Func); ok {
				f := &fn{
					pkg:     pkg,
					pkgInfo: pkgInfo,
					tfn:     tpeFunc,
					pos:     tpeFunc.Scope().Pos(),
				}
				fs = append(fs, f)
			}
		}
	}

	sort.Sort(fs)
	if fs.Len() > 1 {
		end := fs[len(fs)-1].tfn.Scope().End()
		fs = append(fs, &fn{pos: end})
	}
	return fs
}

func (f fns) searchByPos(pos token.Pos) *fn {
	i := sort.Search(f.Len(), func(i int) bool {
		return f[i].pos >= pos
	})

	// not found
	if i > f.Len()-1 {
		return nil
	}

	switch {
	case f[i].pos == pos: // pos = fn.pos
		return f[i]
	case i == 0: // pos < min(fn.pos)
		return nil
	default: // fn.pos < pos < fn(i+1).pos
		return f[i-1]
	}
}
