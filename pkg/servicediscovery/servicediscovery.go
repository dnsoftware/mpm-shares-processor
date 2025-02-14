package servicediscovery

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.etcd.io/etcd/client/v3"
)

type ServiceDiscovery struct {
	client         *clientv3.Client
	discoveryBase  string            // базовый путь к зарегистрированным сервисам
	services       map[string]string // ключи и значения (IP:port) внешних интерфейсов микросервиса в базе etcd
	leaseID        clientv3.LeaseID
	stopChannel    chan struct{}
	contextTimeout time.Duration // в секундах
	TTL            int64         // Время жизни записи в секундах
}

// NewServiceDiscovery создает новый экземпляр ServiceDiscovery
// и регистрирует один сервис
// если нужно зарегистрировать еще один сервис (например gRPC или что-то еще на другом порту - используем RegisterService)
func NewServiceDiscovery(cfg clientv3.Config, discoveryBase string, serviceKey, serviceAddr string, contextTimeout time.Duration, TTL int64) (*ServiceDiscovery, error) {
	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к etcd: %w", err)
	}

	sd := &ServiceDiscovery{
		client:         client,
		discoveryBase:  discoveryBase,
		services:       make(map[string]string),
		stopChannel:    make(chan struct{}),
		contextTimeout: contextTimeout,
		TTL:            TTL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), sd.contextTimeout*time.Second)
	defer cancel()
	lease, err := sd.client.Grant(ctx, sd.TTL)
	if err != nil {
		return nil, err
	}
	sd.leaseID = lease.ID

	err = sd.RegisterService(serviceKey, serviceAddr)
	if err != nil {
		return nil, err
	}

	return sd, nil
}

// RegisterService регистрирует сервис в etcd
func (sd *ServiceDiscovery) RegisterService(serviceKey, serviceAddr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), sd.contextTimeout*time.Second)
	defer cancel()

	sd.services[serviceKey] = serviceAddr

	_, err := sd.client.Put(ctx, sd.discoveryBase+"/"+serviceKey, serviceAddr, clientv3.WithLease(sd.leaseID))
	if err != nil {
		return err
	}

	// Запускаем автообновление TTL
	go sd.keepAlive()

	log.Printf("Сервис зарегистрирован: %s -> %s", serviceKey, serviceAddr)
	return nil
}

// keepAlive продлевает TTL сервиса
func (sd *ServiceDiscovery) keepAlive() {
	for {
		select {
		case <-sd.stopChannel:
			return
		default:
			ctx, cancel := context.WithTimeout(context.Background(), sd.contextTimeout*time.Second)
			_, err := sd.client.KeepAliveOnce(ctx, sd.leaseID)
			cancel()
			if err != nil {
				log.Printf("Ошибка обновления TTL: %v", err)
				return
			}
			time.Sleep(time.Duration(sd.TTL/2) * time.Second)
		}
	}
}

// DiscoverService ищет сервис по AppID
func (sd *ServiceDiscovery) DiscoverService(serviceKey string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sd.contextTimeout*time.Second)
	resp, err := sd.client.Get(ctx, sd.discoveryBase+"/"+serviceKey)
	cancel()
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("сервис %s не найден", serviceKey)
	}

	return string(resp.Kvs[0].Value), nil
}

// DiscoverAllServices получает список всех зарегистрированных сервисов
func (sd *ServiceDiscovery) DiscoverAllServices() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sd.contextTimeout*time.Second)
	resp, err := sd.client.Get(ctx, sd.discoveryBase, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, err
	}

	services := make(map[string]string)
	for _, kv := range resp.Kvs {
		if string(kv.Key) == sd.discoveryBase {
			continue
		}

		shortKey, ok := strings.CutPrefix(string(kv.Key), sd.discoveryBase+"/")
		if ok {
			services[shortKey] = string(kv.Value)
		}
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("не найдены активные сервисы с префиксом %s", sd.discoveryBase)
	}

	return services, nil
}

// Close удаляет сервис из реестра и завершает работу
func (sd *ServiceDiscovery) Close() {
	close(sd.stopChannel)
	ctx, cancel := context.WithTimeout(context.Background(), sd.contextTimeout*time.Second)
	defer cancel()
	sd.client.Revoke(ctx, sd.leaseID)
	for _, serviceKey := range sd.services {
		sd.client.Delete(ctx, serviceKey)
	}
	sd.client.Close()
}

func (sd *ServiceDiscovery) WaitDependencies(dependencies []string) {
	for {
		activeServices, err := sd.DiscoverAllServices()
		if err != nil {
			fmt.Println(err.Error())
		}

		time.Sleep(2 * time.Second)
		dependenciesOK := true
		for _, val := range dependencies {
			_, ok := activeServices[val]
			if !ok {
				fmt.Println("\033[33m[WARN]\033[0m Wait " + val + " service")
				dependenciesOK = false
				break
			}
		}

		if dependenciesOK {
			return
		}
	}
}
