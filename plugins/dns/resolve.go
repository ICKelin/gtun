package dns

import (
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gone/cache"
)

type Resolver struct {
	clientCache *cache.Cache
}

type LocalRecord struct {
	ExpiredIn int64
	Msg       *dns.Msg
}

func NewLocalRecord(msg *dns.Msg) *LocalRecord {
	localRecord := &LocalRecord{}
	localRecord.Msg = msg

	// 无应答记录，设置过期值比当前时间戳小
	if len(msg.Answer) <= 0 {
		localRecord.ExpiredIn = time.Now().Unix() - 1
	} else {
		localRecord.ExpiredIn = time.Now().Unix() + int64(msg.Answer[0].Header().Ttl)
	}
	return localRecord
}

// 缓存是否过期
func (this *LocalRecord) Expired() bool {
	return time.Now().Unix() > this.ExpiredIn
}

func NewResolver() *Resolver {
	return &Resolver{
		clientCache: cache.NewCache(cache.ALGO_HASH),
	}
}

func (this *Resolver) Resolve(query *dns.Msg, srv string) (*dns.Msg, error) {
	conn, err := net.Dial("udp", srv+":53")
	if err != nil {
		return nil, err
	}

	begin := int64(0)
	if GetConfig().IsDebug() {
		begin = time.Now().UnixNano()
	}

	question := GetDNSQuestions(query)
	if GetConfig().IsCacheOn() {
		if ele, err := this.clientCache.Get(question); err == nil {
			if localRecord, ok := ele.(*LocalRecord); ok {
				if GetConfig().IsDebug() {
					fmt.Printf("resolve\t%-50s success %10dms upper:%s\n", GetDNSQuestions(query), (time.Now().UnixNano()-begin)/1000/1000, "local cache")
				}
				// 校验TTL值
				if !localRecord.Expired() {
					response := query.Copy()
					response.Answer = localRecord.Msg.Answer
					return response, nil
				} else {
					this.clientCache.Del(question)
				}
			}
		}
	}

	response, err := this.resolve(query, conn)
	if err != nil {
		if GetConfig().IsDebug() {
			fmt.Printf("resolve\t%-50s fail %10dms upper:%s\n", question, (time.Now().UnixNano()-begin)/1000/1000, conn.RemoteAddr().String())
		}
		return nil, err
	}

	if GetConfig().IsDebug() {
		fmt.Printf("resolve\t%-50s success %10dms upper:%s\n", question, (time.Now().UnixNano()-begin)/1000/1000, conn.RemoteAddr().String())
	}

	if GetConfig().IsCacheOn() {
		question = GetDNSQuestions(response)
		localRecord := NewLocalRecord(response)
		this.clientCache.Add(question, localRecord)
	}

	return response, err
}

func (this *Resolver) resolve(query *dns.Msg, conn net.Conn) (*dns.Msg, error) {
	data, err := query.Pack()
	if err != nil {
		return nil, err
	}

	response, err := this.resolveUDP(conn, data)
	if err != nil {
		return nil, err
	}

	// 是否有ns记录,ns记录优先级高于A记录
	if HasNsRecord(response) {
		return HandleNSResponse(response, conn)
	}

	// 是否有A记录,a记录优先级高于cname记录
	if HasARecord(response) {
		return response, nil
	}

	if HasAAAARecord(response) {
		glog.INFO(response.Answer)
		return response, nil
	}

	if HasPTRRecord(response) {
		return response, nil
	}

	// 是否有cname记录
	if HasCNameRecord(response) {
		return HandleCNameResponse(response, conn)
	}
	return response, nil
}

func (this *Resolver) resolveUDP(conn net.Conn, data []byte) (*dns.Msg, error) {
	conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	if _, err := conn.Write(data); err != nil {
		return nil, fmt.Errorf("write to upper server %s error: %s", conn.RemoteAddr().String(), err.Error())
	}
	conn.SetWriteDeadline(time.Time{})

	buff := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	n, err := conn.Read(buff)
	if err != nil {
		return nil, fmt.Errorf("read from upper server %s error %s", conn.RemoteAddr().String(), err.Error())
	}
	conn.SetReadDeadline(time.Time{})

	msg := &dns.Msg{}
	err = msg.Unpack(buff[:n])

	return msg, err
}
