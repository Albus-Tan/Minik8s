package kubectl

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"minik8s/config"
	"net/http"
)

type metadata struct {
	Name            string `json:"name"`
	ResourceVersion string `json:"resourceVersion"`
	Uid             string `json:"uid"`
	NameSpace       string `json:"namespace"`
}

type sepc struct {
	NodeName    string `json:"nodeName"`
	PodCIDR     string `json:"podCIDR"`
	Replicas    int    `json:"replicas"`
	MinReplicas int    `json:"minReplicas"`
	MaxReplicas int    `json:"maxReplicas"`
}
type status struct {
	Phase         string `json:"phase"`
	PodIp         string `json:"podIp"`
	LastScaleTime string `json:"lastScaleTime"`
}

type pod struct {
	Metadata metadata `json:"metadata"`
	Spec     sepc     `json:"spec"`
	Status   status   `json:"status"`
}
type pods struct {
	Pods []pod `json:"items"`
}

type node struct {
	Metadata metadata `json:"metadata"`
	Status   status   `json:"status"`
	Sepc     sepc     `json:"spec"`
}
type nodes struct {
	Nodes []node `json:"items"`
}

type replicaset struct {
	Metadata metadata `json:"metadata"`
	Spec     sepc     `json:"spec"`
	Status   status   `json:"status"`
}
type replicasets struct {
	Replicasets []replicaset `json:"items"`
}

type hpa struct {
	Metadata metadata `json:"metadata"`
	Spec     sepc     `json:"spec"`
	Status   status   `json:"status"`
}
type hpas struct {
	Hpas []hpa `json:"items"`
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get pods or namespaces.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := args[0]
		switch s {
		case "pods":
			//get localhost:8080/api/pods:name
			url := config.ApiUrl() + "pods/"

			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			var pods pods
			err := json.Unmarshal(str, &pods)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("%-20s\t%-40s\t%-8s\t%-15s\t%-15s\n", "NAME", "UID", "NODE", "STATUS", "IP")
			for _, pod := range pods.Pods {
				fmt.Printf("%-20s\t%-40s\t%-8s\t%-15s\t%-15s\n", pod.Metadata.Name, pod.Metadata.Uid, pod.Spec.NodeName, pod.Status.Phase, pod.Status.PodIp)
			}
		case "pod":
			if len(args) < 2 {
				fmt.Println("please input the pod name")
				return
			}
			podname := args[1]
			url := config.ApiUrl() + "pods/"
			url = url + podname
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			var pod pod
			err := json.Unmarshal(str, &pod)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("%-20s\t%-40s\t%-8s\t%-15s\t%-15s\n", "NAME", "UID", "NODE", "STATUS", "IP")
			fmt.Printf("%-20s\t%-40s\t%-8s\t%-15s\t%-15s\n", pod.Metadata.Name, pod.Metadata.Uid, pod.Spec.NodeName, pod.Status.Phase, pod.Status.PodIp)
		case "podstatus":
			url := config.ApiUrl() + "pods/"
			if len(args) < 2 {
				fmt.Println("please input the pod name")
				return
			}
			podname := args[1]
			url = url + podname + "/status"
			req, _ := http.NewRequest("GET", url, nil)
			namespace := GetNamespace()
			req.Header.Add("namespace", namespace)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			var status status
			err := json.Unmarshal(str, &status)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("%-20s\t%-15s\n", "STATUS", "IP")
			fmt.Printf("%-20s\t%-15s\n", status.Phase, status.PodIp)
		case "clear":
			url := config.ApiUrl() + "clear/"
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)

			fmt.Println(string(str))

		case "nodes":
			url := config.ApiUrl() + "nodes/"
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)

			var nodes nodes
			err := json.Unmarshal(str, &nodes)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("%-20s\t%-40s\t%-15s\t%-15s\n", "NAME", "UID", "STATUS", "IP")
			for _, node := range nodes.Nodes {
				fmt.Printf("%-20s\t%-40s\t%-15s\t%-15s\n", node.Metadata.Name, node.Metadata.Uid, node.Status.Phase, node.Sepc.PodCIDR)
			}
		case "replicasets":
			url := config.ApiUrl() + "replicasets/"
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			var replicasets replicasets
			err := json.Unmarshal(str, &replicasets)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("%-20s\t%-40s\t%-15s\t%-15s\n", "NAME", "UID", "NameSpace", "Replicas")
			for _, replicaset := range replicasets.Replicasets {
				fmt.Printf("%-20s\t%-40s\t%-15s\t%-15d\n", replicaset.Metadata.Name, replicaset.Metadata.Uid, replicaset.Metadata.NameSpace, replicaset.Spec.Replicas)
			}

		case "replicaset":
			if len(args) < 2 {
				fmt.Println("please input the replicaset name")
				return
			}
			replicasetname := args[1]
			url := config.ApiUrl() + "replicasets/"
			url = url + replicasetname
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			var replicaset replicaset
			err := json.Unmarshal(str, &replicaset)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("%-20s\t%-40s\t%-15s\t%-15s\n", "NAME", "UID", "NameSpace", "Replicas")
			fmt.Printf("%-20s\t%-40s\t%-15s\t%-15d\n", replicaset.Metadata.Name, replicaset.Metadata.Uid, replicaset.Metadata.NameSpace, replicaset.Spec.Replicas)

		case "replicasetstatus":
			if len(args) < 2 {
				fmt.Println("please input the replicaset name")
				return
			}
			replicasetname := args[1]
			url := config.ApiUrl() + "replicasets/"
			url = url + replicasetname + "/status"
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			var status status
			err := json.Unmarshal(str, &status)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("%-20s\t%-15s\n", "STATUS", status.Phase)

		case "hpas":
			url := config.ApiUrl() + "hpas/"
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			var hpas hpas
			err := json.Unmarshal(str, &hpas)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("\n%-20s\t%-40s\t%-15s\t%-15d\t%-15d\n", "NAME", "UID", "MinReplicas", "MaxReplicas", "LastScaleTime")
			for _, hpa := range hpas.Hpas {
				fmt.Printf(
					"%-20s\t%-40s\t%-15d\t%-15d\t%-15s\n",
					hpa.Metadata.Name,
					hpa.Metadata.Uid,
					hpa.Spec.MinReplicas,
					hpa.Spec.MaxReplicas,
					hpa.Status.LastScaleTime,
				)
			}

		case "hpa":
			if len(args) < 2 {
				fmt.Println("please input the hpa name")
				return
			}
			hpaname := args[1]
			url := config.ApiUrl() + "hpas/"
			url = url + hpaname
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			var hpa hpa
			err := json.Unmarshal(str, &hpa)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("\n%-20s\t%-40s\t%-15s\t%-15s\t%-15s\n",
				"NAME", "UID", "MinReplicas", "MaxReplicas", "LastScaleTime",
			)
			fmt.Printf("%-20s\t%-40s\t%-15d\t%-15d\t%-15s\n",
				hpa.Metadata.Name,
				hpa.Metadata.Uid,
				hpa.Spec.MinReplicas,
				hpa.Spec.MaxReplicas,
				hpa.Status.LastScaleTime,
			)

		case "hpastatus":
			if len(args) < 2 {
				fmt.Println("please input the hpa name")
				return
			}
			hpaname := args[1]
			url := config.ApiUrl() + "hpas/"
			url = url + hpaname + "/status"
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			var status status
			err := json.Unmarshal(str, &status)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("%-20s\t%-15s\n", "LastScaleTime")
			fmt.Printf("%-20s\t%-15d\n", status.LastScaleTime)
		default:
			fmt.Println("get error")
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("namespace", "n", "", "kube pods' namespace")
	rootCmd.AddCommand(getCmd)
}
