package formatter

import (
	"os"
	"strings"

	"github.com/daixiang0/gci/pkg/config"
	"github.com/daixiang0/gci/pkg/gci"
	"github.com/daixiang0/gci/pkg/log"
	"github.com/daixiang0/gci/pkg/section"
	"golang.org/x/mod/modfile"
)

func init() {
	log.InitLogger()
}

func formatImports(filePath string, content []byte) ([]byte, error) {
	cfg, err := buildGCIConfig(filePath)
	if err != nil {
		return content, nil
	}

	_, formatted, err := gci.LoadFormat(content, filePath, *cfg)
	if err != nil {
		return nil, err
	}

	return formatted, nil
}

func buildGCIConfig(filePath string) (*config.Config, error) {
	prefix := detectModulePrefix(filePath)

	sections := []section.Section{
		section.Standard{},
		section.Default{},
	}

	if prefix != "" {
		sections = append(sections, section.Custom{Prefix: prefix})
	}

	return &config.Config{
		BoolConfig: config.BoolConfig{
			SkipGenerated: true,
			SkipVendor:    true,
		},
		Sections:          sections,
		SectionSeparators: section.DefaultSectionSeparators(),
	}, nil
}

func detectModulePrefix(filePath string) string {
	modPath := findGoMod(filePath)
	if modPath == "" {
		return ""
	}

	data, err := os.ReadFile(modPath)
	if err != nil {
		return ""
	}

	mf, err := modfile.Parse(modPath, data, nil)
	if err != nil || mf.Module == nil {
		return ""
	}

	return extractOrgPrefix(mf.Module.Mod.Path)
}

func extractOrgPrefix(modulePath string) string {
	parts := strings.Split(modulePath, "/")
	if len(parts) < 2 {
		return ""
	}

	return strings.Join(parts[:2], "/") + "/"
}
