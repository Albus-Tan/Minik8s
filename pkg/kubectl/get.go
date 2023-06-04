package kubectl

import (
	"fmt"
	"github.com/spf13/cobra"
	"minik8s/pkg/apiclient"
)

var getCmd = &cobra.Command{
	Use:     "get <resources> | (<resource> <resource-name>)",
	Example: "get pods {uid}\nget pods\n",
	Short:   "get resources by resource name",
	Args:    cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		s := args[0]
		objType, err := ParseType(s)
		if err != nil {
			fmt.Printf("No %v type of resource, err: %v\n", s, err)
			return
		}

		cli, _ := apiclient.NewRESTClient(objType)

		if len(args) == 1 {

			objList, err := cli.GetAll()
			if err != nil {
				fmt.Printf("%v get failed, err: %v\n", objType, err)
				return
			}

			objList.PrintBrief()

			return
		} else {
			if len(args) >= 2 {
				name := args[1]
				obj, err := cli.Get(name)
				if err != nil {
					fmt.Printf("%v get failed, err: %v\n", objType, err)
					return
				}
				obj.PrintBrief()
				return
			}
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("namespace", "n", "", "kube object' namespace")
	rootCmd.AddCommand(getCmd)
}
