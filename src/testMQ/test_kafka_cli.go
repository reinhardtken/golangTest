/*golang如何使用sarama访问kafka
golang连接kafka有三种client认证方式：

无认证
TLS认证
SASL/PLAIN认证, (其他SASL/SCRAM, SASL/GSSAPI都不支持)

下面一个客户端代码例子访问kafka服务器，来发送和接受消息。

使用方式


命令行参数

$ ./kafkaclient -h
Usage of ./client:
  -ca string
        CA Certificate (default "ca.pem")
  -cert string
        Client Certificate (default "cert.pem")
  -command string
        consumer|producer (default "consumer")
  -host string
        Common separated kafka hosts (default "localhost:9093")
  -key string
        Client Key (default "key.pem")
  -partition int
        Kafka topic partition
  -password string
        SASL password
  -sasl
        SASL enable
  -tls
        TLS enable
  -topic string
        Kafka topic (default "test--topic")
  -username string
        SASL username


作为producer启动

$ ./kafkaclient -command producer \
  -host kafka1:9092,kafka2:9092

## SASL/PLAIN enable
$ ./kafkaclient -command producer \
  -sasl -username kafkaclient -password kafkapassword \
  -host kafka1:9092,kafka2:9092

## TLS-enabled
$ ./kafkaclient -command producer \
  -tls -cert client.pem -key client.key -ca ca.pem \
  -host kafka1:9093,kafka2:9093

producer发送消息给kafka：
> aaa
2018/12/15 07:11:21 Produced message: [aaa]
> bbb
2018/12/15 07:11:30 Produced message: [bbb]
> quit


作为consumer启动

$ ./kafkaclient -command consumer \
  -host kafka1:9092,kafka2:9092

## SASL/PLAIN enabled
$ ./kafkaclient -command consumer \
  -sasl -username kafkaclient -password kafkapassword \
  -host kafka1:9092,kafka2:9092

## TLS-enabled
$ ./kafkaclient -command consumer \
  -tls -cert client.pem -key client.key -ca ca.pem \
  -host kafka1:9093,kafka2:9093

consumer从kafka接受消息：
2018/12/15 07:11:21 Consumed message: [aaa], offset: [4]
2018/12/15 07:11:30 Consumed message: [bbb], offset: [5]*/



package main

import (
"flag"
"fmt"
"log"
"os"
"io/ioutil"
"bufio"
"strings"

"crypto/tls"
"crypto/x509"

"github.com/Shopify/sarama"
)

var (
	command     string
	hosts       string
	topic       string
	partition   int
	saslEnable  bool
	username    string
	password    string
	tlsEnable   bool
	clientcert  string
	clientkey   string
	cacert      string
)

func main() {
	flag.StringVar(&command,    "command",      "consumer",         "consumer|producer")
	flag.StringVar(&hosts,      "host",         "localhost:9093",   "Common separated kafka hosts")
	flag.StringVar(&topic,      "topic",        "test--topic",      "Kafka topic")
	flag.IntVar(&partition,     "partition",    0,                  "Kafka topic partition")

	flag.BoolVar(&saslEnable,   "sasl",          false,             "SASL enable")
	flag.StringVar(&username,   "username",      "",                "SASL Username")
	flag.StringVar(&password,   "password",      "",                "SASL Password")

	flag.BoolVar(&tlsEnable,    "tls",          false,              "TLS enable")
	flag.StringVar(&clientcert, "cert",         "cert.pem",         "Client Certificate")
	flag.StringVar(&clientkey,  "key",          "key.pem",          "Client Key")
	flag.StringVar(&cacert,     "ca",           "ca.pem",           "CA Certificate")
	flag.Parse()

	config := sarama.NewConfig()
	if saslEnable {
		config.Net.SASL.Enable   = true
		config.Net.SASL.User     = username
		config.Net.SASL.Password = password
	}

	if tlsEnable {
		//sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
		tlsConfig, err := genTLSConfig(clientcert, clientkey, cacert)
		if err != nil {
			log.Fatal(err)
		}

		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	client, err := sarama.NewClient(strings.Split(hosts, ","), config)
	if err != nil {
		log.Fatalf("unable to create kafka client: %q", err)
	}

	if command == "consumer" {
		consumer, err := sarama.NewConsumerFromClient(client)
		if err != nil {
			log.Fatal(err)
		}
		defer consumer.Close()
		loopConsumer(consumer, topic, partition)
	} else {
		producer, err := sarama.NewAsyncProducerFromClient(client)
		if err != nil {
			log.Fatal(err)
		}
		defer producer.Close()
		loopProducer(producer, topic, partition)
	}
}

func genTLSConfig(clientcertfile, clientkeyfile, cacertfile string) (*tls.Config, error) {
	// load client cert
	clientcert, err := tls.LoadX509KeyPair(clientcertfile, clientkeyfile)
	if err != nil {
		return nil, err
	}

	// load ca cert pool
	cacert, err := ioutil.ReadFile(cacertfile)
	if err != nil {
		return nil, err
	}
	cacertpool := x509.NewCertPool()
	cacertpool.AppendCertsFromPEM(cacert)

	// generate tlcconfig
	tlsConfig := tls.Config{}
	tlsConfig.RootCAs = cacertpool
	tlsConfig.Certificates = []tls.Certificate{clientcert}
	tlsConfig.BuildNameToCertificate()
	// tlsConfig.InsecureSkipVerify = true // This can be used on test server if domain does not match cert:
	return &tlsConfig, err
}

func loopProducer(producer sarama.AsyncProducer, topic string, partition int) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
		} else if text == "exit" || text == "quit" {
			break
		} else {
			producer.Input() <- &sarama.ProducerMessage{Topic: topic, Key: nil, Value: sarama.StringEncoder(text)}
			log.Printf("Produced message: [%s]\n",text)
		}
		fmt.Print("> ")
	}
}

func loopConsumer(consumer sarama.Consumer, topic string, partition int) {
	partitionConsumer, err := consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
	if err != nil {
		log.Println(err)
		return
	}
	defer partitionConsumer.Close()

	for {
		msg := <-partitionConsumer.Messages()
		log.Printf("Consumed message: [%s], offset: [%d]\n", msg.Value, msg.Offset)
	}
}