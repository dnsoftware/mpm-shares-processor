package kafka_writer

import "github.com/IBM/sarama"

type SaramaHeadersCarrier []sarama.RecordHeader

/**** Реализация интерфейса TextMapCarrier ****/
/**/
func (c *SaramaHeadersCarrier) Get(key string) string {
	for _, h := range *c {
		if string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c *SaramaHeadersCarrier) Set(key, value string) {
	*c = append(*c, sarama.RecordHeader{Key: []byte(key), Value: []byte(value)})
}

func (c *SaramaHeadersCarrier) Keys() []string {
	keys := make([]string, len(*c))
	for i, h := range *c {
		keys[i] = string(h.Key)
	}
	return keys
}

/**/
