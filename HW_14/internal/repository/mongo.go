package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"HW_14/internal/model/game"
	"HW_14/internal/model/move"
	"HW_14/internal/model/player"
)

// чтобы запрос не висел вечно
func newCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// подключаемся к монге и проверяем что она отвечает
func ConnectMongo(uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	ctx, cancel := newCtx()
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// id должен быть уникальным
func createIndex(coll *mongo.Collection) error {
	ctx, cancel := newCtx()
	defer cancel()
	_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

// берём самый большой id и прибавляем 1
func nextID(coll *mongo.Collection) int {
	ctx, cancel := newCtx()
	defer cancel()
	var doc struct {
		ID int `bson:"id"`
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "id", Value: -1}})
	err := coll.FindOne(ctx, bson.M{}, opts).Decode(&doc)
	if err != nil {
		// коллекция пустая
		return 1
	}
	return doc.ID + 1
}

func insert(coll *mongo.Collection, doc any) error {
	ctx, cancel := newCtx()
	defer cancel()
	_, err := coll.InsertOne(ctx, doc)
	return err
}

func findByID(coll *mongo.Collection, id int, result any) bool {
	ctx, cancel := newCtx()
	defer cancel()
	err := coll.FindOne(ctx, bson.M{"id": id}).Decode(result)
	return err == nil
}

func findAll(coll *mongo.Collection, result any) error {
	ctx, cancel := newCtx()
	defer cancel()
	opts := options.Find().SetSort(bson.D{{Key: "id", Value: 1}})
	cur, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return err
	}
	return cur.All(ctx, result)
}

func replace(coll *mongo.Collection, id int, doc any) error {
	ctx, cancel := newCtx()
	defer cancel()
	_, err := coll.ReplaceOne(ctx, bson.M{"id": id}, doc)
	return err
}

func deleteByID(coll *mongo.Collection, id int) (bool, error) {
	ctx, cancel := newCtx()
	defer cancel()
	res, err := coll.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}

// игроки

type Players struct {
	coll *mongo.Collection
}

func NewPlayers(db *mongo.Database) (*Players, error) {
	coll := db.Collection("players")
	err := createIndex(coll)
	if err != nil {
		return nil, err
	}
	return &Players{coll: coll}, nil
}

func (s *Players) NextID() int {
	return nextID(s.coll)
}

func (s *Players) Add(p *player.Player) error {
	return insert(s.coll, p)
}

func (s *Players) Get(id int) *player.Player {
	var p player.Player
	if !findByID(s.coll, id, &p) {
		return nil
	}
	return &p
}

func (s *Players) All() []*player.Player {
	list := []*player.Player{}
	err := findAll(s.coll, &list)
	if err != nil {
		return []*player.Player{}
	}
	return list
}

func (s *Players) Update(p *player.Player) error {
	return replace(s.coll, p.ID, p)
}

func (s *Players) Delete(id int) (bool, error) {
	return deleteByID(s.coll, id)
}

// игры

type Games struct {
	coll *mongo.Collection
}

func NewGames(db *mongo.Database) (*Games, error) {
	coll := db.Collection("games")
	err := createIndex(coll)
	if err != nil {
		return nil, err
	}
	return &Games{coll: coll}, nil
}

func (s *Games) NextID() int {
	return nextID(s.coll)
}

func (s *Games) Add(g *game.Game) error {
	return insert(s.coll, g)
}

func (s *Games) Get(id int) *game.Game {
	var g game.Game
	if !findByID(s.coll, id, &g) {
		return nil
	}
	return &g
}

func (s *Games) All() []*game.Game {
	list := []*game.Game{}
	err := findAll(s.coll, &list)
	if err != nil {
		return []*game.Game{}
	}
	return list
}

func (s *Games) Update(g *game.Game) error {
	return replace(s.coll, g.ID, g)
}

func (s *Games) Delete(id int) (bool, error) {
	return deleteByID(s.coll, id)
}

// ходы

type Moves struct {
	coll *mongo.Collection
}

func NewMoves(db *mongo.Database) (*Moves, error) {
	coll := db.Collection("moves")
	err := createIndex(coll)
	if err != nil {
		return nil, err
	}
	return &Moves{coll: coll}, nil
}

func (s *Moves) NextID() int {
	return nextID(s.coll)
}

func (s *Moves) Add(m *move.Move) error {
	return insert(s.coll, m)
}

func (s *Moves) Get(id int) *move.Move {
	var m move.Move
	if !findByID(s.coll, id, &m) {
		return nil
	}
	return &m
}

func (s *Moves) All() []*move.Move {
	list := []*move.Move{}
	err := findAll(s.coll, &list)
	if err != nil {
		return []*move.Move{}
	}
	return list
}
