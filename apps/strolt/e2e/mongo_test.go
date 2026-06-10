package e2e_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mongoIndexName = "idx_name"

// MongoDoc exercises what mongodump/mongorestore must preserve: unicode,
// arrays, nested documents and raw binary data.
type MongoDoc struct {
	ID      int      `bson:"_id"`
	Name    string   `bson:"name"`
	Tags    []string `bson:"tags"`
	Meta    Meta     `bson:"meta"`
	Payload []byte   `bson:"payload"`
}

type Meta struct {
	Lang  string `bson:"lang"`
	Stars int    `bson:"stars"`
}

var mongoDoc = MongoDoc{
	ID:   1,
	Name: "üñîçødé-документ-🚀",
	Tags: []string{"backup", "restore", "проверка"},
	Meta: Meta{
		Lang:  "go",
		Stars: 42,
	},
	Payload: binaryPayload(),
}

type MongoSuite struct {
	suite.Suite

	c *MongoConn
}

type MongoConn struct {
	client     *mongo.Client
	database   string
	collection string
}

func (s *MongoSuite) SetupSuite() {
	c, err := mongoConnect()
	s.Require().NoError(err)
	s.c = c
}

func (s *MongoSuite) TearDownSuite() {
	s.NoError(s.c.close())
}

func (s *MongoSuite) BeforeTest(suiteName, testName string) {
	s.Require().NoError(s.c.drop())
	s.Require().NoError(s.c.createCollection())
	s.Require().NoError(s.c.insertData())
	s.Require().NoError(s.c.checkValidData())
}

func (s *MongoSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.c.checkValidData())
}

func (s *MongoSuite) TestMongo() {
	s.roundTrip("e2e")
}

func (s *MongoSuite) TestMongo_copy() {
	s.roundTrip("e2e-copy")
}

func (s *MongoSuite) roundTrip(serviceName string) {
	s.T().Helper()

	s.Require().NoError(strolt("backup", "--service", serviceName, "--task", "mongo", "--y"))

	s.Require().NoError(s.c.drop())

	latestSnapshotID, err := stroltGetLatestSnapshotID(serviceName, "mongo", "restic-mongo")
	s.Require().NoError(err)

	s.NoError(strolt("restore", "--service", serviceName, "--task", "mongo", "--destination", "restic-mongo", "--snapshot", latestSnapshotID, "--y"))
}

//nolint:thelper
func MongoSuiteTest(t *testing.T) {
	tt := timeTook("MongoSuiteTest")

	suite.Run(t, new(MongoSuite))
	tt.stop()
}

func mongoConnect() (*MongoConn, error) {
	port, err := containerManager.GetMongoPort()
	if err != nil {
		return nil, err
	}

	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI("mongodb://localhost:"+port))
	if err != nil {
		return nil, err
	}

	// Connect is lazy; ping to make sure the server is actually reachable
	// before the suite starts mutating data.
	if err := client.Ping(connectCtx, nil); err != nil {
		return nil, errors.Join(err, client.Disconnect(context.Background()))
	}

	return &MongoConn{
		client:     client,
		database:   "strolt",
		collection: "strolt",
	}, nil
}

func (c *MongoConn) close() error {
	disconnectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.client.Disconnect(disconnectCtx)
}

func (c *MongoConn) drop() error {
	return c.client.Database(c.database).Drop(ctx)
}

func (c *MongoConn) createCollection() error {
	if err := c.client.Database(c.database).CreateCollection(ctx, c.collection); err != nil {
		return err
	}

	// A named secondary index verifies that mongorestore brings back index
	// definitions, not only documents.
	_, err := c.client.Database(c.database).Collection(c.collection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetName(mongoIndexName),
	})

	return err
}

func (c *MongoConn) insertData() error {
	collection := c.client.Database(c.database).Collection(c.collection)
	if _, err := collection.InsertOne(ctx, mongoDoc); err != nil {
		return err
	}

	return nil
}

func (c *MongoConn) checkValidData() error {
	collection := c.client.Database(c.database).Collection(c.collection)

	count, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return err
	}

	if count != 1 {
		return fmt.Errorf("expected 1 document, got %d", count)
	}

	var doc MongoDoc
	if err := collection.FindOne(ctx, bson.D{{Key: "_id", Value: mongoDoc.ID}}).Decode(&doc); err != nil {
		return err
	}

	if doc.Name != mongoDoc.Name || doc.Meta != mongoDoc.Meta || len(doc.Tags) != len(mongoDoc.Tags) {
		return fmt.Errorf("document mismatch: want %+v, got %+v", mongoDoc, doc)
	}

	for i, tag := range mongoDoc.Tags {
		if doc.Tags[i] != tag {
			return fmt.Errorf("document tags mismatch: want %v, got %v", mongoDoc.Tags, doc.Tags)
		}
	}

	if !bytes.Equal(doc.Payload, mongoDoc.Payload) {
		return errors.New("document binary payload corrupted by backup/restore")
	}

	return c.checkIndex()
}

func (c *MongoConn) checkIndex() error {
	cur, err := c.client.Database(c.database).Collection(c.collection).Indexes().List(ctx)
	if err != nil {
		return err
	}

	var indexes []bson.M
	if err := cur.All(ctx, &indexes); err != nil {
		return err
	}

	for _, index := range indexes {
		if index["name"] == mongoIndexName {
			return nil
		}
	}

	return fmt.Errorf("index %q not found after restore, indexes: %v", mongoIndexName, indexes)
}
