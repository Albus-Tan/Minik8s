package kubectl

import (
	"fmt"
	"github.com/spf13/cobra"
	"minik8s/pkg/api/core"
	"minik8s/pkg/apiclient"
	"minik8s/utils"
)

func doCreate(cmd *cobra.Command, args []string) {
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

	if err != nil {
		fmt.Printf("%v created failed, object JsonUnmarshal err: %v\n", objType, err)
		return
	}
	code, resp, err := cli.Post(object)
	if err != nil {
		fmt.Printf("%v created failed, http status code %v, err: %v\n", objType, code, resp.ErrorMsg)
		return
	}

	fmt.Printf("%v created success, uid: %v\n", objType, resp.UID)
}

func doDelete(cmd *cobra.Command, args []string) {
	s := args[0]
	name := args[1]

	objType, err := ParseType(s)
	if err != nil {
		fmt.Printf("No %v type of resource, err: %v\n", s, err)
		return
	}

	cli, _ := apiclient.NewRESTClient(objType)
	code, resp, err := cli.Delete(name)

	if err != nil {
		fmt.Printf("%v delete failed, http status code %v, err: %v\n", objType, code, resp.ErrorMsg)
		return
	}

	fmt.Printf("%v delete success\n", objType)
}
