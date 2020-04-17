package db

import (
	"encoding/json"

	"github.com/go-redis/redis/v7"
	"github.com/polisgo2020/search-Arkronzxc/config"
	"github.com/polisgo2020/search-Arkronzxc/index"
	"github.com/rs/zerolog/log"
)

type IndexRepository struct {
	c *redis.Client
}

func NewIndexRepository(conf *config.Config) (*IndexRepository, error) {
	cli := redis.NewClient(&redis.Options{
		Addr:     conf.DbListen,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := cli.Ping().Result()
	if err != nil {
		log.Err(err)
		return nil, err
	}
	log.Info().Str("result", pong).Msg("connection successful")
	return &IndexRepository{
		c: cli,
	}, nil
}

func (rep *IndexRepository) SaveIndex(i index.Index) error {
	for k, v := range i {
		finalJson, err := json.Marshal(v)
		if err != nil {
			log.Err(err)
			rep.c.FlushDB()
			return err
		}
		err = rep.c.Set(k, finalJson, 0).Err()
		if err != nil {
			log.Err(err).Str("key", k).Interface("value", v).Msg("error while setting values into DB")
			rep.c.FlushDB()
			log.Debug().Msg("db is cleaned")
			return err
		}
	}
	return nil
}

func (rep *IndexRepository) GetIndex(wordArr []string) (*index.Index, error) {
	var ind = make(index.Index)
	for _, v := range wordArr {
		val, err := rep.c.Get(v).Result()
		if err == redis.Nil {
			log.Debug().Str("key", v).Msg("key does not exist")
			return nil, err
		} else if err != nil {
			log.Err(err).Str("key", v).Msg("error while getting data by key")
			return nil, err
		} else {
			var data []string
			err := json.Unmarshal([]byte(val), &data)
			if err != nil {
				log.Err(err).Msg("error while db unmarshalling value")
			}
			ind[v] = data
		}
	}
	return &ind, nil
}
