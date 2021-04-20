package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

const (
	COMMAND_TOKEN = "{}"
)

func findIndexToReplace(array []string, valueToReplace string) int {
	for i, val := range array {
		if val == valueToReplace {
			return i
		}
	}
	return -1
}

func run(rate, inflight int) {

	// obtain non-flag arguments
	args := flag.Args()
	lenArgs := len(args)
	if lenArgs == 0 {
		return
	}

	// find index of '{}' to replace with os.stdin
	indexToReplace := findIndexToReplace(args, COMMAND_TOKEN)
	if indexToReplace == -1 {
		log.Fatal("cannot find '{}'")
	}

	// create a semaphore to control the concurrency
	semaphore := make(chan struct{}, inflight)

	// create a chan to control the rate
	rateLimit := make(chan time.Time, rate) // rate*inflight
	for i := 0; i < rate; i++ {
		rateLimit <- time.Now()
	}
	go func() {
		for t := range time.Tick(time.Second / time.Duration(rate)) { // time.Duration(rate*inflight)
			rateLimit <- t
		}
	}()

	// create scanner for os.Stdin
	scanner := bufio.NewScanner(os.Stdin)

	var wg sync.WaitGroup

	for scanner.Scan() {
		// build command and its args to execute
		commandArgs := make([]string, lenArgs)
		copy(commandArgs, args)
		commandArgs[indexToReplace] = scanner.Text()

		wg.Add(1)
		go func() {
			defer wg.Done()

			// check the rate
			<-rateLimit

			// check the semaphore
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			// execute the command
			cmdOutput, err := exec.Command(commandArgs[0], commandArgs[1:]...).Output()
			if err != nil {
				os.Stdout.WriteString(err.Error())
				return
			}

			// write output to stdout
			if _, err := os.Stdout.Write(cmdOutput); err != nil {
				log.Println(err)
			}
		}()
	}
	wg.Wait()

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func validateFlagArgs(rate, inflight int) error {
	if rate < 1 {
		return errors.New("rate should be greater than 0")
	}
	if inflight < 1 {
		return errors.New("inflight should be greater than 0")
	}
	return nil
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "./ratelimit --rate <N> --inflight <P> <command...>")
		fmt.Fprintln(os.Stderr, "<command...>: the command to launch, '{}' is replaced with the string from stdin")
		flag.PrintDefaults()
	}
}
func main() {

	// init flags
	var flagRate, flagInflight int
	flag.IntVar(&flagRate, "rate", 1, "maximum number of command launches per second")
	flag.IntVar(&flagInflight, "inflight", 1, "maximum number of concurrently running commands")
	flag.Parse()

	// validate flag parameters
	if err := validateFlagArgs(flagRate, flagInflight); err != nil {
		log.Fatal(err)
	}

	run(flagRate, flagInflight)
}
