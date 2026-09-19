package main

import (
	"log"

	"offerhunter/internal/config"
	"offerhunter/internal/server"
)

func main() {
	if config.DeepSeekAPIKey == "" {
		log.Fatal("DEEPSEEK_API_KEY 未配置：请复制 .env.example 为 .env 并填入真实密钥")
	}
	if config.MiMoAPIKey == "" {
		log.Println("警告：MIMO_API_KEY 未配置，PDF/图片简历 OCR 将不可用")
	}

	r := server.NewRouter()
	addr := config.ServerHost + ":" + config.ServerPort
	log.Printf("Offer-Hunter LangChain Edition starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
