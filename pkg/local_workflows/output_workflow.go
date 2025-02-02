package localworkflows

import (
	"fmt"
	"strings"

	iUtils "github.com/snyk/go-application-framework/internal/utils"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/spf13/pflag"
)

var WORKFLOWID_OUTPUT_WORKFLOW workflow.Identifier = workflow.NewWorkflowIdentifier("output")

const (
	OUTPUT_CONFIG_KEY_JSON      = "json"
	OUTPUT_CONFIG_KEY_JSON_FILE = "json-file-output"
)

// InitOutputWorkflow initializes the output workflow
// The output workflow is responsible for handling the output destination of workflow data
// As part of the localworkflows package, it is registered via the localworkflows.Init method
func InitOutputWorkflow(engine workflow.Engine) error {
	outputConfig := pflag.NewFlagSet("output", pflag.ExitOnError)
	outputConfig.Bool(OUTPUT_CONFIG_KEY_JSON, false, "Print json output to console")
	outputConfig.String(OUTPUT_CONFIG_KEY_JSON_FILE, "", "Write json output to file")

	entry, err := engine.Register(WORKFLOWID_OUTPUT_WORKFLOW, workflow.ConfigurationOptionsFromFlagset(outputConfig), outputWorkflowEntryPointImpl)
	entry.SetVisibility(false)

	return err
}

// outputWorkflowEntryPoint defines the output entry point
// the entry point is called by the engine when the workflow is invoked
func outputWorkflowEntryPoint(invocation workflow.InvocationContext, input []workflow.Data, outputDestination iUtils.OutputDestination) (output []workflow.Data, err error) {
	err = nil
	output = []workflow.Data{}

	config := invocation.GetConfiguration()
	debugLogger := invocation.GetLogger()

	printJsonToCmd := config.GetBool(OUTPUT_CONFIG_KEY_JSON)
	writeJsonToFile := config.GetString(OUTPUT_CONFIG_KEY_JSON_FILE)

	for i := range input {
		mimeType := input[i].GetContentType()
		contentLocation := input[i].GetContentLocation()
		if len(contentLocation) == 0 {
			contentLocation = "unknown"
		}

		debugLogger.Printf("Processing '%s' based on '%s' of type '%s'\n", input[i].GetIdentifier().String(), contentLocation, mimeType)

		if strings.Contains(mimeType, OUTPUT_CONFIG_KEY_JSON) { // handle application/json
			singleData := input[i].GetPayload().([]byte)

			// if json data is processed but non of the json related output configuration is specified, default printJsonToCmd is enabled
			if printJsonToCmd == false && len(writeJsonToFile) == 0 {
				printJsonToCmd = true
			}

			if printJsonToCmd {
				outputDestination.Println(string(singleData))
			}

			if len(writeJsonToFile) > 0 {
				debugLogger.Printf("Writing '%s' JSON of length %d to '%s'\n", input[i].GetIdentifier().String(), len(singleData), writeJsonToFile)

				outputDestination.Remove(writeJsonToFile)
				outputDestination.WriteFile(writeJsonToFile, singleData, iUtils.FILEPERM_666)
			}
		} else { // handle text/pain and unknown the same way
			// try to convert payload to a string
			singleDataAsString := ""
			singleData, typeCastSuccessful := input[i].GetPayload().([]byte)
			if !typeCastSuccessful {
				singleDataAsString, typeCastSuccessful = input[i].GetPayload().(string)
				if !typeCastSuccessful {
					err := fmt.Errorf("Unsupported output type: %s", mimeType)
					return output, err
				}
			} else {
				singleDataAsString = string(singleData)
			}

			outputDestination.Println(singleDataAsString)
		}
	}

	return output, err
}

func outputWorkflowEntryPointImpl(invocation workflow.InvocationContext, input []workflow.Data) (output []workflow.Data, err error) {
	outputDestination := iUtils.NewOutputDestination()
	return outputWorkflowEntryPoint(invocation, input, outputDestination)
}
