package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var mainTf = flag.String("output", "./", "the output directory")
var input = flag.String("input", "", "the input yml file")

func main() {
	flag.Parse()
	transformToTerraform(*input, *mainTf)
}

func transformToTerraform(inputFile, outputPath string) {

	if err := os.MkdirAll(outputPath, 0755); err != nil {
		log.Fatal(err)
	}
	outputFile, err := os.OpenFile(filepath.Join(outputPath, "main.tf"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer closeAndLogIfFailure(outputFile)
	project, err := readFileAndMapToProjectConfiguration(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	errs := project.IsValid()
	if len(errs) > 0 {
		str := ""
		for i := range errs {
			if i > 0 {
				str += "\r\n"
			}
			err := errs[i]
			str += err.Error()
		}
		log.Fatal(str)
	}

	terraform, err := project.createTerraformScript()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := outputFile.Write([]byte(terraform)); err != nil {
		log.Fatal(err)
	}
}

type ConfigurationFile struct {
}

func readFileAndMapToProjectConfiguration(file string) (*ProjectConfiguration, error) {
	enVars := os.Environ()
	replacing := make([]string, len(enVars)*2)
	for i, en := range enVars {
		k := strings.Split(en, "=")
		replacing[i*2], replacing[i*2+1] = fmt.Sprintf("{%s}", k[0]), k[1]
	}
	replacers := strings.NewReplacer(replacing...)

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	data = []byte(replacers.Replace(string(data)))
	out := &ProjectConfiguration{}
	if err := yaml.Unmarshal(data, out); err != nil {
		return out, err
	}

	if len(out.GabiProject.ServiceAccounts) > 0 {
		out.ServiceAccountMap = make(map[string][]string, 10)
		for _, s := range out.GabiProject.ServiceAccounts {
			for _, r := range s.Roles {
				if out.ServiceAccountMap[r] == nil {
					out.ServiceAccountMap[r] = make([]string, 0, len(out.GabiProject.ServiceAccounts))
				}
				out.ServiceAccountMap[r] = append(out.ServiceAccountMap[r], s.Name)
			}
		}
	}
	if len(out.GabiProject.CustomRole) > 0 {
		out.CustomRoles = make(map[string]bool, len(out.CustomRoles))
		for i := range out.GabiProject.CustomRole {
			out.CustomRoles[out.GabiProject.CustomRole[i].Name] = true
		}
	}
	mapp := make(map[string]interface{}, 50)
	if err := yaml.Unmarshal(data, mapp); err != nil {
		return out, err
	}
	out.AsMap = mapp
	return out, nil
}
func (c *ConfigurationFile) Then(ifValid func(file *ConfigurationFile) *ConfigurationFile, onError func(err error) error) {

}
func closeAndLogIfFailure(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Println(err)
	}
}
