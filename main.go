package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"unicode/utf8"
)

func main() {
	// listen on port 6379
	lis, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Fatal("something went wrong:", err)
		os.Exit(1)
	}

	defer lis.Close()

	log.Println("listening for connections on port 6379")

	// start accepting incoming connections
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		log.Println("got connection:", conn.RemoteAddr())
		go handleIncomingCommand(conn)
	}
}

type state int64

const (
	ReadArrayLen state = iota
	ReadBulkStringLen
	ReadBulkStringData
	Done
)

func handleIncomingCommand(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	scanner.Split(ScanCRLF)

	arrayLen := 0
	commandArgs := []string{}
	state := ReadArrayLen

	for state != Done {
		scanner.Scan()
		part := scanner.Bytes()

		switch state {
		case ReadArrayLen:
			data, _ := utf8.DecodeLastRune(part)

			arrLen, err := strconv.Atoi(string(data))
			if err != nil {
				returnError(conn, err.Error())
			}

			arrayLen = arrLen
			state = ReadBulkStringLen

		case ReadBulkStringLen:
			data, _ := utf8.DecodeLastRune(part)
			bulkStrLen, err := strconv.Atoi(string(data))
			if err != nil {
				returnError(conn, err.Error())
			}

			if bulkStrLen == -1 {
				commandArgs = append(commandArgs, "null")
				continue
			}

			state = ReadBulkStringData

		case ReadBulkStringData:
			commandArgs = append(commandArgs, string(part))
			if len(commandArgs) == arrayLen {
				state = Done
			} else {
				state = ReadBulkStringLen
			}
		}
	}

	if err := scanner.Err(); err != nil {
		returnError(conn, fmt.Sprintf("Invalid input: %s", err))
	}

	for i, cmd := range commandArgs {
		log.Println(cmd)
		if i == 0 {
			switch cmd {
			case "COMMAND":
				continue
			default:
				// startArg := ""
				// if arrayLen > i+1 {
				// 	startArg = commandArgs[i+1]
				// }
				// returnError(conn, fmt.Sprintf("unknown command '%s', with args beginning with: %s", cmd, startArg))
				returnOk(conn)
			}
		}
	}

	// close conn
	conn.Close()
}

func formatErr(msg string) []byte {
	return []byte(fmt.Sprintf("-ERR %s\r\n", msg))
}

func returnOk(conn net.Conn) {
	conn.Write([]byte("+OK\r\n"))
	conn.Close()
}

func returnError(conn net.Conn, msg string) {
	conn.Write(formatErr(msg))
	conn.Close()
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func ScanCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'\r', '\n'}); i >= 0 {
		// We have a full newline-terminated line.
		return i + 2, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}
