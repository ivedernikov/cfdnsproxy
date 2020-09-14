package main

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/miekg/dns"
)

func main() {
	var port int = 2000
	listenAndServe(port)

}

func listenAndServe(port int) {
	var (
		conn     net.Conn
		listener net.Listener
		err      error
	)
	if listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Cannot listen in tcp/%d, %v", port, err)
	}

	defer listener.Close()

	for {
		if conn, err = listener.Accept(); err == nil {
			go handleConnection(conn)
		} else {
			log.Fatalf("Got error while accepting connection, %v", err)
		}

	}

}

func handleConnection(conn net.Conn) {
	var replyMsg dns.Msg

	defer conn.Close()

	buf, err := getRawData(conn)
	if len(buf) == 0 || err != nil {
		return
	}

	origMsg := dns.Msg{}
	if err := origMsg.Unpack(buf); err != nil {
		log.Printf("Got error while unpacking message lenth %v", len(buf))

	}
	cfAnswer, err := getCloudflareAnswer(origMsg.Question)
	if err != nil {
		log.Printf("%v", err)
		replyMsg = createEmptyReplyMessage(origMsg)
	} else {
		log.Printf("Successfullty got answer len %v from CF", len(cfAnswer))
		replyMsg = createReplyMessageFromAnswer(origMsg, cfAnswer)
	}
	if n, err := writeReply(conn, replyMsg); err == nil {
		log.Printf("Finished with processing question for %v. Written %v bytes", origMsg.Question[0].Name, n)
	}
}

func createEmptyReplyMessage(origMsg dns.Msg) dns.Msg {
	replyMsg := dns.Msg{
		MsgHdr:   dns.MsgHdr{},
		Compress: false,
		Question: []dns.Question{},
		Answer:   []dns.RR{},
		Ns:       []dns.RR{},
		Extra:    []dns.RR{},
	}
	replyMsg.SetReply(&origMsg)
	return replyMsg
}

func createReplyMessageFromAnswer(origMsg dns.Msg, answer []dns.RR) dns.Msg {
	replyMsg := dns.Msg{
		MsgHdr:   dns.MsgHdr{},
		Compress: false,
		Question: []dns.Question{},
		Answer:   []dns.RR{},
		Ns:       []dns.RR{},
		Extra:    []dns.RR{},
	}
	replyMsg.SetReply(&origMsg)
	replyMsg.Answer = answer
	return replyMsg
}

func writeReply(conn net.Conn, replyMsg dns.Msg) (int, error) {
	packed, _ := replyMsg.Pack()

	//check if current connection is compatible with net.PacketConn interface
	if _, ok := conn.(net.PacketConn); ok {
		return conn.Write(packed)
	}
	//and if not write add empty 2 bytes in the end
	l := make([]byte, 2)
	binary.BigEndian.PutUint16(l, uint16(len(packed)))

	n, err := (&net.Buffers{l, packed}).WriteTo(conn)
	return int(n), err
}

func getRawData(conn net.Conn) ([]byte, error) {
	var length uint16
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	m := make([]byte, length)
	if _, err := io.ReadFull(conn, m); err != nil {
		return nil, err
	}

	return m, nil
}

func getCloudflareAnswer(questions []dns.Question) ([]dns.RR, error) {
	var (
		rwTimeout     time.Duration = 20 * time.Second
		cfMsg, retMsg *dns.Msg
	)
	cfMsg = new(dns.Msg)
	cfMsg.Id = uint16(rand.Intn(10000))
	cfMsg.MsgHdr.RecursionDesired = true
	cfMsg.Question = questions

	cfConn, err := dns.DialWithTLS("tcp-tls", "1.1.1.1:853", &tls.Config{})
	if err != nil {
		err = fmt.Errorf("Cannot connect to Cloudflare, error %v", err)
		return nil, err
	}
	defer cfConn.Close()

	// write with the appropriate write timeout
	cfConn.SetDeadline(time.Now().Add(rwTimeout))
	if err = cfConn.WriteMsg(cfMsg); err != nil {
		err = fmt.Errorf("Can't write massage into Cloudflare connection, %v", err)
		return nil, err
	}

	retMsg, err = cfConn.ReadMsg()
	if err == nil && retMsg.Id != cfMsg.Id {
		err = fmt.Errorf("Can't get proper answer from Cloudflare, err = %v, retmsgID = %v, initmsgID = %v", err, retMsg.Id, cfMsg.Id)
	}

	return retMsg.Answer, err
}
