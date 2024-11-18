package db

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBConnection struct {
	ConnectionString string
	Client           *mongo.Client
}

func NewMongoDBConnection(cs string) *MongoDBConnection {
	return &MongoDBConnection{ConnectionString: cs}
}

func (mc *MongoDBConnection) Connect() error {
	clientOptions := options.Client().ApplyURI(mc.ConnectionString)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return fmt.Errorf("Error connecting to MongoDB: %v", err)
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		return fmt.Errorf("Error pinging MongoDB: %v", err)
	}

	mc.Client = client
	return nil
}

func (mc *MongoDBConnection) Disconnect() error {
	if mc.Client != nil {
		return mc.Client.Disconnect(context.TODO())
	}
	return nil
}

func (mc *MongoDBConnection) GetTables() ([]string, error) {
	databases, err := mc.Client.ListDatabaseNames(context.TODO(), bson.M{})
	if err != nil {
		return nil, fmt.Errorf("Error listing databases: %v", err)
	}

	var collections []string
	for _, dbName := range databases {
		cols, err := mc.Client.Database(dbName).ListCollectionNames(context.TODO(), bson.M{})
		if err != nil {
			return nil, fmt.Errorf("Error listing collections in database %s: %v", dbName, err)
		}
		for _, col := range cols {
			collections = append(collections, fmt.Sprintf("%s.%s", dbName, col))
		}
	}

	return collections, nil
}

func (mc *MongoDBConnection) GetTableColumns(tableName string) (map[string]string, error) {
	parts := strings.Split(tableName, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid table name format. Expected 'database.collection'")
	}
	dbName, colName := parts[0], parts[1]

	collection := mc.Client.Database(dbName).Collection(colName)
	cursor, err := collection.Find(context.TODO(), bson.M{}, options.Find().SetLimit(1))
	if err != nil {
		return nil, fmt.Errorf("Error finding documents in collection %s: %v", tableName, err)
	}
	defer cursor.Close(context.TODO())

	var result bson.M
	if cursor.Next(context.TODO()) {
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("Error decoding document: %v", err)
		}
	}

	columns := make(map[string]string)
	for key, value := range result {
		columns[key] = fmt.Sprintf("%T", value)
	}

	return columns, nil
}

func (mc *MongoDBConnection) ExecuteQuery(query string) (string, error) {
	// For simplicity, let's assume the query is in the format: "db.dbname.collection.find({})"
	parts := strings.Split(query, ".")
	if len(parts) < 3 {
		return "", fmt.Errorf("Invalid query format. Expected 'db.collection.operation'")
	}
	dbName, colName, operation := parts[1], parts[2], parts[3]

	collection := mc.Client.Database(dbName).Collection(colName)

	switch {
	case strings.HasPrefix(operation, "find"):
		filter := bson.M{}
		if len(parts) > 3 {
			if err := bson.UnmarshalExtJSON([]byte(strings.Join(parts[3:], ".")), true, &filter); err != nil {
				return "", fmt.Errorf("Error parsing filter: %v", err)
			}
		}
		cursor, err := collection.Find(context.TODO(), filter)
		if err != nil {
			return "", fmt.Errorf("Error executing find operation: %v", err)
		}
		defer cursor.Close(context.TODO())

		var results []bson.M
		if err := cursor.All(context.TODO(), &results); err != nil {
			return "", fmt.Errorf("Error decoding results: %v", err)
		}

		var output strings.Builder
		for _, result := range results {
			output.WriteString(fmt.Sprintf("%v\n", result))
		}
		return output.String(), nil

	default:
		return "", fmt.Errorf("Unsupported operation: %s", operation)
	}
}

func (pc *MongoDBConnection) GetType() DatabaseType {
	return DatabaseType{
		Type:    "MongoDB",
		Command: "mongodb query",
	}
}
