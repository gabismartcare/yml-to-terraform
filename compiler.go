package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func readConfigurationFile(inputFile string) *ProjectConfiguration {
	enVars := os.Environ()
	replacing := make([]string, len(enVars)*2)
	for i, en := range enVars {
		k := strings.Split(en, "=")
		replacing[i*2], replacing[i*2+1] = fmt.Sprintf("{%s}", k[0]), k[1]
	}
	replacers := strings.NewReplacer(replacing...)

	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	data = []byte(replacers.Replace(string(data)))
	out := &ProjectConfiguration{}
	if err := yaml.Unmarshal(data, out); err != nil {
		log.Fatalf("error: %v", err)
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
		for _, st := range out.GabiProject.Storage {
			if len(st.Backup) > 0 {
				out.HasBackup = true
				break
			}
		}
	}
	mapp := make(map[string]interface{}, 50)
	if err := yaml.Unmarshal(data, mapp); err != nil {
		log.Fatalf("error: %v", err)
	}
	out.AsMap = mapp
	return out
}
