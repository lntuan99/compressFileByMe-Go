package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

var wg sync.WaitGroup

const MAX_BYTE uint64 = 1024 * 1024

func readFile(filePath string) []uint64 {
	buff, err := ioutil.ReadFile(filePath)

	if err != nil {
		log.Println(err)
	}

	var freq = make([]uint64, 256)

	for i := range buff {
		freq[buff[i]]++
	}

	return freq
}

func compressFile() {
	var hmTree huffmanTree
	t1 := time.Now()

	buffi, err := ioutil.ReadFile(os.Args[1])

	if err != nil {
		log.Println(err)
	}

	var freq = make([]uint64, 256)

	for i := range buffi {
		freq[buffi[i]]++
	}

	t2 := time.Now()
	fmt.Println("read file ", t2.Sub(t1))

	t1 = time.Now()
	hmTree.buildMinHeap(freq)
	t2 = time.Now()
	fmt.Println("build min heap ", t2.Sub(t1))

	t1 = time.Now()
	hmTree.getAllCode()
	t2 = time.Now()
	fmt.Println("get all code: ", t2.Sub(t1))

	fo, err := os.Create(os.Args[2])
	defer fo.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	var lenFileName uint16

	lenFileName = uint16(len(os.Args[1]))

	t1 = time.Now()
	binary.Write(fo, binary.LittleEndian, lenFileName)
	fo.Write([]byte(os.Args[1]))
	t2 = time.Now()
	fmt.Println("write len file name and file name: ", t2.Sub(t1))

	var size uint64
	var padding uint16

	t1 = time.Now()
	for i := 0; i < 256; i++ {
		binary.Write(fo, binary.LittleEndian, freq[i])
		size += freq[i] * uint64(len(hmTree.codes[uint8(i)]))
	}
	t2 = time.Now()
	fmt.Println("write freq and calc size: ", t2.Sub(t1))

	for size%8 != 0 {
		size++
		padding++
	}

	t1 = time.Now()
	binary.Write(fo, binary.LittleEndian, size)
	binary.Write(fo, binary.LittleEndian, padding)
	t2 = time.Now()
	fmt.Println("write size and padding: ", t2.Sub(t1))

	var lenBuffi uint64 = uint64(len(buffi))
	var temp = make([]byte, 8)
	index := 0
	d := 0
	buffo := make([]byte, MAX_BYTE)

	t1 = time.Now()
	for lenBuffi >= MAX_BYTE {
		for idx := 0; idx < int(MAX_BYTE); idx++ {
			var c uint8 = buffi[idx]
			code := hmTree.codes[c]
			for i := 0; i < len(code); i++ {
				temp[index] = code[i]
				index++

				if index%8 == 0 {
					var ch uint8

					for j := 0; j < 8; j++ {
						if temp[j] == '1' {
							ch |= (128 >> j)
						}

					}

					buffo[d] = ch
					d++
					index = 0

					if d == int(MAX_BYTE) {
						fo.Write(buffo)
						d = 0
					}
				}
			}
		}
		lenBuffi -= MAX_BYTE
	}

	if lenBuffi > 0 && lenBuffi < MAX_BYTE {
		for idx := 0; uint64(idx) < lenBuffi; idx++ {
			var c uint8 = buffi[idx]

			code := hmTree.codes[c]
			for i := 0; i < len(code); i++ {
				temp[index] = code[i]
				index++

				if index%8 == 0 {
					var ch uint8

					for j := 0; j < 8; j++ {
						if temp[j] == '1' {
							ch |= (128 >> j)
						}
					}

					buffo[d] = ch
					d++
					index = 0

					if d == int(MAX_BYTE) {
						fo.Write(buffo[:d])
						d = 0
					}
				}
			}
		}
	}

	if padding != 0 {
		for i := 0; i < int(padding); i++ {
			temp[index] = 0
			index++
		}

		var ch byte

		for j := 0; j < 8; j++ {
			if temp[j] == '1' {
				ch |= (128 >> j)
			}
		}

		buffo[d] = ch
		d++

		if d == int(MAX_BYTE) {
			fo.Write(buffo[:d])
			d = 0
		}
	}

	if d > 0 {
		fo.Write(buffo[:d])
	}

	t2 = time.Now()
	fmt.Println("time compress: ", t2.Sub(t1))
}

func deCompressFile() {
	buff, _ := ioutil.ReadFile(os.Args[1])

	//convert bytes len file name to uint16
	lenFileName := binary.LittleEndian.Uint16(buff[:2])

	buff = buff[2:]

	//get file name bytes
	fileNameBytes := buff[:lenFileName]

	//convert to string
	fileName := string(fileNameBytes)

	buff = buff[lenFileName:]

	freq := make([]uint64, 256)
	for i := 0; i < 256; i++ {
		freq[i] = binary.LittleEndian.Uint64(buff[:8])
		buff = buff[8:]
	}

	fileDecode, err := os.Create(os.Args[2] + "\\" + fileName)
	defer fileDecode.Close()

	if err != nil {
		fmt.Println(err)
	}

	var size uint64
	var padding uint16

	size = binary.LittleEndian.Uint64(buff[:8])
	buff = buff[8:]

	padding = binary.LittleEndian.Uint16(buff[:2])
	buff = buff[2:]

	fmt.Println(size)
	var hmTree huffmanTree

	hmTree.buildMinHeap(freq)

	buffo := make([]byte, MAX_BYTE)

	var lenContent uint64 = uint64(len(buff))

	var i uint64
	curr := &hmTree.minHeap[0]

	for lenContent >= MAX_BYTE {
		var idx uint64
		for idx = 0; idx < MAX_BYTE; idx++ {
			for j := 0; j < 8; j++ {
				if (buff[idx]>>(7-j))&1 == 1 {
					curr = curr.right
				} else {
					curr = curr.left
				}

				if isLeaf(curr) {
					buffo[i] = curr.char
					i++

					if i == MAX_BYTE {
						fileDecode.Write(buffo)
						i = 0
					}

					curr = &hmTree.minHeap[0]
				}
			}
		}

		buff = buff[MAX_BYTE:]
		lenContent -= MAX_BYTE
	}

	if i > 0 {
		fileDecode.Write(buffo[:i])
	}

	i = 0

	if lenContent > 1 {
		var idx uint64
		for idx = 0; idx < lenContent-1; idx++ {
			for j := 0; j < 8; j++ {
				if (buff[idx]>>(7-j))&1 == 1 {
					curr = curr.right
				} else {
					curr = curr.left
				}

				if isLeaf(curr) {
					buffo[i] = curr.char
					i++

					if i == lenContent-1 {
						fileDecode.Write(buffo[:lenContent-1])
						i = 0
					}

					curr = &hmTree.minHeap[0]
				}
			}
		}
	}

	if i > 0 {
		fileDecode.Write(buffo[:i])
	}

	final := buff[lenContent-1]

	for j := 0; j < 8-int(padding); j++ {
		if (final>>(7-j))&1 == 1 {
			curr = curr.right
		} else {
			curr = curr.left
		}

		if isLeaf(curr) {
			buffo[0] = curr.char
			fileDecode.Write(buffo[:1])

			curr = &hmTree.minHeap[0]
		}
	}
}
