package kafka_reader

// Внимание! Добавить запись в кафку вначале, чтобы было что читать
// Тестирование получения шар из кафки
// кафка должна быть запущена и в ней в соответствующем топике должно быть несколько шар
//func TestReader(t *testing.T) {
//
//	filePath, err := logger.GetLoggerMainLogPath()
//	require.NoError(t, err)
//	logger.InitLogger(logger.LogLevelDebug, filePath)
//	//logger.InitLogger(logger.LogLevelProduction, filePath)
//
//	cfg := Config{
//		Brokers:            []string{"localhost:9092", "localhost:9093", "localhost:9094"},
//		Group:              constants.KafkaSharesGroup,
//		Topic:              "test_topic", // constants.KafkaSharesTopic
//		AutoCommitInterval: constants.KafkaSharesAutocommitInterval,
//		AutoCommitEnable:   true,
//	}
//
//	reader, err := NewKafkaReader(cfg, logger.Log())
//	assert.NoError(t, err)
//
//	// сброс в начало
//	err = reader.SetGroupOffset(sarama.OffsetOldest)
//	assert.NoError(t, err)
//
//	handler := &shares.ShareConsumer{}
//	reader.ConsumeMessages(handler)
//
//	// задержка, чтобы сработал автокоммит
//	time.Sleep(10 * time.Second)
//	reader.Close()
//
//}
