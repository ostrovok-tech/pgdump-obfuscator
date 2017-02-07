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
	"fmt"
)

type configFlags []string

func (self *configFlags) String() string {
	return strings.Join(*self, ", ")
}

func (self *configFlags) Set(value string) error {
	*self = append(*self, value)
	return nil
}

func (self *configFlags) ToConfiguration() (*Configuration, error) {
	configuration := &Configuration{}
	for _, v := range *self {
		splittedValues := strings.Split(v, ":")
		if len(splittedValues) == 3 {
			table, column, name := splittedValues[0], splittedValues[1], splittedValues[2]
			scrambler, err := GetScrambleByName(name)
			if err != nil {
				return configuration, err
			}
			configuration.Obfuscations = append(
				configuration.Obfuscations,
				TargetedObfuscation{
					Target{Table: table, Column: column},
					scrambler,
				},
			)
		} else {
			return nil, errors.New(fmt.Sprintf("Inccorrect data in configuration flags!\n"))
		}
	}
	return configuration, nil
}

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

var fieldSeparator = []byte("\t")

func processDataLine(config *Configuration, target *Target, columns []string, line *[]byte) error {
	fields := bytes.Split(*line, fieldSeparator)
	if len(fields) != len(columns) {
		return errors.New("Number of columns does not match number of data fields")
	}

	var value []byte
	for _, to := range config.Obfuscations {
		if to.T.Table != target.Table {
			continue
		}
		// TODO: try map
		columnIndex := find(columns, to.T.Column)
		if columnIndex == -1 {
			return errors.New("Target column not found in earlier header. Wrong table?")
		}
		value = fields[columnIndex]
		if len(value) == 0 {
			continue
		} else if len(value) == 2 && value[0] == '\\' && value[1] == 'N' {
			continue
		} else {
			fields[columnIndex] = to.O(value)
		}
	}

	*line = bytes.Join(fields, fieldSeparator)
	return nil
}

const (
	parseStateInvalid = iota
	parseStateOther
	parseStateCopy
)

var bytesCopyBegin = []byte("COPY ")
var bytesCopyEnd = []byte("\\.\n")
var bytesNewline = []byte("\n")

const copySyntaxDelimiters = " \n'\"(),;"

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

	var currentLineNumber uint64
	var line []byte
	defer func() {
		if r := recover(); r != nil {
			log.Fatalln("Line", currentLineNumber, "error:", r)
		}
	}()

	var err, readErr error
	for currentLineNumber = 0; ; currentLineNumber++ {
		line, readErr = input.ReadBytes('\n')
		if readErr != nil && readErr != io.EOF {
			panic("At ReadBytes")
		}
		if len(line) == 0 {
			goto next
		}

		switch state {
		case parseStateOther:
			if bytes.HasPrefix(line, bytesCopyBegin) {
				state = parseStateCopy
				lineString := string(line)
				tokens := strings.FieldsFunc(lineString, func(r rune) bool {
					return strings.ContainsRune(copySyntaxDelimiters, r)
				})
				if len(tokens) < 4 {
					return errors.New("process: parse error: too few tokens in COPY statement: " + string(line))
				}
				target.Table = tokens[1]
				columns = tokens[2 : len(tokens)-2]
			}
		case parseStateCopy:
			if bytes.Equal(line, bytesCopyEnd) {
				state = parseStateOther
				columns = nil
				target = Target{}
			} else if find(configuredTables, target.Table) != -1 {
				// Data rows
				hasNewlineSuffix := bytes.HasSuffix(line, bytesNewline)
				if hasNewlineSuffix {
					line = line[:len(line)-1]
				}
				err = processDataLine(config, &target, columns, &line)
				if err != nil {
					log.Println("process: line", currentLineNumber, "error:", err)
				} else if hasNewlineSuffix {
					line = append(line, '\n')
				}
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
	var configs configFlags
	inputPath := flag.String("input", "-", "Input filename, '-' for stdin")
	cpuprofile := flag.String("cpuprofile", "", "Write CPU profile to file")
	memprofile := flag.String("memprofile", "", "Write memory profile to file")
	flag.Var(&configs, "c", "Configs, example: auth_user:email:email, auth_user:password:bytes")
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

	configuration, err := configs.ToConfiguration()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	process(configuration, input, output)
}
