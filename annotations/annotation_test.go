package annotations_test

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-annotations/annotations"
	"testing"
)

func Test_Annotations(t *testing.T) {

	if a, e := annotations.NewAnnotation("@GET", "", nil); e != nil {
		t.Error(e)
	} else {
		t.Logf("%#v", a)
	}

}
