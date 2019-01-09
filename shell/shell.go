///////////////////////////////////////////////////////////
// shell.go
// execute a shell command
// ycyxuehan kun1.huang@outlook.com
//////////////////////////////////////////////////////////

package shell

import (
	"io"
	"bufio"
	"os/exec"
	"fmt"
	"strings"

)

type ShellStatus int

const (
	MAX_POOL_SIZE = 100
	
	CREATED = iota
	STARTED
	RUNNING
	EXITED
	ERROR
	UNKOWN
	WAIT
)

func (s *ShellStatus)ToString()string{
	switch *s {
	case CREATED: return "created"
	case STARTED: return "started"
	case RUNNING: return "running"
	case EXITED: return "exited"
	case ERROR: return "error"
	}
	return "unkown"
}


type Shell struct {
	shell string
	Status ShellStatus
	PipLine chan string
	Pid int
}

func New()*Shell{
	return &Shell{
		shell: "/bin/bash",
		Status: UNKOWN,
		PipLine: make(chan string, MAX_POOL_SIZE),
	}
}

//Exec exec a shell cmd.
func(s *Shell)Exec(args... string)error{
	cmd :=  strings.Join(args, " ")
	if cmd == "" {
		return fmt.Errorf("cmd is empty")
	}
	if s.shell == ""{
		s.shell = "/bin/bash"
	}
	command := exec.Command(s.shell, "-c", cmd)
	s.Status = CREATED
	stdout, err := command.StdoutPipe()
	if err != nil {
		s.Status = ERROR
		return err
	}
	err = command.Start()
	s.Status = STARTED
	if err != nil {
		s.Status = ERROR
		return err
	}
	s.Pid = command.Process.Pid
	s.Status = RUNNING
	reader := bufio.NewReader(stdout)
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || err2 == io.EOF {
			if err2 != io.EOF {
				s.SendMsg(err2.Error())
			}
			break
		}
		s.SendMsg(line)
	}

	err = command.Wait()
	if err != nil || command.ProcessState.Success() == false {
		s.Status = ERROR
		return fmt.Errorf("command exec failed: %s", command.ProcessState.String())
	}
	s.Status = EXITED
	return err
}

//SendMsg send msg
func (s *Shell)SendMsg(msg string){
	if len(s.PipLine) == MAX_POOL_SIZE {
		<- s.PipLine
	}
	s.PipLine <- msg
}