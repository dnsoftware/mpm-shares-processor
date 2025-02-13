package loaders

import (
	"fmt"
	"strings"

	"github.com/dnsoftware/mpmslib/pkg/configloader"
	"go.uber.org/zap"
	//"github.com/dnsoftware/alephiumpool/internal/constants"
	//"github.com/dnsoftware/alephiumpool/internal/domain/config"

	"github.com/dnsoftware/mpm-shares-processor/internal/constants"
)

// LoadRemoteConfig загрузка данных конфига с кластера etcd
// basePath - базовый путь к проекту
func LoadRemoteConfig(basePath string, logger *zap.Logger) error {

	startConf, err := configloader.LoadStartConfig(basePath + constants.StartConfigFilename)
	if err != nil {
		return fmt.Errorf("start config load error: %w", err)
	}

	// Получение основного конфига
	localConfigPath := basePath + constants.LocalConfigPath
	clusterNode := strings.Split(startConf.Etcd.Endpoints, ",")
	remoteDataKey := constants.ServiceDiscoveryPath + "/" + startConf.AppID
	caPath := basePath + constants.CaPath
	publicPath := basePath + constants.PublicPath
	privatePath := basePath + constants.PrivatePath

	confLoader, err := configloader.NewConfigLoader(clusterNode, caPath, publicPath, privatePath, localConfigPath, startConf.Etcd.Auth.Username, startConf.Etcd.Auth.Password)
	if err != nil {
		return err
	}

	confText, err := confLoader.LoadRemoteConfig(remoteDataKey)
	if err != nil { // Если удаленный конфиг не загрузился - логируем ошибку и загружаем локальный вариант
		return fmt.Errorf("no remote configs load: %w", err.Error())
	}

	// если нормально загрузился - сохраняем в локальный файл
	err = confLoader.SaveConfigToFile(confText)
	if err != nil {
		return fmt.Errorf("SaveConfigToFile error: %w", err)
	}

	return nil
}
