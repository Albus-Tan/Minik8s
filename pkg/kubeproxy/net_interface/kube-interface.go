package net_interface

import (
	"os/exec"
)

const kubeNetworkDevice = "kube-proxy0"
const ipCommand = "ip"

func AddIPV4(ip string) error {
	return exec.Command(ipCommand, "a", "a", ip+"/32", "dev", kubeNetworkDevice).Run()
}
func DelIPV4(ip string) error {
	return exec.Command(ipCommand, "a", "d", ip+"/32", "dev", kubeNetworkDevice).Run()
}
