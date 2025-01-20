package lexer_test

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-annotations/annotations/lexer"
	"testing"
)

type TokenVal struct {
	tokType lexer.TokenType
	val     string
}

func Test_Annotations(t *testing.T) {
	cases := []TokenVal{
		{lexer.AnnotationToken, "@Get"},
		{lexer.AnnotationToken, "@Path"},
		{lexer.LParenToken, "("},
		{lexer.StringToken, "\"/api/v1\""},
		{lexer.RParenToken, ")"},
	}

	processTestCase(t, "// @Get\n@Path (\"/api/v1\") whatever", cases)

	cases = []TokenVal{
		{lexer.AnnotationToken, "@Path"},
		{lexer.LParenToken, "("},
		{lexer.IdentifierToken, "param"},
		{lexer.EqualToken, "="},
		{lexer.StringToken, "\"/api/v1\""},
		{lexer.CommaToken, ","},
		{lexer.StringToken, "\"ciao\""},
		{lexer.RParenToken, ")"},
	}

	processTestCase(t, "// @Path(param = \"/api/v1\", \"ciao\" )\n// Deprecated:", cases)

	cases = []TokenVal{
		{lexer.AnnotationToken, "@DocxInclude"},
		{lexer.LParenToken, "("},
		{lexer.IdentifierToken, "src"},
		{lexer.EqualToken, "="},
		{lexer.StringToken, "\"fragment.xml\""},
		{lexer.RParenToken, ")"},
	}

	processTestCase(t, `// @DocxInclude (src="fragment.xml") whatever`, cases)
}

func processTestCase(t *testing.T, source string, cases []TokenVal) {
	l := lexer.New(source, lexer.TextState)
	l.Start()

	for _, c := range cases {
		tok, done := l.NextToken()

		if tok.Type == lexer.ErrorToken {
			t.Error(tok.Value)
			return
		}

		if tok != nil {
			t.Logf("[%d] - %s", tok.Type, tok.Value)
		}

		if done {
			t.Error("Expected there to be more tokens, but there weren't")
			return
		}

		if c.tokType != tok.Type {
			t.Errorf("Expected token type %v but got %v", c.tokType, tok.Type)
			return
		}

		if c.val != tok.Value {
			t.Errorf("Expected %q but got %q", c.val, tok.Value)
			return
		}
	}

	tok, done := l.NextToken()
	if !done {
		t.Error("Expected the lexer to be done, but it wasn't.")
		return
	}

	if tok != nil {
		t.Errorf("Did not expect a token, but got %v", *tok)
		return
	}
}
