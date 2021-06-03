package program

import (
	"go/ast"
	"go/token"
	"go/types"
	"sort"

	"golang.org/x/tools/go/loader"
)

type astFile struct {
	pkg     *types.Package
	pkgInfo *loader.PackageInfo
	file    *ast.File
	pos     token.Pos
}

type astFiles []*astFile

func (files astFiles) Len() int           { return len(files) }
func (files astFiles) Less(i, j int) bool { return files[i].pos < files[j].pos }
func (files astFiles) Swap(i, j int)      { files[i], files[j] = files[j], files[i] }

func newASTFiles(pkgs map[*types.Package]*loader.PackageInfo) astFiles {
	files := make(astFiles, 0, len(pkgs))
	for pkg, pkgInfo := range pkgs {
		for _, file := range pkgInfo.Files {
			file := &astFile{
				pkg:     pkg,
				pkgInfo: pkgInfo,
				file:    file,
				pos:     file.Pos(),
			}
			files = append(files, file)
		}
	}

	sort.Sort(files)
	if files.Len() > 1 {
		end := files[len(files)-1].file.End()
		files = append(files, &astFile{pos: end})
	}

	return files
}

func (files astFiles) searchByPos(pos token.Pos) *astFile {
	i := sort.Search(files.Len(), func(i int) bool {
		return files[i].pos >= pos
	})

	// not found
	if i > files.Len()-1 {
		return nil
	}

	switch {
	case files[i].pos == pos: // pos = file.pos
		return files[i]
	case i == 0: // pos < min(file.pos)
		return nil
	default: // file.pos < pos < file(i+1).pos
		return files[i-1]
	}
}
