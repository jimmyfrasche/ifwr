//Command ifwr(1) proxies a command and fails if it writes to stdout or stderr.
//ifwr(1) execs a command and notes if it writes to stdout or stderr.
//If the command completes successfully but wrote to stdout and ifwr(1) was given the -1 flag, it fails.
//If the command completes successfully but wrote to stderr and ifwr(1) was given the -2 flag, it fails.
//The output to stdout and stderr is still written and the program is run to completion.
//
//EXIT CODES
//
//If something goes wrong, the exit code is 254.
//
//If the program has a nonzero exit, that exit code will be returned.
//
//If the program has a zero exit code, and no writes were noted, the exit code is 0.
//
//If the program has a zero exit code, and writes were noted, the exit code is 255.
package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
)

const (
	fail    = 255
	unknown = 254
)

type didWrite struct {
	io.Writer
	wrote bool
	track *bool
}

func (d *didWrite) Write(p []byte) (int, error) {
	if len(p) > 0 {
		d.wrote = true
	}
	return d.Writer.Write(p)
}

func (d *didWrite) failed() bool {
	return *d.track && d.wrote
}

func exitCode(err *exec.ExitError) int {
	w, ok := err.Sys().(interface {
		ExitStatus() int
	})
	if !ok {
		return unknown
	}
	return w.ExitStatus()
}

func main() {
	log.SetFlags(0)

	trackStdout := flag.Bool("1", false, "fail if stdout written")
	trackStderr := flag.Bool("2", false, "fail if stderr written")
	flag.Usage = func() {
		log.Printf("%s: [flags] cmd args*", os.Args[0])
		flag.PrintDefaults()
		log.Println("Both -1 and -2 may be set. If neither are specified, -2 is set implicitly.")
	}
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Println("No command given")
		flag.Usage()
		os.Exit(2)
	}
	if !*trackStdout && !*trackStderr {
		*trackStderr = true
	}

	Stdout := &didWrite{Writer: os.Stdout, track: trackStdout}
	Stderr := &didWrite{Writer: os.Stderr, track: trackStderr}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr

	if err := cmd.Run(); err != nil {
		//program failed, propagate exit code
		if exiterr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitCode(exiterr))
		}
		//otherwise just pass the error along
		log.Println(err)
		os.Exit(unknown)
	}

	if Stdout.failed() || Stderr.failed() {
		os.Exit(fail)
	}
}
