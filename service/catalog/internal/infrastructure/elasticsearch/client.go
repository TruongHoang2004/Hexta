package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v8"
	"gitlab.com/ecommercehub1/catalog/config"
	"gitlab.com/ecommercehub1/catalog/internal/common/log"
)

func NewElasticsearchClient() (*elasticsearch.Client, error) {
	logger := log.GetLogger()
	cfg := elasticsearch.Config{
		Addresses: config.AppConfig.Elasticsearch.Addresses,
		Username:  config.AppConfig.Elasticsearch.Username,
		Password:  config.AppConfig.Elasticsearch.Password,
	}

	logger.Info("🔌 Connecting to Elasticsearch at " + config.AppConfig.Elasticsearch.Addresses[0])
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		logger.Error("Failed to create elasticsearch client: %v", err)
		return nil, err
	}

	res, err := client.Info()
	if err != nil {
		logger.Error("Failed to connect to elasticsearch: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		logger.Error("Elasticsearch connection returned error: %s", res.String())
	} else {
		logger.Info("✅ Elasticsearch connected")
	}

	return client, nil
}
