package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		panic(fmt.Errorf("program should accept exactly one arg, passed %d arguments", len(os.Args)))
	}
	profile := os.Args[1]
	fmt.Printf("GAC: Working with gimme-aws-creds profile: %s\n", profile)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	awsCredsFile := homeDir + "/.aws/credentials"
	profileFile := homeDir + "/.gac/" + profile + ".txt"

	if _, err := os.Stat(profileFile); os.IsNotExist(err) {
		fmt.Println("GAC: profile file not found")
		fmt.Println("GAC: using gimme-aws-creds to obtain new creds")
		callGimmeAwsCreds(profile)
		fmt.Println("GAC: saving aws creds")
		copyFile(awsCredsFile, profileFile)
		return
	} else if err != nil {
		panic(err)
	}

	t := credsExpiration(profileFile)
	if time.Now().Before(t) {
		fmt.Println("GAC: reusing creds valid until", t)
		copyFile(profileFile, awsCredsFile)
		return
	}

	fmt.Println("GAC: creds expired", t)
	fmt.Println("GAC: using gimme-aws-creds to obtain new creds")

	callGimmeAwsCreds(profile)
	fmt.Println("GAC: saving aws creds")
	copyFile(awsCredsFile, profileFile)
}

func callGimmeAwsCreds(profile string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gimme-aws-creds", "-p", profile)
	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	doneStdOut := make(chan struct{})
	go func() {
		defer close(doneStdOut)

		for {
			p := make([]byte, 2048)
			n, err := stdOut.Read(p)
			if n > 0 {
				fmt.Print(string(p[:n]))
			}
			if err != nil {
				if err == io.EOF {
					return
				}
				fmt.Println(fmt.Errorf("reading stdout: %w", err))
				return
			}
		}
	}()

	stdErr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	doneStdErr := make(chan struct{})
	go func() {
		defer close(doneStdErr)

		for {
			p := make([]byte, 2048)
			n, err := stdErr.Read(p)
			if n > 0 {
				fmt.Print(string(p[:n]))
			}
			if err != nil {
				if err == io.EOF {
					return
				}
				fmt.Println(fmt.Errorf("reading stderr: %w", err))
				return
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	<-doneStdOut
	<-doneStdErr

	err = cmd.Wait()
	if err != nil {
		panic(err)
	}
}

func readFile(f string) []byte {
	b, err := os.ReadFile(f)
	if err != nil {
		panic(err)
	}
	return b
}

func copyFile(src, dst string) {
	b := readFile(src)

	MkdirAllFromFile(dst)
	err := os.WriteFile(dst, b, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func credsExpiration(credsFile string) time.Time {
	b := readFile(credsFile)

	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		lParts := strings.Fields(line)
		if len(lParts) == 3 && lParts[0] == "x_security_token_expires" {
			const layout = "2006-01-02T15:04:05-07:00"
			t, err := time.Parse(layout, lParts[2])
			if err != nil {
				panic(err)
			}
			return t
		}
	}
	panic(fmt.Errorf("did not find x_security_token_expires in creds file, credsFile: %s", credsFile))
}

func MkdirAll(dirPath string) {
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func MkdirAllFromFile(filePath string) {
	idx := strings.LastIndex(filePath, "/")
	MkdirAll(filePath[:idx])
}
