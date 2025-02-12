#!/bin/bash

# Создаём приватный ключ для сервера
openssl genrsa -out server.key 2048

# Создаём запрос на подпись сертификата (CSR)
openssl req -new -key server.key -out server.csr \
  -subj "/C=RU/ST=State/L=City/O=Organization/OU=Server/CN=save-get-shares-server"

# Создаём серверный сертификат (Подписываем серверный сертификат с использованием корневого сертификата.)
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days 3650 -sha256
