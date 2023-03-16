package domain

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	dynamic_struct "github.com/ihatiko/dynamic-struct"
	"github.com/ihatiko/go-chef/models"
	config_parser "github.com/ihatiko/go-chef/parse/config-parser"
	"github.com/ihatiko/log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

//go:embed templates
var templates embed.FS

//go:embed template.yml
var config []byte

func BuildDomain(args []string) {
	path := `C:\testProject`
	domainName := strings.ToLower("gray-dogs")
	packageName := strings.ReplaceAll(domainName, "-", "_")
	formattedFragmentName := toFragmentName(domainName)
	projectName, err := gerProjectName(path)
	if err != nil {
		log.Fatal(err)
	}
	capitalizeFragment := fmt.Sprintf("%s%s", strings.ToLower(string(formattedFragmentName[0])), formattedFragmentName[1:])
	env := dynamic_struct.ConstructStruct(map[string]any{
		"PackageName":                     packageName,
		"FormattedFragmentName":           formattedFragmentName,
		"CapitalizeFormattedFragmentName": capitalizeFragment,
		"DomainName":                      domainName,
		"ProjectName":                     projectName,
		"OpenApiTransport":                true,
		"GrpcTransport":                   true,
		"DiBinding":                       false,
	})
	//TODO GrpcTransport compile
	if err != nil {
		log.Fatal(err)
	}
	dirs, err := os.ReadDir(filepath.Join(path, "internal"))
	if err != nil {
		log.Fatal(err)
	}
	if !checkDir(dirs, "server") {
		log.Fatal("folder server does not exist")
	}
	if err = upsertFeaturesFolder(dirs, path); err != nil {
		log.Fatal(err)
	}
	dirs, err = os.ReadDir(filepath.Join(path, "internal/features"))
	if err != nil {
		log.Fatal(err)
	}
	if findDomain(dirs, domainName) {
		return
	}
	err = os.Mkdir(filepath.Join(path, "internal/features", domainName), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New("").Parse(string(config))
	if err != nil {
		log.Fatal(err)
	}
	buffer := bytes.NewBuffer([]byte{})
	writer := bufio.NewWriter(buffer)
	err = t.ExecuteTemplate(writer, "", env)
	if err != nil {
		log.Fatal(err)
	}
	err = writer.Flush()
	if err != nil {
		log.Fatal(err)
	}

	viperCfg, err := config_parser.LoadConfig(buffer.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config_parser.ParseConfig[models.Config](viperCfg)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Tree.Settings = &models.Settings{
		DomainSettings: &models.DomainConfig{
			ProjectPath: path,
			ProjectName: projectName,
		},
	}
	BuildTree(cfg.Tree, env)
	fmt.Println(cfg)
}

func BuildFiles(path string, node *models.Node, obj any) {
	for _, file := range node.GeneratedFiles {
		b, err := templates.ReadFile(fmt.Sprintf("templates/%s", file.Template))

		if err != nil {
			panic(err)
		}
		t, err := template.New("").Parse(string(b))
		p := filepath.Join(path, node.Name, fmt.Sprintf("%s.%s", file.Name, file.Extension))
		f, err := os.Create(p)
		if err != nil {
			panic(err)
		}
		err = t.ExecuteTemplate(f, "", obj)
		if err != nil {
			panic(err)
		}
	}
}

func BuildTree(tree *models.Tree, env any) {
	for _, nd := range tree.DomainComponents {
		BuildNodes(tree.Settings.DomainSettings.ProjectPath, nd, env)
	}
}

func BuildNodes(path string, node *models.Node, env any) {
	if len(node.Nodes) > 0 || len(node.GeneratedFiles) > 0 {
		os.Mkdir(filepath.Join(path, node.Name), os.ModePerm)
	}
	for _, nd := range node.Nodes {
		BuildNodes(filepath.Join(path, node.Name), nd, env)
		BuildFiles(filepath.Join(path, node.Name), nd, env)
	}
}
func toFragmentName(domainName string) string {
	re, err := regexp.Compile(`[^\w]`)
	if err != nil {
		log.Fatal(err)
	}
	fragmentName := re.ReplaceAllString(domainName, " ")
	var formattedFragmentName = ""
	for _, data := range strings.Split(fragmentName, " ") {
		if len(data) == 1 {
			formattedFragmentName += strings.ToUpper(string(data[0]))
			continue
		}
		formattedFragmentName += strings.ToUpper(string(data[0])) + data[1:]
	}
	return strings.TrimSpace(formattedFragmentName)
}

func findDomain(dirs []os.DirEntry, data string) bool {
	for _, dir := range dirs {
		if dir.Name() == data {
			return true
		}
	}

	return false
}

func upsertFeaturesFolder(dirs []os.DirEntry, path string) error {
	if !checkDir(dirs, "features") {
		return os.Mkdir(filepath.Join(path, "internal/features"), os.ModePerm)
	}
	return nil
}
func checkDir(dirs []os.DirEntry, name string) bool {
	for _, dir := range dirs {
		if dir.Name() == name {
			return true
		}
	}
	return false
}

func gerProjectName(path string) (string, error) {
	combinedPath := filepath.Join(path, "go.mod")
	f, err := os.Open(combinedPath)
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(f)
	line, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}
	projectName := strings.Replace(string(line), "module ", "", 1)
	return projectName, nil
}
