package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
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
	configuration := readConfigurationFile(inputFile)
	if _, err := outputFile.Write([]byte(configuration.ToTerraform())); err != nil {
		log.Fatal(err)
	}
}
func closeAndLogIfFailure(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Println(err)
	}
}
