package annotations

import (
	"net/http"
	"strings"
)

func IsHttpMethod(a Annotation) bool {
	s := strings.ToLower(a.GetName())
	return s == "@get" || s == "@put" || s == "@post" || s == "@delete" || s == "@head" || s == "@patch"
}

// HttpGetMethod
type HttpGetMethod struct {
	annotationImpl `mapstructure:",squash"`
}

func (a HttpGetMethod) GetValue() string {
	return http.MethodGet
}

func (a HttpGetMethod) IsValid() bool {
	return a.Value == ""
}

// HttpPostMethod
type HttpPostMethod struct {
	annotationImpl `mapstructure:",squash"`
}

func (a HttpPostMethod) GetValue() string {
	return http.MethodPost
}

func (a HttpPostMethod) IsValid() bool {
	return a.Value == ""
}

// HttpPutMethod
type HttpPutMethod struct {
	annotationImpl `mapstructure:",squash"`
}

func (a HttpPutMethod) GetValue() string {
	return http.MethodPut
}

func (a HttpPutMethod) IsValid() bool {
	return a.Value == ""
}

// HttpDeleteMethod
type HttpDeleteMethod struct {
	annotationImpl `mapstructure:",squash"`
}

func (a HttpDeleteMethod) GetValue() string {
	return http.MethodDelete
}

func (a HttpDeleteMethod) IsValid() bool {
	return a.Value == ""
}

// HttpHeadMethod
type HttpHeadMethod struct {
	annotationImpl `mapstructure:",squash"`
}

func (a HttpHeadMethod) GetValue() string {
	return http.MethodHead
}

func (a HttpHeadMethod) IsValid() bool {
	return a.Value == ""
}

// HttpPatchMethod
type HttpPatchMethod struct {
	annotationImpl `mapstructure:",squash"`
}

func (a HttpPatchMethod) GetValue() string {
	return http.MethodPatch
}

func (a HttpPatchMethod) IsValid() bool {
	return a.Value == ""
}

// HttpPath
type HttpPath struct {
	annotationImpl `mapstructure:",squash"`
}

func (a HttpPath) IsValid() bool {
	return a.Value != ""
}

// HttpDeleteMethod
type MethodParam struct {
	annotationImpl `mapstructure:",squash"`
}

func (a MethodParam) IsValid() bool {
	return a.Value != ""
}
