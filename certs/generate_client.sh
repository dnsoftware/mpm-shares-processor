#!/bin/bash

# Создаём приватный ключ для клиента
openssl genrsa -out client.key 2048

# Создаём запрос на подпись сертификата (CSR)
openssl req -new -key client.key -out client.csr \
  -subj "/C=RU/ST=State/L=City/O=Organization/OU=Client/CN=save-get-shares-client"

# Создаём клиентский сертификат (Подписываем клиентский сертификат с использованием корневого сертификата)
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out client.crt -days 3650 -sha256

