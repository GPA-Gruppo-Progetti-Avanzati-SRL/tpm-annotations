package jaxrs

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"tpm-annotations/annotations"
	annparser "tpm-annotations/annotations/parser"
)

func Scan(files []ScanSource) ([]FileDef, error) {

	fds := make([]FileDef, 0)

	var fs *token.FileSet
	fs = token.NewFileSet()

	for _, fn := range files {
		log.Info().Str("filename", fn.FileName).Msg("start processing file")
		imports := make([]importRef, 0)

		var f *ast.File
		var err error
		if fn.Source == "" {
			f, err = parser.ParseFile(fs, fn.FileName, nil, parser.ParseComments)
		} else {
			f, err = parser.ParseFile(fs, fn.FileName, fn.Source, parser.ParseComments)
		}

		if err != nil {
			log.Error().Err(err).Send()
			return nil, err
		}

		fd := FileDef{Name: fn.FileName, Package: f.Name.Name}
		a, err := parseComments(f.Doc)
		if err != nil {
			return nil, err
		} else {
			fd.Annotations = a
		}

		for _, i := range f.Imports {
			ir := importRef{path: strings.ReplaceAll(i.Path.Value, "\"", "")}
			if i.Name != nil {
				ir.alias = i.Name.Name
			}
			imports = append(imports, ir)
		}
		fd.imports = importResolver{imports}

		for _, decl := range f.Decls {
			switch tdecl := decl.(type) {
			case *ast.FuncDecl:
				fui, err := parseFuncDecl(fs, tdecl, f.Comments, &fd.imports)
				if err != nil {
					return nil, err
				} else {
					if fui.FType != NO_HANDLER {
						log.Info().Str("function", fui.Name).
							Str("at", fui.Pos.String()).
							Str("methodType", string(fui.FType)).
							Int("numParams", len(fui.Params)).
							Msg("found annotated method")
						fd.Methods = append(fd.Methods, fui)
					}
				}
			}
		}

		if len(fd.Methods) > 0 {
			fds = append(fds, fd)
		} else {
			log.Info().Str("filename", fn.FileName).Msg("no annotated methods could be found in file...skipping")
		}
		// ast.Print(fs, f)
	}

	// Sorting to process by directory package. Groups of files in the same directory / package will output
	// a file named by package (<package>_jaxrs_gen.go) in that directory. This way should be able to handle also
	// the test packages
	sortByPathPackage(fds)

	return fds, nil
}

func parseComments(comments *ast.CommentGroup) ([]annotations.Annotation, error) {

	if comments == nil {
		return nil, nil
	}

	var as = make([]annotations.Annotation, 0)
	for _, d := range comments.List {
		a, err := annparser.Parse(d.Text)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, err
		}

		if len(a) > 0 {
			as = append(as, a...)
		}
	}

	return as, nil
}

func parseFuncDecl(fs *token.FileSet, f *ast.FuncDecl, comments []*ast.CommentGroup, iResolver *importResolver) (FuncInfo, error) {

	fui := FuncInfo{}
	fui.FType = NO_HANDLER
	fui.Pos = fs.Position(f.Pos())
	fui.Name = f.Name.Name

	a, err := parseComments(f.Doc)
	if err != nil {
		return fui, err
	} else {
		if len(a) == 0 {
			log.Trace().Str("function", f.Name.Name).Msg("skipping function with no annotations decoration")
			return fui, nil
		} else {
			fui.Annotations = a
		}
	}

	if f.Recv != nil {
		log.Warn().Str("function", f.Name.Name).Str("at", fui.Pos.String()).Msg("struct methods not supported as handlers")
		return fui, nil
	}

	numParams := 0
	if f.Type.Params != nil {
		numParams = len(f.Type.Params.List)
	}

	numResults := 0
	if f.Type.Results != nil {
		numResults = len(f.Type.Results.List)
	}

	log.Trace().Str("function", f.Name.Name).Int("params", numParams).Int("results", numResults).Send()

	if numResults == 0 {
		// Can be a vannilla handler. Need to check the param for *gin.Context
		if numParams == 1 {
			fui.ParamsOpening = f.Type.Params.Opening
			p := f.Type.Params.List[0]
			fi := parseField(fs, p, iResolver)

			// Should parseFuncDecl based on gin.Context... At the moment parseFuncDecl *gin.Context
			if fi.QType == "github.com/gin-gonic/gin/Context" {
				fui.Params = []FieldInfo{fi}
				fui.FType = GIN_SimpleHandler
			} else {
				log.Trace().Str("function", f.Name.Name).Str("param", fi.String()).Msg("param not of *gin.Context")
			}

		} else {
			log.Trace().Str("function", f.Name.Name).Msg("function is not a gin handler")
		}
	}

	if numResults == 1 {
		// Can be a wrapper..
		p := f.Type.Results.List[0]
		switch tp := p.Type.(type) {
		case *ast.FuncType:
			if tp.Params != nil && len(tp.Params.List) == 1 && tp.Results == nil {
				fi1 := parseField(fs, tp.Params.List[0], iResolver)
				if fi1.QType == "github.com/gin-gonic/gin/Context" {
					fui.FType = GIN_Closure_Func
				}
			}

		default:
			fi := parseField(fs, p, iResolver)
			if fi.QType == "github.com/gin-gonic/gin/HandlerFunc" || fi.QType == "H" {
				fui.Result = fi
				fui.FType = GIN_Closure_HandlerFunc
			}
		}

		if numParams > 0 {
			/*
			 * Should try to match comment annotations.
			 */
			pStart := f.Type.Params.Opening
			fui.ParamsOpening = f.Type.Params.Opening
			for _, p := range f.Type.Params.List {
				fi := parseField(fs, p, iResolver)
				cg := findFieldComment(pStart, fi.Pos, comments)
				if cg != nil {
					a, err := parseComments(cg)
					if err != nil {
						log.Error().Str("function", f.Name.Name).Str("field", fi.Name).Err(err).Msg("error in parsing comments")
						return fui, nil
					}

					fi.Annotations = a
				}
				fui.Params = append(fui.Params, fi)

				// Increment the starting pos to next param.
				pStart = fi.End
			}
		}
	}

	// More than one result: function filtered.
	return fui, nil
}

func findFieldComment(paramsStartPos token.Pos, paramNameStartPos token.Pos, cmms []*ast.CommentGroup) *ast.CommentGroup {

	for _, cg := range cmms {
		if cg.Pos() > paramsStartPos && cg.Pos() < paramNameStartPos {
			return cg
		}
	}

	return nil
}

func parseField(fs *token.FileSet, field *ast.Field, resolver *importResolver) FieldInfo {

	fi := FieldInfo{}
	if len(field.Names) > 0 {
		fi.Name = field.Names[0].Name
	}

	fi.FilePos = fs.Position(field.Pos())
	fi.Pos = field.Pos()
	fi.End = field.End()
	switch tt := field.Type.(type) {
	case *ast.StarExpr:
		fi.Pointer = true
		switch ttt := tt.X.(type) {
		case *ast.SelectorExpr:
			fi.Sel = ttt.Sel.Name
			fi.X = ttt.X.(*ast.Ident).Name
		case *ast.Ident:
			fi.X = ttt.Name
		}
	case *ast.SelectorExpr:
		fi.Sel = tt.Sel.Name
		fi.X = tt.X.(*ast.Ident).Name
	case *ast.Ident:
		fi.X = tt.Name
	default:
		panic(fmt.Sprintf("unrecognized fieldInfo %#v", field.Type))
	}

	if fi.Sel != "" {
		i := resolver.Resolve(fi.X)
		if i == nil {
			panic(fmt.Sprintf("Cannot find import for %s", fi.X))
		}

		fi.QType = i.path + "/" + fi.Sel
	} else {
		fi.QType = fi.X
	}

	return fi
}

func sortByPathPackage(fds []FileDef) {
	sort.Slice(fds, func(i, j int) bool {
		return fds[i].FileFolderAndPackageLessThan(&fds[j])
	})
}
