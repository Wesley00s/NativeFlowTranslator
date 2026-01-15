package main

import (
	"log"
	"os"

	"translator-worker/internal/infra/llm/ollama"
	"translator-worker/internal/infra/queue"
	"translator-worker/internal/service"

	"github.com/joho/godotenv"
)

const (
	TranslationCommandQueue = "video.translation.cmd"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ .env file not found, using system env variables")
	}

	ollamaHost := os.Getenv("OLLAMA_HOST")
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	rabbitMQURL := os.Getenv("RABBITMQ_URL")

	if err := ollama.EnsureModelLoaded(ollamaHost, ollamaModel); err != nil {
		log.Fatalf("❌ Ollama check failed: %v", err)
	}

	provider := ollama.NewOllamaProvider(ollamaHost, ollamaModel)

	log.Println("🐰 Connecting to RabbitMQ...")
	rabbitProducer, err := queue.NewRabbitMQProducer(rabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitProducer.Close()

	translationProcessor := service.NewTranslationProcessor(provider, rabbitProducer)

	transConsumer, err := queue.NewRabbitMQConsumer(rabbitMQURL, TranslationCommandQueue)
	if err != nil {
		log.Fatal(err)
	}
	defer transConsumer.Close()
	transMsgs, _ := transConsumer.StartConsuming()

	log.Println("🚀 Worker Started! Listening for separate Translation...")

	forever := make(chan struct{})

	go func() {
		for d := range transMsgs {
			log.Printf("📥 [Translation CMD] Received: %d bytes", len(d.Body))
			translationProcessor.ProcessMessage(d.Body)
			err := d.Ack(false)
			if err != nil {
				log.Printf("❌ Error acknowledging message: %v", err)
			}
		}
	}()

	<-forever
}
