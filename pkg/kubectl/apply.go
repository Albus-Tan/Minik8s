package kubectl

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply for yaml.",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(args)
		s := GetFilename()
		analyzeFile(s)

	},
}

func analyzeFile(path string) {
	//var unmarshal func([]byte, any) error
	//if strings.HasSuffix(path, "json") {
	//	viper.SetConfigType("json")
	//	unmarshal = json.Unmarshal
	//} else if strings.HasSuffix(path, "yaml") || strings.HasSuffix(path, "yml") {
	//	viper.SetConfigType("yaml")
	//	unmarshal = yaml.Unmarshal
	//} else {
	//	fmt.Printf("Unsupported type! Apply a yaml or json file!\n")
	//	return
	//}

	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file %s\n", path)
		return
	}
	err = viper.ReadConfig(bytes.NewReader(file))
	if err != nil {
		fmt.Printf("Error analyz0ing file %s\n", path)
		return
	}
	//let the file change into a string
	fmt.Println("the file is", string(file))

}

func init() {
	rootCmd.PersistentFlags().StringP("filename", "f", "", "filename index")
	rootCmd.AddCommand(applyCmd)
}
