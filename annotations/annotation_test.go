package annotations_test

import (
	"testing"
	"tpm-annotations/annotations"
)

func Test_Annotations(t *testing.T) {

	if a, e := annotations.NewAnnotation("@GET", "", nil); e != nil {
		t.Error(e)
	} else {
		t.Logf("%#v", a)
	}

}
