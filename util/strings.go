package util

import (
	"regexp"
	"strings"
)

// Code from https://github.com/angular/angular-cli/blob/master/packages/angular_devkit/core/src/utils/strings.ts

const (
	STRING_DASHERIZE_REGEXP    = `[ _]`
	STRING_DECAMELIZE_REGEXP   = `([a-z\d])([A-Z])`
	STRING_CAMELIZE_REGEXP     = `(-|_|\.|\s)+(.)?`
	STRING_CAMELIZE_REGEXP_2   = `^([A-Z])`
	STRING_UNDERSCORE_REGEXP_1 = `([a-z\d])([A-Z]+)`
	STRING_UNDERSCORE_REGEXP_2 = `-|\s+`
)

func Decamelize(s string) string {
	m := regexp.MustCompile(STRING_DECAMELIZE_REGEXP)
	// fmt.Printf("%q\n", m.FindAllString(s, -1))
	return strings.ToLower(m.ReplaceAllString(s, "${1}_${2}"))
}

func Dasherize(s string) string {
	m := regexp.MustCompile(STRING_DASHERIZE_REGEXP)
	return m.ReplaceAllString(Decamelize(s), "-")
}

func Camelize(s string) string {
	m := regexp.MustCompile(STRING_CAMELIZE_REGEXP)
	s1 := m.ReplaceAllStringFunc(s, func(r string) string { return strings.ToUpper(r[len(r)-1:]) })

	m1 := regexp.MustCompile(STRING_CAMELIZE_REGEXP_2)
	return m1.ReplaceAllStringFunc(s1, func(r string) string { return strings.ToLower(r) })
}

func Classify(s string) string {
	sarr := strings.Split(s, ".")
	for i := 0; i < len(sarr); i++ {
		sarr[i] = Capitalize(Camelize(sarr[i]))
	}

	return strings.Join(sarr, ".")
}

func Underscore(s string) string {
	m := regexp.MustCompile(STRING_UNDERSCORE_REGEXP_1)
	s1 := m.ReplaceAllString(s, "${1}_${2}")

	m2 := regexp.MustCompile(STRING_UNDERSCORE_REGEXP_2)
	s2 := m2.ReplaceAllString(s1, "_")
	return strings.ToLower(s2)
}

func Capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}
