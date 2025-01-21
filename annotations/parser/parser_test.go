package parser_test

import (
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-annotations/annotations"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-annotations/annotations/parser"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

type annotationImpl struct {
	Name  string
	Value string
}

func (a annotationImpl) GetName() string {
	return a.Name
}

func (a annotationImpl) GetValue() string {
	return a.Value
}

type DocxInclude struct {
	annotationImpl `mapstructure:",squash"`
	Src            string `json:"src,omitempty" yaml:"src,omitempty" mapstructure:"src,omitempty"`
}

func (a DocxInclude) IsValid() bool {
	return a.Value != "" || a.Src != ""
}

func Test_Parser(t *testing.T) {

	err := annotations.Register("@DocxInclude", reflect.TypeOf(DocxInclude{}))
	require.NoError(t, err)

	cases := []string{
		"/*\n @Get\n@Path(\"/api/v1\") @DocxInclude(src=\"file.xml\") @DocxInclude(\"file.xml\") @PUT @HEAD @DELETE @PATCH @POST testo qualsiasi \n*/",
		`<!-- 
@DocxInclude(src="file.xml") -->`,
	}

	for _, c := range cases {
		ang, err := parser.Parse(c)
		require.NoError(t, err)
		for _, a := range ang {
			fmt.Printf("%#v\n", a)
		}
	}

}
