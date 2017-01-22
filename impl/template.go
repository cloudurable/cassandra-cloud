package impl

import (
	"os"
	"fmt"
	"io/ioutil"
	"text/template"
	lg "github.com/advantageous/go-logback/logging"

)

func ProcessTemplate(inputFileName string, outputFileName string, any interface{}, logger lg.Logger) error {
	bytes, err := ioutil.ReadFile(inputFileName)
	if err != nil {
		logger.Errorf("Unable to load template %s  \n", inputFileName)
		logger.ErrorError("Error was", err)
		return err
	}

	template, err := template.New("test").Parse(string(bytes))
	if err != nil {
		logger.Errorf("Unable to parse template %s  \n", inputFileName)
		logger.ErrorError("Error was", err)
		return err
	}

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		logger.ErrorError(fmt.Sprintf("Unable to open output file %s", outputFileName), err)
		return err
	}
	defer outputFile.Close()
	template.Execute(outputFile, any)
	return nil
}

