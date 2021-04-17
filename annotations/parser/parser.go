package parser

import (
	"errors"
	"fmt"
	"strings"
	"tpm-annotations/annotations"
	"tpm-annotations/annotations/lexer"
)

func Parse(s string) (annotations.AnnotationGroup, error) {

	ang := make([]annotations.Annotation, 0)

	l := lexer.New(s, lexer.TextState)
	l.Start()

	tok, done := l.NextToken()
	for !done {
		switch tok.Type {
		case lexer.AnnotationToken:
			a, t, e := newAnnotation(tok.Value, l)
			if e != nil {
				return nil, e
			}
			ang = append(ang, a)
			if t != nil {
				tok = t
			} else {
				tok, done = l.NextToken()
			}

		case lexer.ErrorToken:
			return nil, errors.New(tok.Value)

		default:
			return nil, errors.New(tok.Value)
		}
	}

	return ang, nil
}

const (
	ParsedStart = iota
	ParsedLParen
	ParsedStringValue
	ParsedParam
	ParsedIdentifier
	ParsedEqual
	ParsedComma
)

func newAnnotation(n string, l *lexer.L) (annotations.Annotation, *lexer.Token, error) {
	v := ""

	state := ParsedStart
	for {
		t, done := l.NextToken()
		if done {
			if state == ParsedStart {
				a, e := annotations.NewAnnotation(n, v, nil)
				return a, nil, e
			} else {
				return nil, nil, errors.New("unexpected eof of source")
			}
		}

		if t.Type == lexer.ErrorToken {
			return nil, nil, errors.New(t.Value)
		}

		switch t.Type {
		case lexer.LParenToken:
			if state == ParsedStart {
				state = ParsedLParen
			} else {
				return nil, nil, errors.New(fmt.Sprintf("found '(' when state is %d", state))
			}
		case lexer.RParenToken:
			if state == ParsedLParen || state == ParsedStringValue || state == ParsedParam {
				a, e := annotations.NewAnnotation(n, v, nil)
				return a, nil, e
			} else {
				return nil, nil, errors.New(fmt.Sprintf("found ')' when state is %d", state))
			}
		case lexer.StringToken:
			switch state {
			case ParsedLParen:
				v = strings.TrimPrefix(strings.TrimSuffix(t.Value, "\""), "\"")
				state = ParsedStringValue
			case ParsedEqual:
				state = ParsedParam
			default:
				return nil, nil, errors.New(fmt.Sprintf("found %s when state is %d", t.Value, state))
			}
		case lexer.CommaToken:
			if state == ParsedParam {
				state = ParsedComma
			} else {
				return nil, nil, errors.New(fmt.Sprintf("found %s when state is %d", t.Value, state))
			}
		case lexer.EqualToken:
			if state == ParsedIdentifier {
				state = ParsedEqual
			} else {
				return nil, nil, errors.New(fmt.Sprintf("found %s when state is %d", t.Value, state))
			}
		case lexer.IdentifierToken:
			if state == ParsedComma || state == ParsedLParen {
				state = ParsedIdentifier
			} else {
				return nil, nil, errors.New(fmt.Sprintf("found %s when state is %d", t.Value, state))
			}
		case lexer.AnnotationToken:
			if state == ParsedStart {
				a, e := annotations.NewAnnotation(n, v, nil)
				return a, t, e
			}
		}
	}

}
