package rest

func (s *Handler) routes() {
	s.router.Get("/coin/{coinSymbol}/hashrate", s.coinHashrate)
	s.router.Get("/wallet/{walletID}/hashrate", s.walletHashrate)

	// Маршрут для WebSocket
	s.router.Get("/ws", s.websocketHandler)

}
