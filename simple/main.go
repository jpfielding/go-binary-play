package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"time"
)

/*
0       1       2       3       4       5       6       7
0123456701234567012345670123456701234567012345670123456701234567
+-------+-------+-------+-------+-------+-------+-------+------+
|    SensorID   |   LocationID  |            Timestamp         |
+-------+-------+-------+-------+-------+-------+-------+------+
|      Temp     |
+---------------+
*/

type packet struct {
	SensorID   uint16
	LocationID uint16
	Ts         uint32
	Temp       uint16
}

func main() {
	// for some id generation
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	// for some piping emulation
	pr, pw := io.Pipe()
	// populate the pipe with some packets
	go func() {
		defer pw.Close()
		for i := 0; i < 10; i++ {
			p := packet{
				SensorID:   uint16(rnd.Intn(3)),
				LocationID: uint16(rnd.Intn(10)),
				Ts:         uint32(time.Now().Unix()),
				Temp:       uint16(rnd.Intn(60) + 30),
			}
			err := binary.Write(pw, binary.BigEndian, p)
			if err != nil {
				fmt.Println(err)
				return
			}
			time.Sleep(time.Duration(rnd.Intn(100)+100) * time.Millisecond)
		}
	}()
	// read from our pipe until empty
	for {
		p := packet{}
		err := binary.Read(pr, binary.BigEndian, &p)
		if err != nil {
			fmt.Println("failed to Read:", err)
			return
		}
		fmt.Printf("%v\n", p)
	}
}
