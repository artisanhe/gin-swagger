package program

import (
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"sort"
	"strings"

	"git.chinawayltd.com/golib/gin-swagger/codegen"

	"github.com/logrusorgru/aurora"
	"golang.org/x/tools/go/loader"
)

type Program struct {
	PackagePath string
	*loader.Program
	astFiles astFiles
	fns      fns
	tys      map[types.Type]tp
}

func PkgContains(pkgs []*types.Package, targetPkg *types.Package) bool {
	for _, pkg := range pkgs {
		if pkg == targetPkg {
			return true
		}
	}
	return false
}

func NewProgram(packagePath string, allowErrors bool) *Program {
	ldr := loader.Config{
		AllowErrors: allowErrors,
	}

	ldr.ParserMode = parser.ParseComments
	ldr.Import(packagePath)

	prog, err := ldr.Load()
	if err != nil {
		panic(err)
	}

	// 根据特殊情况可以不扫描 vendor 部分代码，以加速扫描效率
	//prog.AllPackages = deleteBuiltinAndVendorPackages(packagePath, prog.AllPackages)
	return &Program{
		PackagePath: packagePath,
		Program:     prog,
		astFiles:    newASTFiles(prog.AllPackages),
		fns:         newFns(prog.AllPackages),
		tys:         newTypeCache(prog.AllPackages),
	}
}

func deleteBuiltinAndVendorPackages(packagePath string, pkgs map[*types.Package]*loader.PackageInfo) map[*types.Package]*loader.PackageInfo {
	var (
		bizPackages    = make(map[*types.Package]*loader.PackageInfo, len(pkgs))
		bizVendorPath  = packagePath + "/vendor"
		typeVendorPath = packagePath + "/vendor/" + strings.Split(packagePath, "/")[0]
	)

	for pkg, pkgInfo := range pkgs {
		if pkg != nil {
			pkgPath := pkg.Path()
			if strings.HasPrefix(pkgPath, typeVendorPath) || (strings.HasPrefix(pkgPath, packagePath) &&
				!strings.HasPrefix(pkgPath, bizVendorPath)) {
				bizPackages[pkg] = pkgInfo
			}
		}
	}

	return bizPackages
}

func (program *Program) TypeString(t types.Type) string {
	ty, ok := program.tys[t]
	if ok {
		return ty.Str
	}
	return ""
}

func (program *Program) TypeOf(e ast.Expr) types.Type {
	pkgInfo := program.Package(program.PackageOf(e).Path())

	if tpe := pkgInfo.TypeOf(e); tpe != nil {
		return tpe
	}

	return nil
}

func (program *Program) ValueOf(e ast.Expr) constant.Value {
	pkgInfo := program.Package(program.PackageOf(e).Path())

	if t, ok := pkgInfo.Types[e]; ok {
		if t.Value != nil {
			return t.Value
		}
	}

	return nil
}

func (program *Program) ScopeOf(targetNode ast.Node) *types.Scope {
	pkgInfo := program.Package(program.PackageOf(targetNode).Path())

	for _, scope := range pkgInfo.Scopes {
		if funcDecl, ok := targetNode.(*ast.FuncDecl); ok {
			if funcDecl.Body.Pos() == scope.Pos() {
				return scope
			}
		} else if targetNode.Pos() == scope.Pos() {
			return scope
		}
	}

	return nil
}

func (program *Program) ObjectOf(id *ast.Ident) types.Object {
	pkgInfo := program.Package(program.PackageOf(id).Path())
	obj := pkgInfo.ObjectOf(id)
	return obj
}

func (program *Program) DefOf(id *ast.Ident) types.Object {
	obj := program.ObjectOf(id)

	// find the defined
	switch obj.Type().(type) {
	case *types.Pointer:
		return obj.Type().(*types.Pointer).Elem().(*types.Named).Obj()
	case *types.Named:
		return obj.Type().(*types.Named).Obj()
	default:
		return obj
	}

}

func (program *Program) WhereDecl(targetTpe types.Type) ast.Expr {
	switch targetTpe.(type) {
	case *types.Named:
		namedType := targetTpe.(*types.Named)
		return program.IdentOf(namedType.Obj())
	case *types.Struct:
		if ty, ok := program.tys[targetTpe]; ok {
			return ty.Expr
		}
	default:
		log.Println(aurora.Sprintf(aurora.Red("not found %v"), targetTpe))
	}

	return nil
}

func (program *Program) IdentOf(targetDef types.Object) *ast.Ident {
	pkgInfo := program.Package(targetDef.Pkg().Path())

	for id, def := range pkgInfo.Defs {
		if def == targetDef {
			return id
		}
	}

	return nil
}

func (program *Program) FileOf(node ast.Node) *ast.File {
	file := program.astFiles.searchByPos(node.Pos())
	if file == nil {
		return nil
	}
	return file.file
}

func (program *Program) PackageOf(node ast.Node) *types.Package {
	file := program.astFiles.searchByPos(node.Pos())
	if file == nil {
		return nil
	}
	return file.pkg
}

func (program *Program) PackageInfoOf(node ast.Node) *loader.PackageInfo {
	file := program.astFiles.searchByPos(node.Pos())
	if file == nil {
		return nil
	}
	return file.pkgInfo
}

func (program *Program) WithPkgInfoContains(pos token.Pos) *loader.PackageInfo {
	file := program.astFiles.searchByPos(pos)
	if file == nil {
		return nil
	}
	return file.pkgInfo
}

func (program *Program) WitchFunc(pos token.Pos) *types.Func {
	fn := program.fns.searchByPos(pos)
	if fn == nil {
		return nil
	}
	return fn.tfn
}

func (program *Program) CallExprById(id *ast.Ident) *ast.CallExpr {
	pkgInfo := program.WithPkgInfoContains(id.Pos())

	for expr := range pkgInfo.Types {
		if callExpr, ok := expr.(*ast.CallExpr); ok {
			switch callExpr.Fun.(type) {
			case *ast.Ident:
				if callExpr.Fun == id {
					return callExpr
				}
			case *ast.SelectorExpr:
				selectorExpr := callExpr.Fun.(*ast.SelectorExpr)
				if selectorExpr.Sel == id {
					return callExpr
				}
			}
		}
	}

	return nil
}

func (program *Program) DefsInScope(scope *types.Scope) map[*ast.Ident]types.Object {
	defs := map[*ast.Ident]types.Object{}
	pkgInfo := program.WithPkgInfoContains(scope.Pos())

	if pkgInfo != nil {
		for id, def := range pkgInfo.Defs {
			if scope.Contains(id.Pos()) {
				defs[id] = def
			}
		}
	}

	return defs
}

func (program *Program) UsesInScope(scope *types.Scope) map[*ast.Ident]types.Object {
	uses := map[*ast.Ident]types.Object{}
	pkgInfo := program.WithPkgInfoContains(scope.Pos())

	if pkgInfo != nil {
		for id, obj := range pkgInfo.Uses {
			if scope.Contains(id.Pos()) {
				uses[id] = obj
			}
		}
	}

	return uses
}

func (program *Program) SelectionsInScope(scope *types.Scope) map[*ast.SelectorExpr]*types.Selection {
	selections := map[*ast.SelectorExpr]*types.Selection{}
	pkgInfo := program.WithPkgInfoContains(scope.Pos())

	if pkgInfo != nil {
		for selectorExpr, selection := range pkgInfo.Selections {
			if scope.Contains(selectorExpr.Pos()) {
				selections[selectorExpr] = selection
			}
		}
	}

	return selections
}

type Option struct {
	V     interface{} `json:"v"`
	Value interface{} `json:"value"`
	Label string      `json:"label"`
}

func (program *Program) GetEnumOptionsByType(node ast.Node) (list []Option) {
	if ident, ok := node.(*ast.Ident); ok {
		file := program.FileOf(node)

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)

			if !ok {
				continue
			}

			if genDecl.Tok.String() == "const" {
				for _, spec := range genDecl.Specs {
					switch spec.(type) {
					case *ast.ValueSpec:
						valueSpec, _ := spec.(*ast.ValueSpec)
						var name = valueSpec.Names[0].Name
						obj := program.ObjectOf(valueSpec.Names[0])
						constValue, _ := obj.(*types.Const)
						value, _ := GetConstValue(constValue.Val())

						if name != "_" {
							if strings.HasPrefix(name, codegen.ToUpperSnakeCase(ident.String())) {
								var values = strings.SplitN(name, "__", 2)
								if len(values) == 2 {
									list = append(list, Option{
										V:     value,
										Value: values[1],
										Label: strings.TrimSpace(valueSpec.Comment.Text()),
									})
								}
							} else if obj.Type() == program.TypeOf(ident) {
								list = append(list, Option{
									V:     value,
									Value: value,
									Label: strings.TrimSpace(valueSpec.Comment.Text()),
								})
							}
						}
					}

				}
			}
		}
	}
	return
}

func (program *Program) AstDeclOf(targetNode ast.Node) (ast.Decl, bool) {
	file := program.FileOf(targetNode)
	nodePos := targetNode.Pos()

	for _, decl := range file.Decls {
		if nodePos > decl.Pos() && nodePos < decl.End() {
			return decl, true
		}
	}

	return nil, false
}

func (program *Program) MethodsOf(tpe types.Type) (methods map[*types.Func]*types.Signature) {
	methods = map[*types.Func]*types.Signature{}
	for _, pkgInfo := range program.AllPackages {
		for _, def := range pkgInfo.Defs {
			if funcType, ok := def.(*types.Func); ok {
				if s, ok := funcType.Type().(*types.Signature); ok {
					recv := s.Recv()
					if recv != nil && recv.Type() == tpe {
						methods[funcType] = s
					}
				}
			}
		}
	}
	return
}

type ByCommentPos []*ast.CommentGroup

func (a ByCommentPos) Len() int {
	return len(a)
}
func (a ByCommentPos) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ByCommentPos) Less(i, j int) bool {
	return a[i].Pos() < a[j].Pos()
}

func getCommentsFor(file ast.Node, targetNode ast.Node, commentMap ast.CommentMap) (commentList []*ast.CommentGroup) {
	switch targetNode.(type) {
	// Spec should should merge with comments of its parent Decl
	case ast.Spec:
		for node, comments := range commentMap {
			if genDecl, ok := node.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs {
					if targetNode == spec {
						commentList = append(commentList, comments...)
					}
				}
			}
		}
		if comments, ok := commentMap[targetNode]; ok {
			commentList = append(commentList, comments...)
		}
		// Node has comments
	case *ast.File, *ast.Field, ast.Stmt, ast.Decl:
		if comments, ok := commentMap[targetNode]; ok {
			commentList = comments
		}
	default:
		var deltaPos token.Pos
		var parentNode ast.Node

		deltaPos = -1

		ast.Inspect(file, func(node ast.Node) bool {
			switch node.(type) {
			case *ast.Field, ast.Decl, ast.Spec, ast.Stmt:
				if targetNode.Pos() >= node.Pos() && targetNode.End() <= node.End() {
					nextDelta := targetNode.Pos() - node.Pos()
					if deltaPos == -1 || (nextDelta <= deltaPos) {
						deltaPos = nextDelta
						parentNode = node
					}
				}
			}
			return true
		})

		if parentNode != nil {
			commentList = getCommentsFor(file, parentNode, commentMap)
		}
	}

	sort.Sort(ByCommentPos(commentList))
	return
}

func (program *Program) CommentGroupFor(targetNode ast.Node) []*ast.CommentGroup {
	file := program.FileOf(targetNode)
	commentMap := ast.NewCommentMap(program.Fset, file, file.Comments)
	return getCommentsFor(file, targetNode, commentMap)
}
