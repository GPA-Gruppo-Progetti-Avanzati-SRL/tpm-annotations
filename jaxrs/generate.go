package jaxrs

import (
	"errors"
	"github.com/rs/zerolog/log"
)

type Config struct {
	files []string
}

type Option func(cfg *Config)

func Process(opts ...Option) error {

	cfg := Config{}
	for _, o := range opts {
		o(&cfg)
	}

	ss := make([]ScanSource, 0)
	if len(cfg.files) > 0 {
		for _, f := range cfg.files {
			ss = append(ss, ScanSource{FileName: f})
		}
	}

	if len(ss) == 0 {
		err := errors.New("no files selected for processing")
		log.Error().Err(err).Send()
		return err
	}

	fds, err := Scan(ss)
	if err != nil {
		log.Error().Err(err).Msg("scan terminated with error")
		return err
	}

	if !Validate(fds) {
		log.Error().Err(err).Msg("validation failure of scanning stage")
		return err
	}

	return nil
}

func WithFiles(files []string) Option {
	return func(cfg *Config) {
		cfg.files = files
	}
}
