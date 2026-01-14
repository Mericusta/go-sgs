package common

import "os/exec"

func ExecShellCommandInDir(cmd string, dir string) error {
	shellCmd := ExecShellCommand(cmd)
	shellCmd.Dir = dir
	_, err := shellCmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func ExecShellCommand(cmd string) *exec.Cmd {
	return exec.Command("bash", "-c", cmd)
}
