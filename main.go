package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"
)

type Target struct {
	Database string
	Schema   string
	Table    string
	Column   string
}

type TargetedObfuscation struct {
	T Target
	O func(s []byte) []byte
}

func find(elements []string, one string) int {
	for i, s := range elements {
		if s == one {
			return i
		}
	}
	return -1
}

var fieldSeparator []byte = []byte("\t")

func processDataLine(config *Configuration, target *Target, columns []string, line []byte) []byte {
	fields := bytes.Split(line, fieldSeparator)
	if len(fields) != len(columns) {
		log.Println("Number of columns does not match number of data fields")
		return line
	}

	var value []byte
	for _, to := range config.Obfuscations {
		if to.T.Table != target.Table {
			continue
		}
		// TODO: try map
		columnIndex := find(columns, to.T.Column)
		if columnIndex == -1 {
			log.Println("Target column not found in earlier header. Wrong table?")
			return line
		}
		value = fields[columnIndex]
		if len(value) > 0 {
			fields[columnIndex] = to.O(value)
		}
	}

	line = bytes.Join(fields, fieldSeparator)
	return line
}

const (
	parseStateInvalid = iota
	parseStateOther
	parseStateCopy
)

func process(config *Configuration, input *bufio.Reader, output io.Writer) error {
	target := Target{}
	state := parseStateOther
	var columns []string

	// TODO: try map
	configuredTables := make([]string, 1)
	for _, to := range config.Obfuscations {
		if find(configuredTables, to.T.Table) == -1 {
			configuredTables = append(configuredTables, to.T.Table)
		}
	}

	for {
		line, readErr := input.ReadBytes('\n')
		if readErr != nil && readErr != io.EOF {
			panic("At ReadBytes")
		}
		if len(line) == 0 {
			goto next
		}

		switch state {
		case parseStateOther:
			if bytes.HasPrefix(line, []byte("COPY ")) {
				state = parseStateCopy
				lineString := string(line)
				tokens := strings.FieldsFunc(lineString, func(r rune) bool {
					return strings.ContainsRune(" \n'\"(),;", r)
				})
				if len(tokens) < 4 {
					return errors.New("process: parse error: too few tokens in COPY statement: " + string(line))
				}
				target.Table = tokens[1]
				columns = tokens[2 : len(tokens)-2]
			}
		case parseStateCopy:
			if bytes.Equal(line, []byte("\\.\n")) {
				state = parseStateOther
				columns = nil
				target = Target{}
			} else if find(configuredTables, target.Table) != -1 {
				// Data rows
				line = processDataLine(config, &target, columns, line)
			}
		}
		output.Write(line)

	next:
		if readErr == io.EOF {
			return nil
		}
	}
	return nil
}

func main() {
	inputPath := flag.String("input", "-", "Input filename, '-' for stdin")
	cpuprofile := flag.String("cpuprofile", "", "Write CPU profile to file")
	memprofile := flag.String("memprofile", "", "Write memory profile to file")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		defer f.Close()
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}
		go func() {
			for {
				time.Sleep(5 * time.Second)
				pprof.WriteHeapProfile(f)
			}
		}()
		defer pprof.WriteHeapProfile(f)
		defer f.Close()
	}

	sigIntChan := make(chan os.Signal, 1)
	signal.Notify(sigIntChan, syscall.SIGINT)
	go func() {
		<-sigIntChan
		os.Exit(1)
	}()

	// Initialize input reading
	log.Println("Reading from", *inputPath)
	var inputFile *os.File = os.Stdin
	if *inputPath != "-" {
		var err error
		inputFile, err = os.Open(*inputPath)
		if err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}
		defer inputFile.Close()
	}
	input := bufio.NewReader(inputFile)
	// TODO
	output := os.Stdout

	process(Config, input, output)
}
