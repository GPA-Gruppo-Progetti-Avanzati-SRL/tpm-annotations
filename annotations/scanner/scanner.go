package scanner

import (
	"encoding/json"
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-annotations/annotations"
	parser2 "github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-annotations/annotations/parser"
	"github.com/rs/zerolog/log"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"strings"
)

type FileDef struct {
	Name        string
	Package     string
	Annotations annotations.AnnotationGroup
	imports     importResolver
	Methods     []FuncInfo
}

func (d FileDef) SameGroup(s *FileDef) bool {
	return path.Dir(d.Name) == path.Dir(s.Name)
}

func (d FileDef) AreFolderAndPackageCompatibile(s *FileDef) bool {
	if d.Package == s.Package || path.Dir(d.Name) != path.Dir(s.Name) {
		return true
	}

	return false
}

type importResolver struct {
	imports []importRef
}

type importRef struct {
	path  string
	alias string
}

func (ir *importResolver) Resolve(pkg string) *importRef {
	foundIndex := -1
	for ix, i := range ir.imports {
		if i.alias == pkg {
			return &i
		} else {
			n := strings.LastIndex(i.path, "/")
			if n >= 0 && i.path[n+1:] == pkg {
				foundIndex = ix
			}
		}
	}

	if foundIndex >= 0 {
		return &ir.imports[foundIndex]
	}

	return nil
}

const NO_RESULT FuncType = "NO_RESULT"
const UNK_RESULT FuncType = "UNK_RESULT"

type FuncType string

type FuncTypeResolver interface {
	ResultType(isFun bool, fis ...FieldInfo) FuncType
}

type FuncInfo struct {
	Pos  token.Position
	Name string

	Params        []FieldInfo
	ParamsOpening token.Pos
	ResultType    FuncType
	Receiver      *FieldInfo
	Annotations   annotations.AnnotationGroup
}

func (fui FuncInfo) ToJsonString() string {
	if b, e := json.Marshal(&fui); e != nil {
		return "unmarshal error for function " + fui.Name
	} else {
		return string(b)
	}
}

func (fui FuncInfo) String() string {
	return fmt.Sprintf("%s at %d:%d", fui.Name, fui.Pos.Line, fui.Pos.Column)
}

type FieldInfo struct {
	Pos     token.Pos
	End     token.Pos
	FilePos token.Position
	Name    string
	Pointer bool
	X       string
	Sel     string
	QType   string

	Annotations annotations.AnnotationGroup
}

func (fi *FieldInfo) String() string {
	var stb strings.Builder

	stb.WriteString(fi.Name)
	if fi.Pointer {
		stb.WriteString(" *")
	} else {
		stb.WriteString(" ")
	}

	addDot := false
	if fi.X != "" {
		stb.WriteString(fi.X)
		addDot = true
	}

	if fi.Sel != "" {
		if addDot {
			stb.WriteRune('.')
		}

		stb.WriteString(fi.Sel)
	}

	return stb.String()
}

type ScanSource struct {
	FileName string
	Source   string
}

type ScanConfig struct {
	fTypeResolver FuncTypeResolver
}

type ScanOption func(c *ScanConfig)

func WithResultResolver(r FuncTypeResolver) ScanOption {
	return func(c *ScanConfig) {
		c.fTypeResolver = r
	}
}

/*
 *
 */
func Scan(files []ScanSource, opts ...ScanOption) ([]FileDef, error) {

	cfg := ScanConfig{}
	for _, o := range opts {
		o(&cfg)
	}

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
				fui, err := parseFuncDecl(fs, tdecl, f.Comments, &fd.imports, cfg.fTypeResolver)
				if err != nil {
					return nil, err
				} else {
					if len(fui.Annotations) > 0 {
						log.Info().Str("function", fui.Name).
							Str("at", fui.Pos.String()).
							Int("numberOfParams", len(fui.Params)).
							Interface("resultType", fui.ResultType).
							Interface("receiver", fui.Receiver).
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

	return fds, nil
}

func parseComments(comments *ast.CommentGroup) ([]annotations.Annotation, error) {

	if comments == nil {
		return nil, nil
	}

	var as = make([]annotations.Annotation, 0)
	for _, d := range comments.List {
		a, err := parser2.Parse(d.Text)
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

func parseFuncDecl(fs *token.FileSet, f *ast.FuncDecl, comments []*ast.CommentGroup, iResolver *importResolver, fResolver FuncTypeResolver) (FuncInfo, error) {

	fui := FuncInfo{}
	fui.ResultType = NO_RESULT
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
		recvf := parseField(fs, f.Recv.List[0], iResolver)
		fui.Receiver = &recvf
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

	if fResolver != nil && numResults > 0 {

		isFun := false

		resultFieldList := make([]FieldInfo, 0)
		// Check the first one. If itt is a function parse function params... otherwise parse params.
		// Can be a wrapper..
		if tp, ok := f.Type.Results.List[0].Type.(*ast.FuncType); ok {
			isFun = true
			if tp.Params != nil {
				for _, p := range tp.Params.List {
					fi := parseField(fs, p, iResolver)
					resultFieldList = append(resultFieldList, fi)
				}
			}
		} else {
			for _, rs := range f.Type.Results.List {
				fi := parseField(fs, rs, iResolver)
				resultFieldList = append(resultFieldList, fi)
			}
		}

		fui.ResultType = fResolver.ResultType(isFun, resultFieldList...)
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
	case *ast.FuncType:
		panic(fmt.Sprintf("func type paremeter not supported %#v", field.Type))
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
