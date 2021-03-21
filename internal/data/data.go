package data

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/extra/redisotel"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"

	// init mysql driver
	_ "github.com/go-sql-driver/mysql"

	"kratos-learning/internal/conf"
	"kratos-learning/internal/data/ent"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewArticleRepo)

// Data .
type Data struct {
	db  *ent.Client
	rdb *redis.Client
}

// NewData .
func NewData(conf *conf.Data, logger log.Logger) (*Data, error) {
	log := log.NewHelper("data", logger)
	client, err := ent.Open(
		conf.Database.Driver,
		conf.Database.Source,
	)
	if err != nil {
		log.Errorf("failed opening connection to sqlite: %v", err)
		return nil, err
	}
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Errorf("failed creating schema resources: %v", err)
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         conf.Redis.Addr,
		Password:     conf.Redis.Password,
		DB:           int(conf.Redis.Db),
		DialTimeout:  conf.Redis.DialTimeout.AsDuration(),
		WriteTimeout: conf.Redis.WriteTimeout.AsDuration(),
		ReadTimeout:  conf.Redis.ReadTimeout.AsDuration(),
	})
	rdb.AddHook(redisotel.TracingHook{})
	return &Data{
		db:  client,
		rdb: rdb,
	}, nil
}

// -----------------redis-------------------
func likeKey(id int64) string {
	return fmt.Sprintf("like:%d", id)
}

func (ar *articleRepo) GetArticleLike(ctx context.Context, id int64) (rv int64, err error) {
	get := ar.data.rdb.Get(ctx, likeKey(id))
	rv, err = get.Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return
}

func (ar *articleRepo) IncArticleLike(ctx context.Context, id int64) error {
	_, err := ar.data.rdb.Incr(ctx, likeKey(id)).Result()
	return err
}