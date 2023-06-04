package config

import (
	"os"
	"time"
)

const latestVersion = "v1.0.1"

func Version() string {
	return latestVersion
}

const HttpScheme = "http://"

/*--------------- ApiServer ---------------*/
// Http server gin config
const (
	Port = ":8080"
)

func Host() string {
	return os.Getenv("API_SERVER")
}

func ApiUrl() string {
	return HttpScheme + Host() + Port + "/api/"
}

func ApiServerUrl() string {
	return HttpScheme + Host() + Port
}

// Etcd storage config
const (
	EtcdHost = "localhost"
	EtcdPort = ":2379"
)

/*--------------- Kubelet ---------------*/
// cadvisor config
const (
	CadvisorHost = "localhost"
	CadvisorPort = ":8090"
)

func CadvisorUrl(HostAddress string) string {
	return HttpScheme + HostAddress + CadvisorPort
}

/*--------------- GPU ---------------*/
// HPC config
const (
	PiHost            = "pilogin.hpc.sjtu.edu.cn"
	HPCJobDirPrefix   = "job-"
	HPCHomeDir        = "/lustre/home/acct-stu/stu1653/"
	OutputFileSuffix  = ".out"
	ErrorFileSuffix   = ".err"
	SlurmFileSuffix   = ".slurm"
	CuFileSuffix      = ".cu"
	MailAddressSuffix = "@sjtu.edu.cn"
)

/*--------------- Heartbeat ---------------*/
const (
	HeartbeatInterval      = time.Duration(10) * time.Second
	HeartbeatDeadInterval  = time.Duration(90) * time.Second
	HeartbeatCheckInterval = time.Duration(15) * time.Second
)

/*--------------- Serverless ---------------*/
const (
	FuncDefaultInitInstanceNum = 0  // Default instance number when func template is created
	FuncDefaultMaxInstanceNum  = 10 // Default max instance number for each func template
	FuncDefaultMinInstanceNum  = 0  // Default min instance number for each func template
)

const (
	FuncCallColdBootWait           = time.Duration(200) * time.Millisecond
	FuncInstanceScaleDownNum       = 1
	FuncInstanceScaleDownInterval  = time.Duration(60) * time.Second
	FuncInstanceScaleCheckInterval = time.Duration(10) * time.Second
)
