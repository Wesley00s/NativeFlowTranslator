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
	SyncCommandQueue        = "video.sync.cmd"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		return
	}

	ollamaHost := os.Getenv("OLLAMA_HOST")
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	rabbitMQURL := os.Getenv("RABBITMQ_URL")

	if err := ollama.EnsureModelLoaded(ollamaHost, ollamaModel); err != nil {
		log.Fatalf("❌ Ollama check failed: %v", err)
	}

	provider := ollama.NewOllamaProvider(ollamaHost, ollamaModel)

	aligner := ollama.NewOllamaAligner(ollamaHost, ollamaModel)

	log.Println("🐰 Connecting to RabbitMQ...")
	rabbitProducer, err := queue.NewRabbitMQProducer(rabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitProducer.Close()

	translationProcessor := service.NewTranslationProcessor(provider, rabbitProducer)

	syncService := service.NewSyncService(aligner, rabbitProducer)

	transConsumer, err := queue.NewRabbitMQConsumer(rabbitMQURL, TranslationCommandQueue)
	if err != nil {
		log.Fatal(err)
	}
	defer transConsumer.Close()
	transMsgs, _ := transConsumer.StartConsuming()

	syncConsumer, err := queue.NewRabbitMQConsumer(rabbitMQURL, SyncCommandQueue)
	if err != nil {
		log.Fatal(err)
	}
	defer syncConsumer.Close()
	syncMsgs, _ := syncConsumer.StartConsuming()

	log.Println("🚀 Worker Started! Listening for separate Translation and Sync jobs...")

	forever := make(chan struct{})

	go func() {
		for d := range transMsgs {
			log.Printf("📥 [Translation CMD] Received: %d bytes", len(d.Body))
			translationProcessor.ProcessMessage(d.Body)
			err := d.Ack(false)
			if err != nil {
				return
			}
		}
	}()

	go func() {
		for d := range syncMsgs {
			log.Printf("📥 [Sync CMD] Received: %d bytes", len(d.Body))
			syncService.ProcessSync(d.Body)
			err := d.Ack(false)
			if err != nil {
				return
			}
		}
	}()

	<-forever
}
