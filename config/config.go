package config

const (
	HttpScheme    = "http://"
	Host          = "localhost"
	Port          = ":8080"
	latestVersion = "v1.0.1"
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
