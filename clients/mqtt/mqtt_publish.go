package mqtt

import (
	"crypto/tls"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/shafreeck/fperf"
	"time"

	"io/ioutil"
	"log"
	"math/rand"
	"strings"
)

var idgen func() string

func init() {
	fperf.Register("mqtt-publish", NewMqttClient, "benchmark of mqtt publish")
	idgen = idgenerator()
}

type setting struct {
	username string
	password string
	clean    bool
	clientID string

	topic   string
	qos     uint
	payload string
	load    string

	verbose bool
}
type mqttClient struct {
	cli     *MQTT.Client
	setting setting
	topics  []string
}

func NewMqttClient(flag *fperf.FlagSet) fperf.Client {
	cli := new(mqttClient)
	flag.StringVar(&cli.setting.username, "username", "test", "username used to login")
	flag.StringVar(&cli.setting.password, "password", "test", "password of the username")
	flag.StringVar(&cli.setting.clientID, "clientid", "fperf-mqtt-publish", "ID of this client, this should be uniq")
	flag.BoolVar(&cli.setting.clean, "cleansession", true, "set cleansession flag")

	flag.StringVar(&cli.setting.topic, "topic", "/fperf/mqtt/publish", "topic to publish")
	flag.StringVar(&cli.setting.payload, "payload", "hello world", "what you want to publish")
	flag.StringVar(&cli.setting.load, "loadtopic", "", "path of topic file")
	flag.UintVar(&cli.setting.qos, "qos", 1, "qos should be 0, 1, 2")

	flag.BoolVar(&cli.setting.verbose, "v", false, "verbose")
	flag.Parse()
	if len(cli.setting.load) > 0 {
		if content, err := ioutil.ReadFile(cli.setting.load); err != nil {
			log.Fatal(err)
		} else {
			cli.topics = strings.Split(strings.TrimSpace(string(content)), "\n")
			if cli.setting.verbose {
				log.Println("load topics", cli.topics)
			}
		}
	}
	rand.Seed(time.Now().UnixNano())

	return cli
}

func idgenerator() func() string {
	i := 0
	return func() string {
		i++
		return fmt.Sprintf("%d", i)
	}
}

func (c *mqttClient) Dial(addr string) error {
	opts := MQTT.NewClientOptions().AddBroker(addr)
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	opts.SetTLSConfig(tlsConfig)
	opts.SetClientID(c.setting.clientID + "-" + idgen())
	opts.SetUsername(c.setting.username)
	opts.SetPassword(c.setting.password)
	opts.SetCleanSession(c.setting.clean)
	opts.SetKeepAlive(3600 * time.Second)
	opts.SetProtocolVersion(4)
	c.cli = MQTT.NewClient(opts)
	if token := c.cli.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	if c.setting.verbose {
		fmt.Println("dial to", addr)
	}
	return nil
}

func (c *mqttClient) Request() error {
	var err error
	topic := c.setting.topic
	if len(c.topics) > 0 {
		topic = c.topics[rand.Intn(len(c.topics))]
	}

	payload := []byte(c.setting.payload)
	if c.setting.payload == "now" {
		payload, err = time.Now().MarshalBinary()
		if err != nil {
			return err
		}
	}
	if token := c.cli.Publish(topic, byte(c.setting.qos), false, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
