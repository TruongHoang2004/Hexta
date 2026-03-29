package bootstrap

import (
	"gitlab.com/ecommercehub1/catalog/internal/infrastructure/elasticsearch"
	"go.uber.org/fx"
)

func ElasticsearchModule() fx.Option {
	return fx.Options(
		fx.Provide(elasticsearch.NewElasticsearchClient),
	)
}
