package main

import (
	"log"

	"github.com/orxma/depwise/internal/bot"
)

func main() {
	log.Println("Iniciando ORX TUNNEL Bot...")

	// Iniciar servidor del bot (bloqueante)
	bot.StartBot()
}
