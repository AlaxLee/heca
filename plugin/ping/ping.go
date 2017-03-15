package ping

import (
	"golang.org/x/net/icmp"
	"os"
	"fmt"
	"bytes"
	"net"
	"time"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"errors"
	"encoding/binary"
	"strconv"
	log "github.com/cihub/seelog"
	"github.com/spf13/viper"
)


type pingResult struct {
	available int //1 表示 true, 0 表示 false
	delay float64 //单位 秒
	loss float64
	err error
}

type MyPing struct {
	endpoint string
	interval int
	address  string
	timeout  int
	retry    int
	result   pingResult
}


func NewMyPing(config *viper.Viper) (*MyPing, error){
	endpoint := config.GetString("endpoint")
	interval := config.GetInt("jobInterval")
	address := config.GetString("address")
	timeout := config.GetInt("timeout")
	retry   := config.GetInt("retry")

	var err error
	if endpoint == "" {
		err = errors.New("lack endpoint")
	}
	if address == "" {
		err = errors.New("lack address")
	}
	if timeout <= 0 {
		err = errors.New("timeout must bigger than 0")
	}
	if retry < 0 {
		err = errors.New("retry must bigger than or equal to 0")
	}

	if err != nil {
		return nil, err
	}

	//fmt.Println(address, timeout, retry)

	return &MyPing{endpoint: endpoint, interval: interval, address: address, timeout: timeout, retry: retry, result: pingResult{}}, nil
}



//pingResult 的获取方式如下：
//	available: 连续 ping retry 次，只要有一次能 ping 通，就认为 ping 的通
//	err:       记录每个 ping 失败的原因
//	delay:     只有 ping 通的延迟参与到最终 delay 计算，单位 秒，0 表示都 ping 不通
//	loss:      ping 不通 除以 retry，值小于1，保留4位小数
func (p *MyPing) Do(resultChan chan<- interface{}) {

	var delay float64
	var num int
	var errString string

	p.result.available = 0
	p.result.delay = 0
	p.result.loss = 0
	p.result.err = nil

	for i := 0; i < p.retry; i++ {
		d, e := p.Ping()
		if d > 0 {
			delay += d
			num ++
		}
		if e != nil {
			errString += e.Error() + "\n"
		} else {
			p.result.available = 1
		}
	}

	if errString != "" {
		p.result.err = errors.New(errString)
	}

	if num > 0 {
		p.result.delay = delay/float64(num)
	}

	p.result.loss = float64((p.retry - num)) * 100/ float64(p.retry)
	p.result.loss = keepDecimalPlacesOnFloat64(p.result.loss, 4)

	p.result.delay = keepDecimalPlacesOnFloat64(p.result.delay, 3)

	log.Tracef("ping %s(%s):   available: %d, delay: %.3f, loss: %.2f, err: %s", p.endpoint, p.address, p.result.available, p.result.delay, p.result.loss, p.result.err)

	resultChan <- map[string]interface{} {
		"endpoint": p.endpoint,
		"metric": "ping.available",
		"timestamp": time.Now().Unix(),
		"step": p.interval,
		"value": p.result.available,
		"counterType": "GAUGE",
		"tags": "",
	}

	resultChan <- map[string]interface{} {
		"endpoint": p.endpoint,
		"metric": "ping.delay",
		"timestamp": time.Now().Unix(),
		"step": p.interval,
		"value": p.result.delay,
		"counterType": "GAUGE",
		"tags": "unit=second",
	}

	resultChan <- map[string]interface{} {
		"endpoint": p.endpoint,
		"metric": "ping.loss",
		"timestamp": time.Now().Unix(),
		"step": p.interval,
		"value": p.result.loss,
		"counterType": "GAUGE",
		"tags": "unit=percent",
	}

}



func (p *MyPing) Ping() (delay float64, err error) {
	c, err := net.Dial("ip4:icmp", p.address)
	if err != nil {
		err = errors.New("net.Dial failed: " + err.Error())
		return delay, err
	}
	c.SetDeadline(time.Now().Add(time.Duration(p.timeout) * time.Second))
	defer c.Close()

	typ := ipv4.ICMPTypeEcho
	xid, xseq := os.Getpid()&0xffff, 1
	//fmt.Println("xid: ", xid)


	timestampBuf := new(bytes.Buffer)
	err = binary.Write(timestampBuf, binary.BigEndian, time.Now().UnixNano())
	if err != nil {
		err = errors.New("create unixNano by BigEndian failed: " + err.Error())
		return delay, err
	}
	timestamp := timestampBuf.Bytes()[0:8]

	req := &icmp.Message{
		Type: typ, Code: 0,
		Body: &icmp.Echo{
			ID: xid, Seq: xseq,
			Data: timestamp,
		},
	}

	wb, err := req.Marshal(nil)
	if err != nil {
		err = errors.New("req.Marshal failed: " + err.Error())
		return delay, err
	}

	//dumpBytes(wb)

	if _, err = c.Write(wb); err != nil {
		err = errors.New("c.Write failed: " + err.Error())
		return delay, err
	}
	var m *icmp.Message
	rb := make([]byte, 20+len(wb)) //20 是 IP数据报的header的长度
	for {
		if _, err = c.Read(rb); err != nil {
			err = errors.New("c.Read failed: " + err.Error())
			return delay, err
		}

		rb = ipv4Payload(rb)

		//dumpBytes(rb)

		//icmp.ParseMessage的第一个参数这里是 iana.ProtocolICMP ，其值为1
		//之所以不直接写 iana.ProtocolICMP，是因为 iana 的包路径是 "golang.org/x/net/internal/iana"
		//貌似带internal的包不能直接被import，在build的时候会报错 "use of internal package not allowed"
		if m, err = icmp.ParseMessage(1, rb); err != nil {
			err = errors.New("icmp.ParseMessage failed: " + err.Error())
			return delay, err
		}
		switch m.Type {
		case ipv4.ICMPTypeEcho, ipv6.ICMPTypeEchoRequest:
			continue
		}
		break
	}

	//result, err := m.Body.Marshal(1)
	//if err != nil {
	//	err = errors.New("m.Body.Marshal failed: " + err.Error())
	//	return delay, err
	//}
	//dumpBytes(result)

	echo, ok := m.Body.(*icmp.Echo)
	if !ok {
		err = errors.New("m.Body is not *icmp.Echo")
		return delay, err
	}

	var i int64
	binary.Read(bytes.NewBuffer(echo.Data), binary.BigEndian, &i)
	delay = float64(time.Now().UnixNano() - i)/1000000000
	if delay <= 0 {
		delay = 0.000000001
	}

	return delay, nil
}

//处理获取的内容（IP数据报），第一个字节的低4位的值乘以4，得到ip header的长度
func ipv4Payload(b []byte) []byte {
	if len(b) < 20 {
		return b
	}
	hdrlen := int(b[0]&0x0f) << 2
	return b[hdrlen:]
}


/*
func dumpBytes(b []byte) {
	for i, a := range b {
		fmt.Printf("%02x ", a)
		if i % 8 == 7 {
			fmt.Print(" ")
		}
		if i % 16 == 15 {
			fmt.Print("\n")
		}
	}
	fmt.Print("\n")
}
*/


func keepDecimalPlacesOnFloat64(value float64, n uint) float64{
	format := fmt.Sprintf("%%.%df", n)
	result, err := strconv.ParseFloat( fmt.Sprintf(format, value), 64 )
	if err != nil {
		log.Warn(err)
		result = value
	}

	return result
}
