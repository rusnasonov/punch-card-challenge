package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	errorLog := log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	rand.Seed(time.Now().UnixNano())
	listener, err := net.Listen("tcp", "[::]:2019")
	if err != nil {
		errorLog.Fatal(err)
	}
	defer listener.Close()
	infoLog.Println("Listening on [::]:2019")
	for {
		conn, err := listener.Accept()
		if err != nil {
			errorLog.Println(err)
			continue
		}
		go handleRequest(conn, infoLog, errorLog)
	}
}

func handleRequest(conn net.Conn, infoLog, errorLog *log.Logger) {
	defer conn.Close()
	ooops := []byte("OOOooooops :(\n")
	ooopsSlow := []byte("OOOooooops, so slooow :(\n")
	firstMsg := "THIS IS TRUE WAY - TRY HARDER"
	final := "THIS IS KEY"
	maxAttempts := 100
	timeout := time.Second * 3
	buf := make([]byte, 255)
	infoLog.Printf("Handle connection from %s\n", conn.RemoteAddr().String())
	_, err := conn.Write([]byte("Send 'start' if you are ready\n"))
	if err != nil {
		errorLog.Println(err)
		return
	}
	n, err := conn.Read(buf)
	if err != nil {
		errorLog.Println(err)
		return
	}
	if strings.Trim(string(buf[:n]), "\n") != "start" {
		infoLog.Printf("%s Fail on start\n", conn.RemoteAddr().String())
		conn.Write(ooops)
		return
	}
	card, err := makePunchCard(firstMsg)
	if err != nil {
		errorLog.Println(err)
		conn.Write(ooops)
		return
	}
	if err = writeCard(conn, card); err != nil {
		errorLog.Println(err)
		return
	}
	n, err = conn.Read(buf)
	if err != nil {
		errorLog.Println(err)
		return
	}
	if strings.Trim(string(buf[:n]), "\n") != firstMsg {
		infoLog.Printf("%s Fail to decode first message\n", conn.RemoteAddr().String())
		conn.Write(ooops)
		return
	}
	attempts := 0
	for {
		hash := strings.ToUpper(
			getMD5Hash(
				strconv.Itoa(
					rand.Intn(10000000),
				),
			),
		)
		card, err = makePunchCard(hash)
		if err != nil {
			errorLog.Println(err)
			conn.Write(ooops)
			return
		}
		err = writeCard(conn, card)
		if err != nil {
			errorLog.Println(err)
			return
		}
		start := time.Now()
		n, err = conn.Read(buf)
		end := time.Now()
		if end.Sub(start) > timeout {
			infoLog.Printf("%s Answer timeout\n", conn.RemoteAddr().String())
			conn.Write(ooopsSlow)
			return
		}
		if strings.Trim(string(buf[:n]), "\n") != hash {
			infoLog.Printf("%s Fail to decode message. Attempt %d, Exprected: %s, Got: %s\n",
				conn.RemoteAddr().String(),
				attempts,
				hash,
				string(buf[:n]),
			)
			conn.Write(ooops)
			return
		}
		attempts++
		if attempts == maxAttempts {
			card, err := makePunchCard(final)
			if err != nil {
				errorLog.Println(err)
				conn.Write(ooops)
				return
			}
			infoLog.Printf("%s Finish\n", conn.RemoteAddr().String())
			err = writeCard(conn, card)
			if err != nil {
				errorLog.Println(err)
				return
			}
			return
		}
	}
}

func writeCard(conn net.Conn, card [][]byte) error {
	_, err := conn.Write(bytes.Join(card, []byte("\n")))
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte("\n"))
	if err != nil {
		return err
	}
	return nil
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func makePunchCard(value string) ([][]byte, error) {
	buf := make([][]byte, 12)
	if len(value) > 80 {
		return buf, fmt.Errorf("Value lenght must less than 80 chars")
	}
	// initialize
	for i := range buf {
		buf[i] = make([]byte, 80)
		for j := range buf[i] {
			buf[i][j] = byte('#')
		}
	}
	encoded, err := punchCardEncoder(value)
	if err != nil {
		return buf, err
	}
	for i := 0; i < 12; i++ {
		for j, r := range encoded {
			buf[i][j] = r[i]
		}
	}
	return buf, nil
}

func punchCardEncoder(value string) ([][]byte, error) {
	buf := [][]byte{}
	encodingTable := map[rune][]byte{
		' ': []byte("############"),
		'&': []byte("X###########"),
		'-': []byte("#X##########"),
		'0': []byte("##X#########"),
		'1': []byte("###X########"),
		'2': []byte("####X#######"),
		'3': []byte("#####X######"),
		'4': []byte("######X#####"),
		'5': []byte("#######X####"),
		'6': []byte("########X###"),
		'7': []byte("#########X##"),
		'8': []byte("##########X#"),
		'9': []byte("###########X"),
		'A': []byte("X##X########"),
		'B': []byte("X###X#######"),
		'C': []byte("X####X######"),
		'D': []byte("X#####X#####"),
		'E': []byte("X######X####"),
		'F': []byte("X#######X###"),
		'G': []byte("X########X##"),
		'H': []byte("X#########X#"),
		'I': []byte("X##########X"),
		'J': []byte("#X#X########"),
		'K': []byte("#X##X#######"),
		'L': []byte("#X###X######"),
		'M': []byte("#X####X#####"),
		'N': []byte("#X#####X####"),
		'O': []byte("#X######X###"),
		'P': []byte("#X#######X##"),
		'Q': []byte("#X########X#"),
		'R': []byte("#X#########X"),
		'/': []byte("##XX########"),
		'S': []byte("##X#X#######"),
		'T': []byte("##X##X######"),
		'U': []byte("##X###X#####"),
		'V': []byte("##X####X####"),
		'W': []byte("##X#####X###"),
		'X': []byte("##X######X##"),
		'Y': []byte("##X#######X#"),
		'Z': []byte("##X########X"),
	}
	for _, char := range value {
		s, exists := encodingTable[char]
		if !exists {
			return buf, fmt.Errorf("Can't encode char '%v'", char)
		}
		buf = append(buf, s)
	}
	return buf, nil
}
