package mqtt

import (
	"crypto/tls"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/shafreeck/fperf/client"
	"time"
)

func init() {
	client.Register("mqtt_publish", NewMqttClient, "benchmark of mqtt publish")
}

type setting struct {
	username string
	password string
	clean    bool
	clientID string

	topic   string
	qos     uint
	payload string
}
type mqttClient struct {
	cli     *MQTT.Client
	setting setting
}

func NewMqttClient(flag *client.FlagSet) client.Client {
	cli := new(mqttClient)
	flag.StringVar(&cli.setting.username, "username", "test", "username used to login")
	flag.StringVar(&cli.setting.password, "password", "test", "password of the username")
	flag.StringVar(&cli.setting.clientID, "clientid", "fperf-mqtt-publish", "ID of this client, this should be uniq")
	flag.BoolVar(&cli.setting.clean, "cleansession", true, "set cleansession flag")

	flag.StringVar(&cli.setting.topic, "topic", "/fperf/mqtt/publish", "topic to publish")
	flag.StringVar(&cli.setting.payload, "payload", "hello world", "what you want to publish")
	flag.UintVar(&cli.setting.qos, "qos", 1, "qos should be 0, 1, 2")
	flag.Parse()
	return cli
}

func (c *mqttClient) Dial(addr string) error {
	opts := MQTT.NewClientOptions().AddBroker(addr)
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	opts.SetTLSConfig(tlsConfig)
	opts.SetClientID(c.setting.clientID)
	opts.SetUsername(c.setting.username)
	opts.SetPassword(c.setting.password)
	opts.SetCleanSession(c.setting.clean)
	opts.SetKeepAlive(3600 * time.Second)
	c.cli = MQTT.NewClient(opts)
	if token := c.cli.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *mqttClient) Request() error {
	if token := c.cli.Publish(c.setting.topic, byte(c.setting.qos), false, c.setting.payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
