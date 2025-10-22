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
	ctx        *context.Context
	client     *mongo.Client
	database   string
	collection string
}

func (s *MongoSuite) SetupSuite() {
	c, err := mongoConnect()
	s.NoError(err)
	s.c = c
}

func (s *MongoSuite) TearDownSuite() {
	s.NoError(s.c.close())
}

func (s *MongoSuite) BeforeTest(suiteName, testName string) {
	s.NoError(s.c.drop())
	s.NoError(s.c.createCollection())
	s.NoError(s.c.insertData())
	s.NoError(s.c.checkValidData())
}

func (s *MongoSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.c.checkValidData())
}

func (s *MongoSuite) TestMongo() {
	s.NoError(strolt("backup", "--service", "e2e", "--task", "mongo", "--y"))

	s.NoError(s.c.drop())

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e", "mongo", "restic-mongo")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e", "--task", "mongo", "--destination", "restic-mongo", "--snapshot", latestSnapshotID, "--y"))
}

//nolint:thelper
func MongoSuiteTest(t *testing.T) {
	tt := timeTook("MongoSuiteTest")

	suite.Run(t, new(MongoSuite))
	tt.stop()
}

func mongoConnect() (*MongoConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	port, err := containerManager.GetMongoPort()
	if err != nil {
		return nil, err
	}

	uri := "mongodb://localhost:" + port
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))

	return &MongoConn{
		ctx:        &ctx,
		client:     client,
		database:   "strolt",
		collection: "strolt",
	}, err
}

func (c *MongoConn) close() error {
	return c.client.Disconnect(*c.ctx)
}

func (c *MongoConn) drop() error {
	return c.client.Database(c.database).Drop(context.TODO())
}

func (c *MongoConn) createCollection() error {
	return c.client.Database(c.database).CreateCollection(context.TODO(), c.collection)
}

func (c *MongoConn) insertData() error {
	collection := c.client.Database(c.database).Collection(c.collection)
	if _, err := collection.InsertOne(context.TODO(), user); err != nil {
		return err
	}

	return nil
}

func (c *MongoConn) checkValidData() error {
	collection := c.client.Database(c.database).Collection(c.collection)

	cur, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return err
	}

	var users []User
	if err := cur.All(context.TODO(), &users); err != nil {
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
