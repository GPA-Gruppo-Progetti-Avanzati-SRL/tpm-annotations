package annotations

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strings"
)

func init() {
	registry = make(map[string]reflect.Type)
	registry["@get"] = reflect.TypeOf(HttpGetMethod{})
	registry["@put"] = reflect.TypeOf(HttpPutMethod{})
	registry["@post"] = reflect.TypeOf(HttpPostMethod{})
	registry["@delete"] = reflect.TypeOf(HttpDeleteMethod{})
	registry["@head"] = reflect.TypeOf(HttpHeadMethod{})
	registry["@patch"] = reflect.TypeOf(HttpPatchMethod{})
	registry["@path"] = reflect.TypeOf(HttpPath{})
	registry["@param"] = reflect.TypeOf(MethodParam{})
}

var registry map[string]reflect.Type

func Register(n string, t reflect.Type) error {

	if _, ok := reflect.New(t).Elem().Interface().(Annotation); !ok {
		return errors.New(fmt.Sprintf("non annotation structure registered for %s annotation", n))
	}

	if registry == nil {
		registry = make(map[string]reflect.Type)
	}

	n = strings.ToLower(n)
	if _, ok := registry[n]; !ok {
		registry[n] = t
	}

	return nil
}

type AnnotationGroup []Annotation

type Annotation interface {
	GetName() string
	GetValue() string
	IsValid() bool
}

// annotationImpl: generic implementation
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

func NewAnnotation(n string, v string, params map[string]interface{}) (Annotation, error) {

	if registry == nil {
		return nil, errors.New(fmt.Sprintf("the requested annotation cannot be found in registry: %s", n))
	}

	if t, ok := registry[strings.ToLower(n)]; !ok {
		return nil, errors.New(fmt.Sprintf("the requested annotation cannot be found in registry: %s", n))
	} else {
		var result = reflect.New(t)
		input := make(map[string]interface{})
		input["Name"] = n
		if v != "" {
			input["Value"] = v
		}
		if params != nil {
			input["params"] = params
		}

		err := mapstructure.Decode(input, result.Interface())
		if err != nil {
			return nil, err
		}

		a := result.Elem().Interface().(Annotation)
		if !a.IsValid() {
			return nil, errors.New(fmt.Sprintf("annotation %s is not valid (check properties and provided params)", n))
		}

		return a, nil
	}

}

// Validation related methods
func (ang AnnotationGroup) GetFirstIn(n ...string) Annotation {
	for _, a := range ang {
		m := strings.ToLower(a.GetName())
		for _, s := range n {
			if strings.ToLower(s) == m {
				return a
			}
		}
	}

	return nil
}

func (ang AnnotationGroup) Accept(n ...string) error {
	for _, a := range ang {
		found := false
		m := strings.ToLower(a.GetName())
		for _, s := range n {
			if strings.ToLower(s) == m {
				found = true
				break
			}
		}

		if !found {
			return errors.New(fmt.Sprintf("annotation %s cannot be found", a.GetName()))
		}
	}

	return nil
}

func (ang AnnotationGroup) NoDuplicates() error {

	var err error

	found := make(map[string]struct{})
	for _, a := range ang {
		aid := strings.ToLower(a.GetName())
		if _, ok := found[aid]; !ok {
			found[aid] = struct{}{}
		} else {
			return errors.New(fmt.Sprintf("%s has been dupped", a.GetName()))
		}
	}

	return err
}

/*
func (ang AnnotationGroup) ZeroOrOneOf(acceptedAnns ...string) error {

	var err error

	switch len(ang) {
	case 0:

	case 1:
		if len(acceptedAnns) > 0 {

			found := false
			a := strings.ToLower(ang[0].GetName())
			for _, s := range acceptedAnns {
				if strings.ToLower(s) == a {
					found = true
					break
				}
			}

			if !found {
				err = errors.New("accepted annotations not found")
			}
		}
	default:
		err = errors.New("too many annotations: 0 or 1 annotations accepted")
	}

	return err
}
*/

func (ang AnnotationGroup) MustHaveExactlyOneOutOf(ans ...string) error {

	var err error

	found := ""
	for _, a := range ang {
		aid := strings.ToLower(a.GetName())
		for _, acc := range ans {
			if aid == strings.ToLower(acc) {
				if found == "" {
					found = a.GetName()
				} else {
					return errors.New(fmt.Sprintf("%s annotation conflicts with %s... only one accepted", found, acc))
				}
			}
		}
	}

	return err
}

/*
func (ang AnnotationGroup) ShouldHaveAtMostOneOf(acceptedAnns ...string) error {

	var err error

	found := make(map[string]struct{})
	for _, a := range ang {
		aid := strings.ToLower(a.GetName())
		for _, acc := range acceptedAnns {
			if aid == strings.ToLower(acc) {
				if _, ok := found[aid]; !ok {
					found[aid] = struct{}{}
				} else {
					return errors.New("annotations duplicates")
				}
			}
		}
	}

	return err
}

*/
