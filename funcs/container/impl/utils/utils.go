// Package utils funcs export to docker
package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/utils/exec"
)

func ContainerInspectMap(containerName string) (result map[string]interface{}, err error) {
	if strings.TrimSpace(containerName) == "" {
		return result, fmt.Errorf("containerName不能为空")
	}
	cmd := exec.New().Command("docker", "inspect", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result, fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	containers := make([]map[string]interface{}, 0)
	err = json.Unmarshal(output, &containers)
	if err != nil || len(containers) == 0 {
		return result, err
	}
	return containers[0], nil
}

func GetValueN(data map[string]interface{}, keys ...string) interface{} {
	val, _ := GetValue(data, keys...)
	return val
}

func GetValue(data map[string]interface{}, keys ...string) (interface{}, bool) {
	for i, key := range keys {
		if i == len(keys)-1 {
			val, ok := data[key]
			return val, ok
		}
		data, _ = data[key].(map[string]interface{})
	}

	return nil, false
}

func GetBool(facts map[string]interface{}, keys ...string) bool {
	val := GetValueN(facts, keys...)
	res, ok := val.(bool)
	return ok && res
}

func GetKubectlVersion() (result string, err error) {
	cmd := exec.New().Command("kubectl", "version", "--client", "--short=true")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result, err
	}
	if outputStr := string(output); strings.Contains(outputStr, "Client Version:") {
		result = strings.TrimSpace(strings.TrimPrefix(outputStr, "Client Version:"))
		return strings.TrimPrefix(result, "v"), err
	}

	return result, fmt.Errorf("no version")
}
