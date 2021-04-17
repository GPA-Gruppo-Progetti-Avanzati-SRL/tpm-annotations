package lexer

func TextState(l *L) StateFunc {

	r := l.Next()
	for r != EOFRune && r != '@' {
		r = l.Next()
	}

	if r == EOFRune {
		return nil
	}

	l.Rewind()
	l.Ignore()
	return AnnotationState
}

func AnnotationState(l *L) StateFunc {

	// Consumer the rwinded '@'
	r := l.Next()

	r = l.Next()
	for (r >= 'a' && r <= 'z') || r == '_' || (r >= 'A' && r <= 'Z') {
		r = l.Next()
	}

	l.Rewind()

	// Single '@' are ignored.
	if len(l.Current()) == 1 {
		l.Ignore()
		return TextState
	}

	l.Emit(AnnotationToken)

	l.Take(" \t\n\r")
	l.Ignore()
	r = l.Next()
	if r == '(' {
		l.Emit(LParenToken)
		return AnnotationStateBody
	} else {
		l.Rewind()
	}

	return TextState
}

func AnnotationStateBody(l *L) StateFunc {

	l.Take(" \t\n\r")
	l.Ignore()

	r := l.Next()
	switch sr := r; {
	case sr == '=':
		l.Emit(EqualToken)
		return AnnotationStateBody
	case sr == ',':
		l.Emit(CommaToken)
		return AnnotationStateBody
	case sr == ')':
		l.Emit(RParenToken)
		return TextState
	case sr == '"':
		return StringState
	case (sr >= 'a' && sr <= 'z') || sr == '_' || (sr >= 'A' && sr <= 'Z'):
		return AnnotationStateParamIdentifier
	default:
		return l.Errorf(ErrorToken, "unexpected token %q", r)
	}

	return nil
}

func StringState(l *L) StateFunc {

	r := l.Next()
	for r != EOFRune && r != '"' {
		if r == '\\' {
			r = l.Next()
			if r == '"' {
				r = l.Next()
			}
		}
		r = l.Next()
	}

	if r == EOFRune {
		return l.Errorf(ErrorToken, "string not properly terminated")
	}

	l.Emit(StringToken)
	return AnnotationStateBody
}

func AnnotationStateParamIdentifier(l *L) StateFunc {
	r := l.Next()
	for (r >= 'a' && r <= 'z') || r == '_' || (r >= 'A' && r <= 'Z') {
		r = l.Next()
	}

	l.Rewind()
	l.Emit(IdentifierToken)
	return AnnotationStateBody
}
