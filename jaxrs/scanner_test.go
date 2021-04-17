package jaxrs_test

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"testing"
	"tpm-annotations/jaxrs"
)

func Test_Scanner(t *testing.T) {

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Trace().Str("cwd", dir).Send()

	fds, err := jaxrs.Scan([]jaxrs.ScanSource{
		{FileName: "../test-data/example-go.txt"},
		{FileName: "../test-data/example-empty-go.txt"},
		//{FileName: "../test-data/example-nonvalidating-go.txt" },
	})
	if err != nil {
		log.Fatal().Err(err).Msg("scan terminated with error")
	}

	if !jaxrs.Validate(fds) {
		log.Fatal().Err(err).Msg("validation failure of scanning stage")
	}

	groupFds := make([]jaxrs.FileDef, 0)
	for _, f := range fds {
		if len(groupFds) == 0 {
			groupFds = append(groupFds, f)
		} else {
			if f.FileFolderAndPackageEqualTo(&groupFds[0]) {
				groupFds = append(groupFds, f)
			} else {
				// Process group and clear
				log.Info().Msg("start generating jaxrs declarations")
				for _, f := range groupFds {
					log.Trace().Str("filename", f.Name).Str("package", f.Package).Send()
				}
				groupFds = make([]jaxrs.FileDef, 0)
			}
		}
	}

	if len(groupFds) > 0 {
		// Process group
		log.Info().Msg("start generating jaxrs declarations")
		for _, f := range groupFds {
			log.Trace().Str("filename", f.Name).Str("package", f.Package).Send()
		}
	}
}
