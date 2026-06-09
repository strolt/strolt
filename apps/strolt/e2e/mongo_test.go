package e2e_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
	s.Require().NoError(strolt("backup", "--service", "e2e", "--task", "mongo", "--y"))

	s.Require().NoError(s.c.drop())

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e", "mongo", "restic-mongo")
	s.Require().NoError(err)

	s.NoError(strolt("restore", "--service", "e2e", "--task", "mongo", "--destination", "restic-mongo", "--snapshot", latestSnapshotID, "--y"))
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
	return c.client.Database(c.database).CreateCollection(ctx, c.collection)
}

func (c *MongoConn) insertData() error {
	collection := c.client.Database(c.database).Collection(c.collection)
	if _, err := collection.InsertOne(ctx, user); err != nil {
		return err
	}

	return nil
}

func (c *MongoConn) checkValidData() error {
	collection := c.client.Database(c.database).Collection(c.collection)

	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return err
	}

	var users []User
	if err := cur.All(ctx, &users); err != nil {
		return err
	}

	if len(users) == 0 {
		return errors.New("not found records")
	}

	if len(users) != 1 {
		return errors.New("count records > 1")
	}

	if users[0].Username != user.Username || users[0].Password != user.Password || users[0].ID != user.ID {
		return errors.New("record not match with mock")
	}

	return nil
}
