package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	database, collection string
	ctx                  context.Context
	conn                 *mongo.Client
}

type Inserter interface {
	InsertGameStats(documents []interface{}) error
}

func NewMongo(ctx context.Context, dsn, database, collection string) (*Mongo, error) {
	conn, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, err
	}
	return &Mongo{
		database:   database,
		collection: collection,
		ctx:        ctx,
		conn:       conn,
	}, nil
}

func (m *Mongo) InsertGameStats(documents []interface{}) error {
	_, err := m.conn.
		Database(m.database).
		Collection(m.collection).
		InsertMany(m.ctx, documents)
	return err
}
