package jaxrs

import (
	"encoding/json"
	"fmt"
	"go/token"
	"path/filepath"
	"strings"
	"tpm-annotations/annotations"
)

type FileDef struct {
	Name        string
	Package     string
	Annotations annotations.AnnotationGroup
	imports     importResolver
	Methods     []FuncInfo
}

func (fd *FileDef) GetFileFolder() string {
	return filepath.Dir(fd.Name)
}

func (fd *FileDef) FileFolderAndPackageLessThan(another *FileDef) bool {
	rc := false
	switch strings.Compare(fd.GetFileFolder(), another.GetFileFolder()) {
	case 0:
		if strings.Compare(fd.Package, another.Package) < 0 {
			rc = true
		}
	case -1:
		rc = true
	default:
		rc = false
	}

	return rc
}

func (fd *FileDef) FileFolderAndPackageEqualTo(another *FileDef) bool {
	return strings.Compare(fd.GetFileFolder(), another.GetFileFolder()) == 0
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

type FuncType string

const (
	NO_HANDLER              = "no-handler"
	GIN_SimpleHandler       = "gin-handler"
	GIN_Closure_HandlerFunc = "gin-closure-handlerfunc"
	GIN_Closure_Func        = "gin-closure-handler"
)

type FuncInfo struct {
	FType FuncType
	Pos   token.Position
	Name  string

	Params        []FieldInfo
	ParamsOpening token.Pos
	Result        FieldInfo
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
