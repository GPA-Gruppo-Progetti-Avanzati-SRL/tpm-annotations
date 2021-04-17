package jaxrs

import (
	"github.com/rs/zerolog/log"
	"strings"
)

func Validate(fds []FileDef) bool {

	rc := true
	for _, fd := range fds {

		httpMethodsMissingPathAnnotation := make(map[string]struct{})
		groupPath := false
		if err := fd.Annotations.Accept("@Path"); err != nil {
			rc = false
			log.Error().Str("allowed", "@Path").Str("ctx", "file").Err(err).Msg("not allowed annotation present")
		} else {
			if err := fd.Annotations.NoDuplicates(); err != nil {
				log.Error().Str("ctx", "file").Err(err).Msg("duplicates present")
				rc = false
			} else {
				if fd.Annotations.GetFirstIn("@path") != nil {
					groupPath = true
				}
			}
		}

		for _, m := range fd.Methods {

			if err := m.Annotations.Accept("@Path", "@GET", "@PUT", "@POST", "@PATCH", "@HEAD", "@DELETE"); err != nil {
				rc = false
				log.Error().Str("ctx", "function").Str("function", m.Name).Str("at", m.Pos.String()).Err(err).Msg("not allowed annotation present")
			}

			if err := m.Annotations.NoDuplicates(); err != nil {
				log.Error().Str("ctx", "function").Str("function", m.Name).Str("at", m.Pos.String()).Err(err).Send()
				rc = false
			}

			if err := m.Annotations.MustHaveExactlyOneOutOf("@GET", "@PUT", "@POST", "@PATCH", "@HEAD", "@DELETE"); err != nil {
				rc = false
				log.Error().Str("ctx", "function").Str("function", m.Name).Str("at", m.Pos.String()).Err(err).Msg("conflicting http methods")
			}

			methodPath := false
			if m.Annotations.GetFirstIn("@path") != nil {
				methodPath = true
			}

			// Check only if the above validates and no @Path annotation is found
			if rc && !methodPath {
				httpMethod := m.Annotations.GetFirstIn("@get", "@put", "@post", "@delete", "@head", "@patch")
				if _, ok := httpMethodsMissingPathAnnotation[strings.ToLower(httpMethod.GetName())]; ok {
					rc = false
					log.Error().Str("rule", "pathMissing").Str("function", m.Name).Str("at", m.Pos.String()).Msg("multiple methods have missing path but equal http method")
				} else {
					httpMethodsMissingPathAnnotation[strings.ToLower(httpMethod.GetName())] = struct{}{}
				}
				if !groupPath {
					rc = false
					log.Error().Str("rule", "pathMissing").Str("function", m.Name).Str("at", m.Pos.String()).Msg("path is missing but is not provided at the group level either")
				}
			}

			for _, mp := range m.Params {

				if err := mp.Annotations.Accept("@PARAM"); err != nil {
					rc = false
					log.Error().Str("ctx", "param").Str("function", m.Name).Str("param", mp.Name).Str("at", mp.FilePos.String()).Err(err).Msg("not allowed annotation present")
				}

				switch m.FType {
				case GIN_SimpleHandler:
					if len(mp.Annotations) > 0 {
						log.Error().Str("ctx", "param").Str("function", m.Name).Str("param", mp.Name).Str("at", mp.FilePos.String()).Msg("no annotation in simple handlers")
					}
				default:
					if a := mp.Annotations.GetFirstIn("@Param"); a == nil {
						log.Warn().Str("ctx", "param").Str("function", m.Name).Str("param", mp.Name).Str("at", mp.FilePos.String()).Msg("method param is missing @Param annotation")
					}
				}
			}
		}
	}

	return rc
}
