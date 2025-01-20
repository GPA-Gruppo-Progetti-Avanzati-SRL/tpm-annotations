package gaxrs

import (
	"errors"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-annotations/annotations/scanner"
	"github.com/rs/zerolog/log"
	"path/filepath"
	"sort"
	"strings"
)

const (
	GIN_SimpleHandler       = "gin-handler"
	GIN_Closure_HandlerFunc = "gin-closure-handlerfunc"
	GIN_Closure_Func        = "gin-closure-handler"
)

type G struct {
	FileName string
	Package  string

	Name      string
	Path      string
	Resources []R
}

type R struct {
	Name       string
	HttpMethod string
	Path       string
	Method     string
	Params     []RP
	Closure    bool
}

type RP struct {
	Name    string
	CtxName string
	Import  string
	IsStar  bool
	QType   string
}

func BuildModel(fds []scanner.FileDef) ([]G, []error) {

	errs := make([]error, 0)
	gs := make([]G, 0)

	for _, fd := range fds {

		httpMethodsMissingPathAnnotation := make(map[string]struct{})

		var filename = filepath.Base(fd.Name)
		var extension = filepath.Ext(filename)
		if extension != "" {
			filename = filename[0 : len(filename)-len(extension)]
		}

		g := G{Package: fd.Package, FileName: fd.Name, Name: filename}

		// Check if there is a gaxrs annotation
		if fd.Annotations.HasAtLeastOneOf("@Path") {
			if err := fd.Annotations.NoDuplicates(); err != nil {
				errs = append(errs, err)
				log.Error().Str("ctx", "file").Err(err).Msg("duplicates present")
			} else {
				g.Path = fd.Annotations.GetFirstIn("@path").GetValue()
			}
		}

		for _, m := range fd.Methods {

			r := R{Method: m.Name, Name: m.Name}

			var methodErr bool
			if m.Annotations.HasAtLeastOneOf("@Path", "@GET", "@PUT", "@POST", "@PATCH", "@HEAD", "@DELETE") {
				if err := m.Annotations.NoDuplicates(); err != nil {
					errs = append(errs, err)
					methodErr = true
					log.Error().Str("ctx", "function").Str("function", m.Name).Str("at", m.Pos.String()).Err(err).Send()
				}

				if err := m.Annotations.MustHaveExactlyOneOutOf("@GET", "@PUT", "@POST", "@PATCH", "@HEAD", "@DELETE"); err != nil {
					errs = append(errs, err)
					methodErr = true
					log.Error().Str("ctx", "function").Str("function", m.Name).Str("at", m.Pos.String()).Err(err).Msg("conflicting http methods")
				}

				// Determine the type of handler.
				switch {
				case m.Receiver != nil:
					methodErr = true
					errs = append(errs, errors.New("methods with receiver functions not supported"))
					log.Error().Str("ctx", "function").Str("function", m.Name).Str("at", m.Pos.String()).Msg("methods with receiver functions not supported")

				case m.ResultType == scanner.NO_RESULT && len(m.Params) == 1:
					if m.Params[0].QType == "github.com/gin-gonic/gin/Context" {
						// Make it explicit.
						r.Closure = false
					}
				case m.ResultType == GIN_Closure_HandlerFunc:
					r.Closure = true
				case m.ResultType == GIN_Closure_Func:
					r.Closure = true
				default:
					methodErr = true
					errs = append(errs, errors.New("unsupported method signature"))
					log.Error().Str("ctx", "function").Str("function", m.Name).Str("at", m.Pos.String()).Msg("unsupported method signature")
				}

				if a := m.Annotations.GetFirstIn("@path"); a != nil {
					r.Path = a.GetValue()
				}

				httpMethod := m.Annotations.GetFirstIn("@get", "@put", "@post", "@delete", "@head", "@patch")
				r.HttpMethod = httpMethod.GetValue()

				// Check only if the above validates and no @Path annotation is found
				if !methodErr && r.Path == "" {

					if _, ok := httpMethodsMissingPathAnnotation[strings.ToLower(httpMethod.GetName())]; ok {
						methodErr = true
						errs = append(errs, errors.New("multiple methods have missing path but equal http method"))
						log.Error().Str("rule", "pathMissing").Str("function", m.Name).Str("at", m.Pos.String()).Msg("multiple methods have missing path but equal http method")
					} else {
						httpMethodsMissingPathAnnotation[strings.ToLower(httpMethod.GetName())] = struct{}{}
					}
					if g.Path == "" {
						methodErr = true
						errs = append(errs, errors.New("path is missing but is not provided at the group level either"))
						log.Error().Str("rule", "pathMissing").Str("function", m.Name).Str("at", m.Pos.String()).Msg("path is missing but is not provided at the group level either")
					}
				}

				if !methodErr {
					for _, mp := range m.Params {

						rp := RP{Name: mp.Name}

						if err := mp.Annotations.Accept("@PARAM"); err != nil {
							methodErr = true
							errs = append(errs, err)
							log.Error().Str("ctx", "param").Str("function", m.Name).Str("param", mp.Name).Str("at", mp.FilePos.String()).Err(err).Msg("not allowed annotation present")
						}

						if r.Closure {
							if a := mp.Annotations.GetFirstIn("@Param"); a == nil {
								log.Warn().Str("ctx", "param").Str("function", m.Name).Str("param", mp.Name).Str("at", mp.FilePos.String()).Msg("method param is missing @Param annotation")
								rp.CtxName = mp.Name
							} else {
								rp.CtxName = a.GetValue()
							}

							rp.IsStar = mp.Pointer
							rp.QType = mp.X
							if mp.Sel != "" {
								rp.QType = mp.X + "." + mp.Sel
							}

							r.Params = append(r.Params, rp)
							// TODO: import resolution.
						} else {
							// Todo: controllo da rimuovere in futuro.
							if len(mp.Annotations) > 0 {
								methodErr = true
								errs = append(errs, errors.New("no annotation in simple handlers"))
								log.Error().Str("ctx", "param").Str("function", m.Name).Str("param", mp.Name).Str("at", mp.FilePos.String()).Msg("no annotation in simple handlers")
							}
						}
					}
				}

				if !methodErr {
					g.Resources = append(g.Resources, r)
				}

			}
		}

		if len(g.Resources) > 0 {
			gs = append(gs, g)
		}
	}

	// Sorting to process by directory package. Groups of files in the same directory / package will output
	// a file named by package (<package>_jaxrs_gen.go) in that directory. This way should be able to handle also
	// the test packages
	if len(gs) > 0 {
		sortByPathPackage(gs)
	}

	return gs, errs
}

func sortByPathPackage(gs []G) {
	sort.Slice(gs, func(i, j int) bool {
		return fileFolderAndPackageLessThan(&gs[i], &gs[j])
	})
}

func fileFolderAndPackageLessThan(aGroup *G, another *G) bool {
	rc := false
	switch strings.Compare(filepath.Dir(aGroup.FileName), filepath.Dir(another.FileName)) {
	case 0:
		if strings.Compare(aGroup.Package, another.Package) < 0 {
			rc = true
		}
	case -1:
		rc = true
	default:
		rc = false
	}

	return rc
}
