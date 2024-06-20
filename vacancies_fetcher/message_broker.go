package vacancies_fetcher

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
	"time"
	"vacancies/vacancies_fetcher/config"
	"vacancies/vacancies_fetcher/internal"
	"vacancies/vacancies_fetcher/repository"
)

type Rabbit struct {
	Channel *amqp.Channel
	Queue   amqp.Queue
	Repo    *repository.Repo
}

func NewRabbit(cfg *config.Config, channel *amqp.Channel) *Rabbit {
	q, err := channel.QueueDeclare(
		"vacancies", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare a queue")
	return &Rabbit{
		Channel: channel,
		Queue:   q,
	}
}

func (r *Rabbit) Send(vacancies []internal.Vacancy, keyWord string) error {
	hrefs := make([]string, len(vacancies))
	titles := make([]string, len(vacancies))
	for i, vacancy := range vacancies {
		hrefs[i] = vacancy.Reference
		titles[i] = vacancy.Title
	}

	message := internal.VacancyQueueDTO{
		KeyWord: keyWord,
		Hrefs:   hrefs,
		Titles:  titles,
	}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = r.Channel.PublishWithContext(ctx,
		"",           // exchange
		r.Queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		})
	return r.Channel.Close()
}

func (r *Rabbit) StartReceiving(wg *sync.WaitGroup) error {
	msgs, err := r.Channel.Consume(
		r.Queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for d := range msgs {
			var vac internal.VacancyQueueDTO
			err = json.Unmarshal(d.Body, &vac)
			if err != nil {
				log.Printf("Error unmarshalling vacancy queue due to %s", err.Error())
			}
			vacParams := repository.AddVacanciesParams{
				KeyWord: vac.KeyWord,
				Hrefs:   vac.Hrefs,
				Titles:  vac.Titles,
			}
			err = r.Repo.AddVacancies(ctx, vacParams)
			if err != nil {
				log.Printf("Error adding vacancy to repository due to %s", err.Error())
			}
		}
		wg.Done()
	}()
	return nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
