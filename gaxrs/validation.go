package gaxrs

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-annotations/annotations/scanner"
	"github.com/rs/zerolog/log"
	"strings"
)

// Validate
// Deprecated
func Validate(fds []scanner.FileDef) bool {

	rc := true
	for _, fd := range fds {

		httpMethodsMissingPathAnnotation := make(map[string]struct{})
		groupPath := false

		// Check if there is a gaxrs annotation
		if fd.Annotations.HasAtLeastOneOf("@Path") {
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

			if m.Annotations.HasAtLeastOneOf("@Path", "@GET", "@PUT", "@POST", "@PATCH", "@HEAD", "@DELETE") {
				if err := m.Annotations.NoDuplicates(); err != nil {
					log.Error().Str("ctx", "function").Str("function", m.Name).Str("at", m.Pos.String()).Err(err).Send()
					rc = false
				}

				if err := m.Annotations.MustHaveExactlyOneOutOf("@GET", "@PUT", "@POST", "@PATCH", "@HEAD", "@DELETE"); err != nil {
					rc = false
					log.Error().Str("ctx", "function").Str("function", m.Name).Str("at", m.Pos.String()).Err(err).Msg("conflicting http methods")
				}

				// Determine the type of handler.
				switch {
				case m.Receiver != nil:
					rc = false
					log.Error().Str("ctx", "function").Str("function", m.Name).Str("at", m.Pos.String()).Msg("methods with receiver functions not supported")
				case m.ResultType == scanner.NO_RESULT && len(m.Params) == 1:
					if m.Params[0].QType == "github.com/gin-gonic/gin/Context" {
						// TODO: qui non va bene. si perde...
						m.ResultType = GIN_SimpleHandler
					}
				case m.ResultType == GIN_Closure_HandlerFunc:
				case m.ResultType == GIN_Closure_Func:
				default:
					rc = false
					log.Error().Str("ctx", "function").Str("function", m.Name).Str("at", m.Pos.String()).Msg("unsupported method signature")
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

				if rc {
					for _, mp := range m.Params {

						if err := mp.Annotations.Accept("@PARAM"); err != nil {
							rc = false
							log.Error().Str("ctx", "param").Str("function", m.Name).Str("param", mp.Name).Str("at", mp.FilePos.String()).Err(err).Msg("not allowed annotation present")
						}

						switch m.ResultType {
						case GIN_SimpleHandler:
							// Todo: controllo da rimuovere in futuro.
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
		}
	}

	return rc
}
