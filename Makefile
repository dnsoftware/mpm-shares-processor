PROJECT="MPMS share processor"

default:
	echo ${PROJECT}

test:
	go test -v -count=1 ./...


.PHONY: default test cover
cover:
	go test -v -coverpkg=$(go list ./... | grep -v "proto" | tr '\n' ',') -coverprofile=coverage.out -covermode=count ./...
	go tool cover -func coverage.out | grep total | awk '{print $3}'

cover2:
	go test $(shell go list ./... | grep -v "proto") -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	rm coverage.out

cover3:
	go test $(shell go list ./... | grep -v "proto") -coverprofile=coverage.out -covermode=count ./...
	go tool cover -func coverage.out | grep total | awk '{print $3}'

.PHONY: staticlint
staticlint:
	go build -o ./staticlint ./cmd/staticlint
	go vet -vettool=staticlint ./...


####################### Миграции
# В базе будет создана таблица scheme_migrationa с полями: version и  dirty
# version - номер миграции (индекс файла миграции вида 000001, 000002 и т.д.)
# dirty - status выполнения миграйии, false - успешно, true - не выполнена (грязное выполнение)

# Запуск: make migrate-up DATABASE_URL="postgres://user:password@localhost:5432/mydb?sslmode=disable"
# ну или задаем DATABASE_URL напрямую здесь
DATABASE_URL="postgres://mpmpool:mpmpoolpass@62.113.106.101:6532/mpmpool?sslmode=disable"

# Путь к миграциям
MIGRATIONS_DIR = ./migration

# Команда для выполнения миграций
MIGRATE = migrate

# Если DATABASE_URL не задан, выводим ошибку
#ifndef DATABASE_URL
#$(error DATABASE_URL is not set. Pass it as an argument: make <target> DATABASE_URL=<connection_string>)
#endif

# Создание миграции
# Вызов:  make migrate-create migname=create_table_coin
# тут name - корень имени файла миграции
migrate-create:
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# Применить все миграции
# Вызов:  make migrate-up
migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

# Откатить одну миграцию
migrate-down:
	$(MIGRATE) -dir $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

# Откатить все миграции
migrate-reset:
	$(MIGRATE) -dir $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down -all

# Применить определённое количество миграций
migrate-step:
	$(MIGRATE) -dir $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up $(n)

# Откатить определённое количество миграций
migrate-down-step:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down $(n)


.PHONY: protogen
protogen:
	# здесь /home/dmitry/include/googleapis - путь к официальному репозиторию googleapis
	# который нужно предварительно склонировать командой: git clone https://github.com/googleapis/googleapis.git в эту (или другую) директорию
	# подробности читать тут: https://laradrom.ru/tag/proto/
	protoc --go_out=. --go-grpc_out=. -I.  -I/home/dmitry/include/googleapis proto/miners.proto
	protoc --go_out=. --go-grpc_out=. -I.  -I/home/dmitry/include/googleapis proto/shares.proto
