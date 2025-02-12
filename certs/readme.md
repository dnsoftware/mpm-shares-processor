### Генерация TLS сертификатов 
смотри описание https://laradrom.ru/languages/golang/golang-generacziya-sertifikatov-s-sobstvennym-kornevym-sertifikatom-dlya-ispolzovaniya-mtls-dlya-raboty-mikroservisov/

Корневой сертификат (ca.crt, ca.key) генерируется в mpm-miners-processor - забираем оттуда

--------

**generate_server.sh** - генерация серверного сертификата

**server.key**: приватный ключ сервера.

**server.csr**: запрос на подпись сертификата.

**server.crt**: подписанный сертификат сервера.

---

**generate_client.sh** - генерация клиентского сертификата

**client.key**: приватный ключ сервера.

**client.csr**: запрос на подпись сертификата.

**client.crt**: подписанный сертификат сервера.
