package scanner_test

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-annotations/annotations/scanner"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"testing"
)

const (
	UNK_RESULT              scanner.FuncType = "UNK_RESULT"
	GIN_SimpleHandler       scanner.FuncType = "gin-handler"
	GIN_Closure_HandlerFunc scanner.FuncType = "gin-closure-handlerfunc"
	GIN_Closure_Func        scanner.FuncType = "gin-closure-handler"
)

type resultDetect struct {
}

func (r *resultDetect) ResultType(isFun bool, fis ...scanner.FieldInfo) scanner.FuncType {

	var rt scanner.FuncType
	rt = UNK_RESULT

	if isFun {
		if len(fis) == 1 {
			qtype := fis[0].QType
			if qtype == "github.com/gin-gonic/gin/Context" {
				rt = GIN_Closure_Func
			}
		}
	} else {
		qtype := fis[0].QType
		if qtype == "github.com/gin-gonic/gin/HandlerFunc" || qtype == "H" {
			rt = GIN_Closure_HandlerFunc
		}
	}

	return rt
}

func Test_Scanner(t *testing.T) {

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Trace().Str("cwd", dir).Send()

	funcResolver := resultDetect{}

	fds, err := scanner.Scan([]scanner.ScanSource{
		{FileName: "../../test-data/example-go.txt"},
		{FileName: "../../test-data/example-empty-go.txt"},
		{FileName: "../../test-data/example-nonvalidating-go.txt"},
	}, scanner.WithResultResolver(&funcResolver))
	if err != nil {
		log.Fatal().Err(err).Msg("scan terminated with error")
	}

	groupFds := make([]scanner.FileDef, 0)
	for _, f := range fds {
		if len(groupFds) == 0 {
			groupFds = append(groupFds, f)
		} else {
			if !f.AreFolderAndPackageCompatibile(&groupFds[0]) {
				log.Error().
					Str("pkg1", f.Package).Str("pkg2", groupFds[0].Package).
					Str("f1", f.Name).Str("f2", groupFds[0].Name).
					Msg("folder and package are not compatible. files in same folder need to have same package.")
			} else {
				if f.SameGroup(&groupFds[0]) {
					groupFds = append(groupFds, f)
				} else {
					// Process group and clear
					log.Info().Msg("start generating gaxrs declarations")
					for _, f := range groupFds {
						log.Trace().Str("filename", f.Name).Str("package", f.Package).Send()
					}
					groupFds = make([]scanner.FileDef, 0)
					// TODO: MAI should append to group...
					groupFds = append(groupFds, f)
				}
			}
		}
	}

	if len(groupFds) > 0 {
		// Process group
		log.Info().Msg("start generating gaxrs declarations")
		for _, f := range groupFds {
			log.Trace().Str("filename", f.Name).Str("package", f.Package).Send()
		}
	}
}
