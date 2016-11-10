package commands

import (
	"github.com/urfave/cli"
	"github.com/streadway/amqp"
	"github.com/spf13/viper"
)


// ConfigCommands returns config commands
func RabbitMQCommands() []cli.Command {
	return []cli.Command{
		{
			Name:        "mq",
			Usage:       "options for mq",
			Before:      func(c *cli.Context) error {
				_, err := getConfig(MQHost)
				failOnError(err, "")
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:    "subscribe",
					Aliases: []string{"s"},
					Usage:   "subsctibe ",
					Flags:   []cli.Flag{
						cli.StringFlag{
							Name: "exchange, x",
							Usage: `-x "exchange_name"`,
						},
						cli.StringFlag{
							Name: "type, t",
							Usage: `-t "type"`,
						},
						cli.BoolFlag{
							Name: "durable, d",
						},
					},
					Action: func(c *cli.Context) {
						conn, err := amqp.Dial(viper.GetString(MQHost))
						failOnError(err, "Failed to connect to RabbitMQ")
						defer closeGracefully(conn)

						ch, err := conn.Channel()
						failOnError(err, "Failed to open a channel")
						defer closeGracefully(ch)

						err = ch.ExchangeDeclare(
							c.String("exchange"), // name
							c.String("type"), // type
							c.Bool("durable"), // durable
							false, // auto-deleted
							false, // internal
							false, // no-wait
							nil, // arguments
						)
						failOnError(err, "Failed to declare an exchange")

						q, err := ch.QueueDeclare(
							"", // name
							c.Bool("durable"), // durable
							true, // delete when unused
							true, // exclusive
							false, // no-wait
							nil, // arguments
						)
						failOnError(err, "Failed to declare a queue")

						log.Infof("Binding queue [%s] to exchange [%s] with routing key [%s]",
							q.Name, c.String("exchange"), c.Args().First())
						err = ch.QueueBind(
							q.Name, // queue name
							c.Args().First(), // routing key
							c.String("exchange"), // exchange
							false, // no-wait
							nil, // arguments
						)
						failOnError(err, "Failed to bind a queue")

						messages, err := ch.Consume(
							q.Name, // queue
							"", // consumer
							true, // auto ack
							false, // exclusive
							false, // no local
							false, // no wait
							nil, // args
						)
						failOnError(err, "Failed to register a consumer")

						forever := make(chan bool)

						go func() {
							for d := range messages {
								log.Debugf(" [x] %s", d.Body)
							}
						}()

						log.Debugf(" [*] Waiting for logs. To exit press CTRL+C")
						<-forever
					},
				},
			},
		},
	}
}
