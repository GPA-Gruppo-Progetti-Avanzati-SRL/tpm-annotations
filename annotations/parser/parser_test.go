package parser_test

import (
	"fmt"
	"testing"
	"tpm-annotations/annotations/parser"
)

func Test_Parser(t *testing.T) {

	ang, err := parser.Parse("/*\n @Get\n@Path(\"/api/v1\") @PUT @HEAD @DELETE @PATCH @POST testo qualsiasi \n*/")
	if err != nil {
		t.Fatal(err)
	}

	for _, a := range ang {
		fmt.Printf("%#v\n", a)
	}

}
