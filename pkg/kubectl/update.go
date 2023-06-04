package kubectl

import (
	"fmt"
	"github.com/spf13/cobra"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/apiclient"
	"minik8s/utils"
)

var updateCmd = &cobra.Command{
	Use:   "update <resource> (<resource-name>) -f <filename>",
	Short: "update resources",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		s := args[0]

		objType, err := ParseType(s)
		if err != nil {
			fmt.Printf("No %v type of resource, err: %v\n", s, err)
			return
		}

		filename := GetFilename()
		jsonData, err := utils.GetFormJsonData(filename)
		if err != nil {
			fmt.Println("File parse err:", err)
			return
		}

		cli, _ := apiclient.NewRESTClient(objType)
		object := core.CreateApiObject(objType)
		err = object.JsonUnmarshal(jsonData)

		var name string
		if len(args) >= 2 {
			name = args[1]
		} else {
			if objType == types.FuncTemplateObjectType {
				name = object.(*core.Func).Name
			} else {
				name = object.GetUID()
			}
		}

		code, resp, err := cli.Put(name, object)
		if err != nil {
			fmt.Printf("%v put failed, http status code %v, err: %v, %v\n", objType, code, resp.ErrorMsg, err)
			return
		}

		fmt.Printf("%v put success, resource version: %v\n", objType, resp.ResourceVersion)

		return

	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
