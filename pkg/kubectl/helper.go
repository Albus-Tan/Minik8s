package kubectl

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func ExecuteCommand(name string, subname string, args ...string) (string, error) {
	args = append([]string{subname}, args...)

	cmd := exec.Command(name, args...)
	bytes, err := cmd.CombinedOutput()

	return string(bytes), err
}

func GetNamespace() string {
	namespace := "default"
	name, err := rootCmd.PersistentFlags().GetString("namespace")
	if err != nil {
		fmt.Println("the err is", err)
	}
	if name != "" {
		namespace = name
	}
	return namespace
}

func GetFilename() string {
	filename := ""
	name, err := rootCmd.PersistentFlags().GetString("filename")
	if err != nil {
		fmt.Println("the err is", err)
	}
	if name != "" {
		filename = name
	}
	return filename
}

func Error(cmd *cobra.Command, args []string, err error) {
	fmt.Fprintf(os.Stderr, "execute %s args:%v error:%v\n", cmd.Name(), args, err)
	os.Exit(1)
}
