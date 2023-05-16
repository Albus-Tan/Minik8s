package config

const latestVersion = "v1.0.1"

func Version() string {
	return latestVersion
}

const HttpScheme = "http://"

/*--------------- ApiServer ---------------*/
// Http server gin config
const (
	Host = "localhost"
	Port = ":8080"
)

func ApiUrl() string {
	return HttpScheme + Host + Port + "/api/"
}

func ApiServerUrl() string {
	return HttpScheme + Host + Port
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
	PiHost          = "pilogin.hpc.sjtu.edu.cn"
	HPCJobDirPrefix = "job-"
	HPCHomeDir      = "/lustre/home/acct-stu/stu1653/"
)
