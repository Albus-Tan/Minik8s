package kubectl

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"minik8s/pkg/apiclient"
)

var describeCmd = &cobra.Command{
	Use:   "describe <resource type>",
	Short: "describe api objects.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := args[0]

		objType, err := ParseType(s)
		if err != nil {
			fmt.Printf("No %v type of resource, err: %v\n", s, err)
			return
		}

		cli, _ := apiclient.NewRESTClient(objType)
		list, err := cli.GetAll()
		if err != nil {
			fmt.Printf("Get all %v failed, err: %v\n", objType, err)
			return
		}

		for _, obj := range list.GetIApiObjectArr() {
			jsonData, err := obj.JsonMarshal()
			if err != nil {
				fmt.Printf("JsonMarshal of obj type %v failed, err: %v\n", objType, err)
				return
			}

			yamlData, err := yaml.JSONToYAML(jsonData)
			if err != nil {
				fmt.Printf("JSONToYAML of obj type %v failed, err: %v\n", objType, err)
				return
			}

			fmt.Println(string(yamlData))
		}
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)
}
