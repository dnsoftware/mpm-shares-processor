package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"

	"github.com/dnsoftware/mpm-shares-processor/internal/usecase/analitics"
)

// Handler представляет HTTP сервер
type Handler struct {
	analitics *analitics.AnaliticsUsecase
	router    *chi.Mux
}

func NewHandler(analitics *analitics.AnaliticsUsecase) *Handler {
	s := &Handler{
		analitics: analitics,
		router:    chi.NewRouter(),
	}
	s.router.Use(middleware.Logger)
	s.routes()

	return s
}

func (s *Handler) Routes() *chi.Mux {
	return s.router
}

func (s *Handler) coinHashrate(w http.ResponseWriter, r *http.Request) {

	type Hashrate struct {
		Hashrate uint64 `json:"hashrate"`
	}
	coinSymbol := strings.ToUpper(chi.URLParam(r, "coinSymbol"))

	hr, err := s.analitics.CoinHashrate(coinSymbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashrate := Hashrate{
		Hashrate: hr,
	}

	json.NewEncoder(w).Encode(hashrate)
}

func (s *Handler) walletHashrate(w http.ResponseWriter, r *http.Request) {

	type Hashrate struct {
		Hashrate uint64 `json:"hashrate"`
	}
	walletStr := chi.URLParam(r, "walletID")

	walletID, err := strconv.ParseInt(walletStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hr, err := s.analitics.MinerHashrate(walletID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashrate := Hashrate{
		Hashrate: hr,
	}

	json.NewEncoder(w).Encode(hashrate)
}

// Создаем экземпляр WebSocket апгрейдера
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все источники (можно настроить более безопасно)
	},
}

func (s *Handler) websocketHandler(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем WebSocket-соединение
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка подключения WebSocket:", err)
		return
	}
	defer conn.Close()

	for {
		// Чтение сообщения от клиента
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Ошибка при чтении сообщения:", err)
			break
		}

		// Логируем полученное сообщение и отправляем его + дополнение обратно клиенту
		fmt.Printf("Получено сообщение: %s\n", p)
		if err := conn.WriteMessage(messageType, append(append(p, 43), p...)); err != nil {
			log.Println("Ошибка при отправке сообщения:", err)
			break
		}
	}
}
