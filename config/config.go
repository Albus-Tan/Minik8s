package config

const (
	HttpScheme    = "http://"
	Host          = "localhost"
	Port          = ":8080"
	latestVersion = "v1.0.1"
)

const (
	EtcdHost = "111.186.56.24"
	EtcdPort = ":2380"
)

func Version() string {
	return latestVersion
}

func ApiUrl() string {
	return HttpScheme + Host + Port + "/api/"
}

func ApiServerUrl() string {
	return HttpScheme + Host + Port
}
