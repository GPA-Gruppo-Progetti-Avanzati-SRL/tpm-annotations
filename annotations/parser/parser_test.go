package parser_test

import (
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-annotations/annotations/parser"
	"testing"
)

func Test_Parser(t *testing.T) {

	ang, err := parser.Parse("/*\n @Get\n@Path(\"/api/v1\") @DocxInclude(src=\"file.xml\") @PUT @HEAD @DELETE @PATCH @POST testo qualsiasi \n*/")
	if err != nil {
		t.Fatal(err)
	}

	for _, a := range ang {
		fmt.Printf("%#v\n", a)
	}

}
