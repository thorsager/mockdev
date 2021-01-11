package scripts

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

var Shell = "/bin/sh"

func Execute(script string, envVars map[string]string) ([]byte, []byte, error) {
	var envVarList = asVarList(envVars)
	//args := argSplitter.FindAllString(l, -1)
	cmd := exec.CommandContext(context.Background(), Shell, "-c", script)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	cmd.Env = append(os.Environ(), envVarList...)
	err = cmd.Start()
	if err != nil {
		return nil, nil, err
	}

	outSlurp, _ := ioutil.ReadAll(stdout)
	errSlurp, _ := ioutil.ReadAll(stderr)

	err = cmd.Wait()
	if err != nil {
		return outSlurp, errSlurp, fmt.Errorf("while executing '%s': %w", script, err)
	}
	return outSlurp, errSlurp, nil
}

func asVarList(envVars map[string]string) []string {
	var varList []string
	for k, v := range envVars {
		varList = append(varList, fmt.Sprintf("%s=%s", k, v))
	}
	return varList
}
