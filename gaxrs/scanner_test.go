package gaxrs_test

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-annotations/annotations/scanner"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-annotations/gaxrs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"testing"
)

type resultDetect struct {
}

func (r *resultDetect) ResultType(isFun bool, fis ...scanner.FieldInfo) scanner.FuncType {

	var rt scanner.FuncType
	rt = scanner.UNK_RESULT

	if isFun {
		if len(fis) == 1 {
			qtype := fis[0].QType
			if qtype == "github.com/gin-gonic/gin/Context" {
				rt = gaxrs.GIN_Closure_Func
			}
		}
	} else {
		qtype := fis[0].QType
		if qtype == "github.com/gin-gonic/gin/HandlerFunc" || qtype == "H" {
			rt = gaxrs.GIN_Closure_HandlerFunc
		}
	}

	return rt
}

func Test_Scanner(t *testing.T) {

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	funcResolver := resultDetect{}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Trace().Str("cwd", dir).Send()

	log.Info().Msgf("%s %s", strings.Repeat("=", 20), "GAXRS: Scanning Stage")
	fds, err := scanner.Scan([]scanner.ScanSource{
		{FileName: "../test-data/example-go.txt"},
		{FileName: "../test-data/example-empty-go.txt"},
		{FileName: "../test-data/example-nonvalidating-go.txt"},
		{FileName: "../test-data/example-import-alias-go.txt"},
	}, scanner.WithResultResolver(&funcResolver))
	if err != nil {
		log.Fatal().Err(err).Msg("scan terminated with error")
	}

	gs, errs := gaxrs.BuildModel(fds)
	if len(errs) > 0 {
		log.Fatal().Err(err).Msg("validation failure of scanning stage")
	}

	log.Info().Msgf("%s %s", strings.Repeat("=", 20), "GAXRS: Generation Model")
	for _, g := range gs {
		log.Info().Str("name", g.Name).Str("filename", g.FileName).Str("package", g.Package).Msg("group")
		for _, r := range g.Resources {
			log.Info().Str("name", r.Name).Str("httpMethod", r.HttpMethod).Str("func", r.Method).Bool("closure", r.Closure).Int("numberOfParams", len(r.Params)).Msg("resource")
			for _, rp := range r.Params {
				log.Info().Str("name", rp.Name).Str("contextName", rp.CtxName).Str("qualifiedType", rp.QType).Bool("pointer", rp.IsStar).Msg("param")
			}
		}
	}

	/*
		if !gaxrs.Validate(fds) {
			log.Fatal().Err(err).Msg("validation failure of scanning stage")
		}
	*/

	gaxrs.Generate(gs)
}
