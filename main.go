package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func challenge(host, port string, infoLog, errorLog *log.Logger) {
	rand.Seed(time.Now().UnixNano())
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))
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

func solve(host, port string, infoLog, errorLog *log.Logger) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	buf := make([]byte, 1024)
	if err != nil {
		errorLog.Fatal(err)
	}
	_, err = bufio.NewReader(conn).Read(buf)
	fmt.Fprint(conn, "start")
	for {
		n, err := bufio.NewReader(conn).Read(buf)
		if err == io.EOF {
			infoLog.Println("== FINISH ==")
			return
		}
		if err != nil {
			errorLog.Fatal(err)
		}
		infoLog.Println(string(buf))
		card := bytes.Split(
			bytes.Trim(buf[:n], "\n"),
			[]byte("\n"),
		)
		decoded, err := punchCardDecoder(card)
		if err != nil {
			errorLog.Fatal(err)
		}
		fmt.Fprint(conn, strings.Trim(decoded, " "))
		infoLog.Println(decoded)
		time.Sleep(time.Millisecond * 100)
	}
}

func main() {
	errorLog := log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	isSolve := flag.Bool("solve", false, "run solver")
	host := flag.String("host", "[::]", "host")
	port := flag.String("port", "2019", "port")
	flag.Parse()
	if *isSolve {
		solve(*host, *port, infoLog, errorLog)
	} else {
		challenge(*host, *port, infoLog, errorLog)
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
			infoLog.Printf("%s Fail to decode message. Attempt %d, Expected: %s, Got: %s\n",
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
	_, err := conn.Write(
		append(bytes.Join(card, []byte("\n")), []byte("\n")...),
	)
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
	encodingTable := map[rune]string{
		' ': "############",
		'&': "X###########",
		'-': "#X##########",
		'0': "##X#########",
		'1': "###X########",
		'2': "####X#######",
		'3': "#####X######",
		'4': "######X#####",
		'5': "#######X####",
		'6': "########X###",
		'7': "#########X##",
		'8': "##########X#",
		'9': "###########X",
		'A': "X##X########",
		'B': "X###X#######",
		'C': "X####X######",
		'D': "X#####X#####",
		'E': "X######X####",
		'F': "X#######X###",
		'G': "X########X##",
		'H': "X#########X#",
		'I': "X##########X",
		'J': "#X#X########",
		'K': "#X##X#######",
		'L': "#X###X######",
		'M': "#X####X#####",
		'N': "#X#####X####",
		'O': "#X######X###",
		'P': "#X#######X##",
		'Q': "#X########X#",
		'R': "#X#########X",
		'/': "##XX########",
		'S': "##X#X#######",
		'T': "##X##X######",
		'U': "##X###X#####",
		'V': "##X####X####",
		'W': "##X#####X###",
		'X': "##X######X##",
		'Y': "##X#######X#",
		'Z': "##X########X",
	}
	for _, char := range value {
		s, exists := encodingTable[char]
		if !exists {
			return buf, fmt.Errorf("Can't encode char '%v'", char)
		}
		buf = append(buf, []byte(s))
	}
	return buf, nil
}

func punchCardDecoder(card [][]byte) (string, error) {
	decoded := ""
	decodingTable := map[string]rune{
		"############": ' ',
		"X###########": '&',
		"#X##########": '-',
		"##X#########": '0',
		"###X########": '1',
		"####X#######": '2',
		"#####X######": '3',
		"######X#####": '4',
		"#######X####": '5',
		"########X###": '6',
		"#########X##": '7',
		"##########X#": '8',
		"###########X": '9',
		"X##X########": 'A',
		"X###X#######": 'B',
		"X####X######": 'C',
		"X#####X#####": 'D',
		"X######X####": 'E',
		"X#######X###": 'F',
		"X########X##": 'G',
		"X#########X#": 'H',
		"X##########X": 'I',
		"#X#X########": 'J',
		"#X##X#######": 'K',
		"#X###X######": 'L',
		"#X####X#####": 'M',
		"#X#####X####": 'N',
		"#X######X###": 'O',
		"#X#######X##": 'P',
		"#X########X#": 'Q',
		"#X#########X": 'R',
		"##XX########": '/',
		"##X#X#######": 'S',
		"##X##X######": 'T',
		"##X###X#####": 'U',
		"##X####X####": 'V',
		"##X#####X###": 'W',
		"##X######X##": 'X',
		"##X#######X#": 'Y',
		"##X########X": 'Z',
	}
	if len(card) != 12 {
		return "", fmt.Errorf("Wrong row count. Got %d, expected 12", len(card))
	}
	for row := 0; row < 12; row++ {
		if len(card[row]) != 80 {
			return "", fmt.Errorf("Wrong column count [row %d]. Got %d, expected 80", row, len(card))
		}
	}
	for column := 0; column < 80; column++ {
		buf := []byte{}
		for row := 0; row < 12; row++ {
			buf = append(buf, card[row][column])
		}
		s, exists := decodingTable[string(buf)]
		if !exists {
			return "", fmt.Errorf("Can't decode column %s", string(buf))
		}
		decoded += string(s)
	}
	return decoded, nil
}
