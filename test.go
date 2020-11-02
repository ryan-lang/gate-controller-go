package main

import (
    "encoding/hex"
    "fmt"
    "github.com/tarm/serial"
    "log"
    "time"
)

func main() {
    fmt.Println("running serial test on ttyUSB0")

    config := &serial.Config{
        Name:        "/dev/ttyUSB0",
        Baud:        38400,
        ReadTimeout: time.Millisecond * 500,
    }

    var gateID byte
    var msgType byte
    gateID = 1
    msgType = 'V'

    gateVerCmd := buildGateCommand(gateID, msgType, nil)
    fmt.Println(hex.EncodeToString(gateVerCmd[:]))

    stream, err := serial.OpenPort(config)
    if err != nil {
        log.Fatal(err)
    }

    // read buffer
    rbuf := make([]byte, 512)

    // working buffer
    var wbuf []byte

    awaitingResp := false

    for {

        fmt.Print("<- reading stream...")
        n, _ := stream.Read(rbuf)

        if n == 0 && !awaitingResp {
            fmt.Println("-> writing command")
            stream.Write(gateVerCmd)
            awaitingResp = true
        }
        if n > 0 {
            s := hex.EncodeToString(rbuf[:n])
            fmt.Println(s)

            // concat the received bytes to the working buffer
            wbuf = append(wbuf, rbuf[:n]...)

            s2 := hex.EncodeToString(wbuf[:])
            fmt.Printf("wbuf: %s \n", s2)

            msg := parseMessage(&wbuf)

            s3 := hex.EncodeToString(wbuf[:])
            fmt.Printf("wbuf after msg: %s \n", s3)

            if len(msg) > 0 {
                s := string(msg[:])
                fmt.Printf("msg: %s \n---------------------------------\n", s)
                awaitingResp = false
            }
        }

    }
}

const wantFF = 1
const wantID = 2
const wantMsgSize = 3
const wantMsg = 4
const wantChecksum = 5

func parseMessage(bufptr *[]byte) []byte {

    buf := *bufptr

    //var stcID byte
    i := 0
    max_i := len(buf) - 1

    var msg_size int
    msg_i := 0
    var msg []byte

    want := wantFF
    byteSum := 0

    for ; i <= max_i; i++ {
        switch want {

        case wantFF:
            if buf[i] == 0xff {
                want++
            }

        case wantID:
            //stcID = buf[i]
            want++

        case wantMsgSize:
            msg_size = int(buf[i])
            want++

        case wantMsg:
            msg = append(msg, buf[i])
            msg_i++
            if msg_i >= msg_size {
                want++
            }

        case wantChecksum:
            byteSum += int(buf[i])
            fmt.Printf("checksum total: %d \n", byteSum)

            if byteSum%256 == 0 {
                // all is good. Truncate the buffer and return the message
                *bufptr = buf[i+1:]
                return msg
            } else {
                fmt.Printf("!!! byteSum (%d) fails checksum (%d) \n", byteSum, byteSum%256)
            }

        }

        // running total for checksum eval
        if want > wantFF {
            byteSum += int(buf[i])
        }
    }

    // if here we didn't find a fully formed message
    return nil
}

func buildGateCommand(rs485Address byte, msgType byte, msg []byte) []byte {
    // start symbol (0xFF) + Rs485 address of gate + message length + message type + message (optional) + checksum
    cmd := []byte{0xff, rs485Address, byte(len(msg) + 1), msgType}

    if len(msg) > 0 {
        cmd = append(cmd, msg...)
    }

    // calc checksum
    var sum int
    for i := range cmd {
        sum += int(cmd[i])
        fmt.Printf("val %d, sum %d \n", cmd[i], sum)
    }
    fmt.Printf("sum %d \n", sum)

    sum %= 256
    fmt.Printf("modulus %d \n", sum)

    sum = ^sum
    fmt.Printf("complement %d \n", sum)

    sum += 1
    fmt.Printf("checksum %d \n", sum)

    cmd = append(cmd, byte(sum))

    return cmd
}
