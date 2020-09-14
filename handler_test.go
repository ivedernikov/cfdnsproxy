package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"testing"

	"github.com/miekg/dns"
)

func TestGetRawData(t *testing.T) {
	message := "Hi there!\n"

	l, err := net.Listen("tcp", ":2000")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go func() {
		conn, err := net.Dial("tcp", ":2000")
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		conn.Write([]byte(message))
	}()
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf, err := ioutil.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}

		if msg := string(buf[:]); msg != message {
			t.Fatalf("Unexpected message:\nGot:\t\t%s\nExpected:\t%s\n", msg, message)
		}

		if len([]byte(message)) != len(buf) {
			t.Fatalf("Unexpected length of received byte slice. Got %v, Expected %v", len(buf), len([]byte(message)))
		}
		return // Done
	}

}

func TestCloudflareAnswer(t *testing.T) {
	questions := []dns.Question{
		{
			Name:   "example.com.",
			Qclass: 1,
			Qtype:  1,
		},
	}
	indirectRR, err := getCloudflareAnswer(questions)
	if err != nil {
		t.Errorf("Got error from indirect Cloudflare, %v", err)
	}
	actual := strings.Split(fmt.Sprintf("%v", indirectRR[0]), "\t")[4]
	fmt.Printf("DEBUG: %v", actual)
	client := new(dns.Client)
	msg := new(dns.Msg)
	msg.Question = questions
	directResponse, _, err := client.Exchange(msg, "1.1.1.1:53")
	if err != nil {
		t.Errorf("Got error from direct Cloudflare request, %v", err)

	}
	expected := strings.Split(fmt.Sprintf("%v", directResponse.Answer[0]), "\t")[4]
	if expected != actual {
		t.Errorf("Got different responses from Cloudlfare. Actual %v, Expected %v", indirectRR, directResponse.Answer)
	}

}
