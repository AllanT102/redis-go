package main

import (
	"fmt"
	// Uncomment this block to pass the first stage
	"net"
	"os"
	"strings"
	"bufio"
	"strconv"
	"errors"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

// handleConnection reads commands from the client, processes them and sends responses.
func handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		args, err := parseRESPCommand(scanner)
		if err != nil {
			fmt.Fprintf(conn, "-ERR %s\r\n", err)
			continue
		}

		switch strings.ToUpper(args[0]) {
		case "ECHO":
			fmt.Fprintf(conn, buildBulkString(args[1:]))
			break
		case "PING":
			fmt.Fprintf(conn, buildBulkString([]string{"PONG"}))
			break
		default:
			fmt.Fprintf(conn, "-ERR unknown command '%s'\r\n", args[0])
		}
	}
}
		
// ["type of array", "arg_1", arg2 ...]
// *2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n
// assumes single type arrays
// parseRESPCommand parses a RESP encoded command string and returns a slice of its components.
func parseRESPCommand(scanner *bufio.Scanner) ([]string, error) {
	command := scanner.Text()
	if !strings.HasPrefix(command, "*") {
		return nil, fmt.Errorf("expected array format")
	}

	numArgs, err := strconv.Atoi(command[1:])
	if err != nil {
		return nil, errors.New("Error converting string to int")
	}
	var results []string
	for numArgs > 0 && scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == '*' || line[0] == '$' {
			continue
		}
		results = append(results, line)
		fmt.Println(line)
		numArgs -= 1
	}
	return results, nil
}

func buildBulkString(data []string) string {
	bulkString := ""
	for _, arg := range data {
		bulkString += fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
	}
	return bulkString
} 

// $4\r\nECHO -> [$4, ECHO] -> doesn't account for ECHO\r\n being a string
func parseBulkString(parts []string, currPart int) string {
	return parts[currPart + 1]
}
