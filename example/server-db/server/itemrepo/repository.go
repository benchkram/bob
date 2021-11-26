package itemrepo

import (
	"context"
	"encoding/json"
	"server-db/server"
	"server-db/server/database"
	"server-db/server/item"
	"time"

	"github.com/go-redis/redis/v8"
)

type Repository struct {
	db *database.Database
}

type RedisItem struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int    `json:"createdAt"`
}

func (r *Repository) CreateItem(i item.Item) error {
	ri := RedisItem{
		Id:        i.Id(),
		Name:      i.Name(),
		CreatedAt: int(i.CreatedAt().Unix()),
	}

	b, err := json.Marshal(ri)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = r.db.Set(ctx, i.Id(), string(b), 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Item(id string) (*item.Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	s, err := r.db.Get(ctx, id).Result()
	if err == redis.Nil {
		return nil, server.ErrItemNotFound
	} else if err != nil {
		return nil, err
	}

	var ri RedisItem
	err = json.Unmarshal([]byte(s), &ri)
	if err != nil {
		return nil, err
	}

	i := item.New(
		item.WithId(ri.Id),
		item.WithName(ri.Name),
		item.WithCreatedAt(time.Unix(int64(ri.CreatedAt), 0)),
	)

	return &i, nil
}

func New(db *database.Database) *Repository {
	return &Repository{
		db: db,
	}
}
