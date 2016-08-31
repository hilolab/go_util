package main

import (
	"encoding/binary"
	"fmt"
	"github.com/axgle/mahonia"
	//"bufio"
	"bytes"
	"io/ioutil"
	"net"
	"os"
)

//IP纯真库是little-endian字节序
//结构分为：文件头，记录区，索引区
//文件头：8个bytes，前4 bytes 是第一个IP索引，后4 bytes 是最后一个IP索引
//索引区：7个bytes，前4 bytes 是IP起始地址索引值，后 3 bytes 是指向记录区
//记录区：前4 bytes 是IP结束地址，第5个 bytes 有可能是模式值（0x01或0x02）
const (
	INDEX_LEN = 7    //每条索引是7个字节
	MODE1     = 0x01 //模式1
	MODE2     = 0x02 //模式2
)

type IPInfo struct {
	Data    []byte
	FristIp uint32
	LastIp  uint32
	TotalIp uint32
	Offset  uint32
	decode  mahonia.Decoder
	debug   bool
}

type Result struct {
	BeginIp string
	EndIp   string
	Country string
	Area    string
}

func main() {
	filename := "./qqwry.dat"
	ip := new(IPInfo)
	ip.Init(filename)

	//ip.FindIp("127.0.0.1")
	//ip.FindIp("8.8.8.8")
	result := ip.FindIp("202.96.128.86")
	fmt.Println(result)
	//ip.FindIp("192.168.1.2")
	//ip.FindIp("202.96.128.68")
}

func (this *IPInfo) Init(filename string) {
	fd, err := os.Open(filename)
	defer fd.Close()

	CheckErr(err)

	this.Data, err = ioutil.ReadAll(fd)
	CheckErr(err)
	buf := this.Data[:8]
	this.FristIp = binary.LittleEndian.Uint32(buf[:4])              //第一条IP索引
	this.LastIp = binary.LittleEndian.Uint32(buf[4:])               //最后一条IP索引
	this.TotalIp = uint32((this.LastIp - this.FristIp) / INDEX_LEN) //IP的总记录条数
	this.decode = mahonia.NewDecoder("GB18030")
}

func (this *IPInfo) readData(length uint32) []byte {
	buf := this.Data[this.Offset : this.Offset+length]
	this.Offset = this.Offset + length

	return buf
}

func (this *IPInfo) read4bytes(length uint32) []byte {
	buf := make([]byte, 4)
	copy(buf, this.readData(length))
	return buf
}

func (this *IPInfo) getLong(length uint32) uint32 {
	buf := this.read4bytes(length)
	return binary.LittleEndian.Uint32(buf)
}

func (this *IPInfo) FindIp(ipv4 string) Result {
	result := Result{}

	ip := binary.BigEndian.Uint32(net.ParseIP(ipv4).To4()) //转为大端序比较

	var start, end, mid, offset uint32
	end = this.TotalIp

	for start <= end {
		mid = (start + end) / 2
		offset = this.FristIp + mid*7

		this.Offset = offset
		beginIp := this.getLong(4)
		if ip < beginIp {
			end = mid + 1
		} else {
			this.Offset = this.getLong(3)
			endIp := this.getLong(4)
			if ip > endIp {
				start = mid
			} else {
				this.Offset = offset
				break
			}
		}
	}

	result.BeginIp = this.GetIp(this.read4bytes(4))
	offset = this.getLong(3)
	this.Offset = offset
	result.EndIp = this.GetIp(this.read4bytes(4))

	mode := this.getLong(1)
	var area, country string

	if mode == MODE1 {
		offset = this.getLong(3)
		this.Offset = offset
		mode = this.getLong(1)
		if mode == MODE2 {
			this.Offset = this.getLong(3)
			country = this.ReadString()
			this.Offset = offset + 4
			mode = this.getLong(1)

			if mode == MODE1 || mode == MODE2 {
				this.Offset = this.getLong(3)
				area = this.ReadString()
			} else {
				area = this.ReadString()
			}
		} else {
			this.Offset--
			country = this.ReadString()
			area = this.ReadString()
		}
	} else if mode == MODE2 {
		this.Offset = this.getLong(3)
		country = this.ReadString()
		this.Offset = offset + 8
		area = this.ReadString()
	} else {
		this.Offset--
		country = this.ReadString()
		area = this.ReadString()
	}

	result.Country = this.decode.ConvertString(country)
	result.Area = this.decode.ConvertString(area)

	return result
}

func (this *IPInfo) ReadString() string {
	ret := bytes.NewBuffer(nil)
	buf := make([]byte, 1)

	for {
		buf = this.readData(1)
		if buf[0] == 0 {
			break
		}

		ret.Write(buf)
	}

	return ret.String()
}

func (i *IPInfo) GetIp(buf []byte) string {
	var b [4]byte
	for k, v := range buf {
		b[k] = v
	}
	Ip := net.IPv4(b[3], b[2], b[1], b[0])
	return Ip.String()
}

func CheckErr(err error) {
	if err != nil {
		fmt.Println("error: ", err)
		os.Exit(-1)
	}
}
