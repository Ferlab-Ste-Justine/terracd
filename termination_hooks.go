package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

type OpResult int64

const (
	OpSuccess OpResult = iota
	OpFailure
	OpSkip
)

type TerminationHooks struct {
	Success TerminationHook
	Failure TerminationHook
	Skip    TerminationHook
	Always  TerminationHook
}

func (hooks *TerminationHooks) Run(result OpResult) error {
	if hooks.Always.IsDefined() {
		return hooks.Always.Run()
	} else if result == OpSuccess && hooks.Success.IsDefined() {
		return hooks.Success.Run()
	} else if result == OpFailure && hooks.Failure.IsDefined() {
		return hooks.Failure.Run()
	} else if result == OpSkip && hooks.Skip.IsDefined() {
		return hooks.Skip.Run()
	}

	return nil
}

type TerminationHook struct {
	Command  TerminationHookCommand
	HttpCall TerminationHookHttpCall `yaml:"http_call"`
}

func (hook *TerminationHook) IsDefined() bool {
	return hook.Command.Command != "" || hook.HttpCall.Method != ""
}

func (hook *TerminationHook) Run() error {
	if hook.Command.Command != "" {
		return hook.Command.Run()
	}

	return hook.HttpCall.Run()
}

type TerminationHookCommand struct {
	Command string
	Args    []string
}

func (cmd *TerminationHookCommand) Run() error {
	cmdShow := strings.Join(append([]string{cmd.Command}, cmd.Args...), " ")
	fmt.Printf("Running termination hook command: %s\n", cmdShow)

	out, err := exec.Command(cmd.Command, cmd.Args...).Output()
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}

type TerminationHookHttpCall struct {
	Method   string
	Endpoint string
}

func (call *TerminationHookHttpCall) Run() error {
	fmt.Printf("Running termination hook http call: %s %s\n", call.Method, call.Endpoint)

	req, reqErr := http.NewRequest(call.Method, call.Endpoint, nil)
	if reqErr != nil {
		return reqErr
	}

	_, resErr := http.DefaultClient.Do(req)
	if resErr != nil {
		return resErr
	}

	return nil
}
