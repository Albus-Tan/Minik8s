package gpu

import (
	"log"
	"minik8s/pkg/api/core"
	"os"
	"path/filepath"
	"testing"
)

func Test_server_generateJobScript(t *testing.T) {
	path, _ := os.Getwd()
	path = filepath.Join(path, "../../examples/job/matrix-sum.json")

	file, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Error reading file %s, err %v\n", path, err)
	}

	job := &core.Job{}
	err = job.JsonUnmarshal(file)
	if err != nil {
		t.Errorf("Error json unmarshal: %v", err)
	}

	script := GenerateJobScript(job)

	log.Printf(script)
}
