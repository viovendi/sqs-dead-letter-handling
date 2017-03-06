package main // import "github.com/Jimdo/sqs-dead-letter-requeue"

import (
	"log"
	"os"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/sqs"
	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	app       = kingpin.New("dead-letter-requeue", "Requeues messages from a SQS dead-letter queue to the active one.")
	queueName = app.Arg("queue-name", "Name of the SQS queue (e.g. prod-mgmt-website-data-www100-jimdo-com).").Required().String()
        deadLetterQueueName = app.Arg("dead-letter-queue-name", "Name of the SQS queue (e.g. prod-mgmt-website-data-www100-jimdo-com-dead-letter).").Required().String()
        cycles = app.Arg("cycles", "Number of cycles to requeue").Required().Int()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	cycles := *cycles
         
        activeQueueName := *queueName

	deadLetterQueueName := *deadLetterQueueName

	auth, err := aws.EnvAuth()
	if err != nil {
		log.Fatal(err)
		return
	}

	conn := sqs.New(auth, aws.EUCentral)

	deadLetterQueue, err := conn.GetQueue(deadLetterQueueName)
	if err != nil {
		log.Fatal(err)
		return
	}

	activeQueue, err := conn.GetQueue(activeQueueName)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Looking for messages to requeue.")
	iterations := 0
        for iterations < cycles {
		iterations += 1
                resp, err := deadLetterQueue.ReceiveMessageWithParameters(
			map[string]string{
				"WaitTimeSeconds":     "20",
				"MaxNumberOfMessages": "10",
				"VisibilityTimeout":   "20"})
		if err != nil {
			log.Fatal(err)
			return
		}

		messages := resp.Messages
		numberOfMessages := len(messages)
		if numberOfMessages == 0 {
			log.Printf("Requeuing messages done.")
			return
		} else {
			log.Printf("Moving %v message(s)...", numberOfMessages)
		}

		_, err = activeQueue.SendMessageBatch(messages)
		if err != nil {
			log.Fatal(err)
			return
		}

		_, err = deadLetterQueue.DeleteMessageBatch(messages)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}
