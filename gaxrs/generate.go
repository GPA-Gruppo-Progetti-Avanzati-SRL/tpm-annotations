package gaxrs

import (
	"embed"
	"errors"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-annotations/util"
	"github.com/rs/zerolog/log"
	"path/filepath"
	"strings"
	"text/template"
)

/*
type Config struct {
	files []string
}

type Option func(cfg *Config)

func Process(opts ...Option) error {

	cfg := Config{}
	for _, o := range opts {
		o(&cfg)
	}

	ss := make([]scanner.ScanSource, 0)
	if len(cfg.files) > 0 {
		for _, f := range cfg.files {
			ss = append(ss, scanner.ScanSource{FileName: f})
		}
	}

	if len(ss) == 0 {
		err := errors.New("no files selected for processing")
		log.Error().Err(err).Send()
		return err
	}

	fds, err := scanner.Scan(ss)
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
*/

//go:embed templates/*
var templates embed.FS

// GenerationContext:
// TODO: to understand the use of it...

type GenerationContext struct {
}

func Generate(gs []G) {

	_, err := templates.ReadFile("templates/README.md")
	if err != nil {
		panic(err)
	}

	groupFds := make([]G, 0)
	for _, f := range gs {
		if len(groupFds) == 0 {
			groupFds = append(groupFds, f)
		} else {
			// Rispetto allo scanner_test non gestisco la non compatibilità dei path.... Dovrei riproporre lo stesso codice
			// togliendolo da la' che se no rimane ripetuto.
			if strings.Compare(filepath.Dir(f.FileName), filepath.Dir(groupFds[0].FileName)) == 0 {
				groupFds = append(groupFds, f)
			} else {
				// Process group and clear
				log.Info().Str("package", groupFds[0].Package).Int("numberOfFiles", len(groupFds)).Msgf("%s %s", strings.Repeat("=", 20), "GAXRS: Package Generation")
				for _, f := range groupFds {
					log.Info().Str("filename", f.Name).Send()
				}
				groupFds = make([]G, 0)
				// TODO: should probably put the item in the group.
				groupFds = append(groupFds, f)
			}
		}
	}

	if len(groupFds) > 0 {
		// Process group
		log.Info().Str("package", groupFds[0].Package).Int("numberOfFiles", len(groupFds)).Msgf("%s %s", strings.Repeat("=", 20), "GAXRS: Package Generation")
		for _, f := range groupFds {
			log.Info().Str("filename", f.Name).Send()
		}
	}
}

func emit(genCtx GenerationContext, resDir string, outFolder string, generatedFileName string, templates []string, formatCode bool) error {
	if t, ok := loadTemplate(resDir, templates...); ok {
		destinationFile := filepath.Join(outFolder, generatedFileName)
		log.Info().Str("dest", destinationFile).Msg("generating text from template")

		if err := parseTemplateWithFuncMapsProcessWrite2File(t, getTemplateUtilityFunctions(), genCtx, destinationFile, formatCode); err != nil {
			log.Error().Err(err).Msg("parse template failed")
			return err
		}
	} else {
		log.Error().Msg("unable to load template ...skipping")
		return errors.New("unable to load template ...skipping")
	}

	return nil
}

func getOutputFilename(prefix string, baseName string) string {
	if prefix != "" {
		return strings.Join([]string{prefix, baseName}, "-")
	}

	return baseName
}

func loadTemplate(resDirectory string, templatePath ...string) ([]util.TemplateInfo, bool) {

	res := make([]util.TemplateInfo, 0)
	for _, tpath := range templatePath {

		tmplContent, err := templates.ReadFile(tpath)
		if err != nil {
			log.Error().Str("path", tpath).Msg("unable to read template")
			return nil, false
		}

		/*
		 * Get the name of the template from the file name.... Hope there is one dot only...
		 * Dunno it is a problem.
		 */
		tname := filepath.Base(tpath)
		if ext := filepath.Ext(tname); ext != "" {
			tname = strings.TrimSuffix(tname, ext)
		}

		res = append(res, util.TemplateInfo{Name: tname, Content: string(tmplContent)})
	}

	return res, true
}

func getTemplateUtilityFunctions() template.FuncMap {

	fMap := template.FuncMap{
		"classify":   util.Classify,
		"dasherize":  util.Dasherize,
		"camelize":   util.Camelize,
		"capitalize": util.Capitalize,
		"underscore": util.Underscore,
		"decamelize": util.Decamelize,
	}

	return fMap
}

func parseTemplateWithFuncMapsProcessWrite2File(templates []util.TemplateInfo, fMaps template.FuncMap, templateData interface{}, outputFile string, formatSource bool) error {

	if pkgTemplate, err := util.ParseTemplates(templates, fMaps); err != nil {
		return err
	} else {
		if err := util.ProcessTemplateWrite2File(pkgTemplate, templateData, outputFile, formatSource); err != nil {
			return err
		}
	}

	return nil
}
