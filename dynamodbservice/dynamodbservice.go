package dynamodbservice

import (
   "../config"
    "fmt"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
    "sort"
)

/*
Product - Go object representation of items that will be managed by the app.
*/
type Product struct {
    Id    int `json:"id"`
    Name  string
    Price float64
}

func (p Product) String() string {
    return fmt.Sprintf("<(Id: %v) {%v} @ %v>", p.Id, p.Name, p.Price)
}

// Products - wrapper for the DynamoDB Go type that will allow local methods to be called from DynamoDB instances.
type Products struct {
    *dynamodb.DynamoDB
}

// Items - global DynamoDB instance.
var Items Products

// TableName - name for the table that will serve as the DynamoDB instance.
const TableName = "Products"

// IdAttribute - attribute name for the partition key.
const IdAttribute = "id"

// GetAll - responds with all of the Products in price-descending order.
func (db Products) GetAll() ([]Product, error) {
    // Price-descending sort
    var temp []Product

    result, err := Items.Scan(&dynamodb.ScanInput{TableName: aws.String(TableName)})
    if err != nil {
        return nil, fmt.Errorf("Query GetAll failed:\n%v", err)
    }

    err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &temp)
    if err != nil {
        return nil, fmt.Errorf("Unmarshalling GetAll failed:\n%v", err)
    }

    // Manually sort the results to get a Price-descending sort
    sort.Slice(temp, func(i, j int) bool { return temp[i].Price > temp[j].Price })

    return temp, nil
}

// AddProduct - adds a new Product to the database.
func (db *Products) AddProduct(newProduct Product) error {
    data, err := dynamodbattribute.MarshalMap(newProduct)
    if err != nil {
        return fmt.Errorf("AddProduct -> Error marshalling product: %v", err)
    }

    // Setup the insert criteria.
    item := &dynamodb.PutItemInput{
        Item:      data,
        TableName: aws.String(TableName),
    }

    // Insert the new Product into the database.
    _, err = Items.PutItem(item)
    if err != nil {
        return fmt.Errorf("AddProduct -> New product could not be added: %v", err)
    }

    return nil
}

// Initialize - sets up the database when the app is run for the first time.
func Initialize() error {
    // Initialize the AWS session.
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(config.Settings.AWS.Region),
        Credentials: credentials.NewStaticCredentials(config.Settings.AWS.AccessKeyID, config.Settings.AWS.SecretAccessKeyID, "")},
    )
    if err != nil {
        return fmt.Errorf("INITIALIZATION ERROR: %v", err)
    }

    // Initialize the DynamoDB instance.
    Items = Products{dynamodb.New(sess, aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))}
    Items.listTables()

    tableExists, err := Items.tableExists(TableName)
    if err != nil {
        return fmt.Errorf("INITIALIZATION ERROR: %v", err)
    }

    if !tableExists {
        createTable()
    } else {
        fmt.Println("Table already exists!")
    }

    return nil
}

// createTable - local helper function that creates the Products DynamoDB table.
func createTable() error {
    fmt.Println("Creating table...")

    // Setup table create criteria.
    input := &dynamodb.CreateTableInput{
        TableName: aws.String(TableName),
        KeySchema: []*dynamodb.KeySchemaElement{
            {
                AttributeName: aws.String(IdAttribute), KeyType: aws.String("HASH"),
            },
        },
        AttributeDefinitions: []*dynamodb.AttributeDefinition{
            {
                AttributeName: aws.String(IdAttribute), AttributeType: aws.String("N"),
            },
        },
        ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
            ReadCapacityUnits: aws.Int64(10), WriteCapacityUnits: aws.Int64(10),
        },
    }

    // Create the table.
    if _, err := Items.CreateTable(input); err != nil {
        fmt.Println("Error during CreateTable:")
         fmt.Printf("%v", err)
    }

    fmt.Printf("Table '%v' successfully created!\n", TableName)

    return nil
}


func (db *Products) tableExists(name string) (bool, error) {
    result, err := db.ListTables(&dynamodb.ListTablesInput{})

    if err != nil {
        fmt.Println("Error during ListTables:")
        return false, fmt.Errorf("%v", err)
    }

    for _, n := range result.TableNames {
        if *n == name {
            return true, nil
        }
    }

    return false, nil
}

func (db *Products) listTables() error {
    result, err := db.ListTables(&dynamodb.ListTablesInput{})

    if err != nil {
        fmt.Println("Error during ListTables:")
        return fmt.Errorf("%v", err)
    }

    fmt.Println("Tables:")
    fmt.Println("")

    for _, n := range result.TableNames {
        fmt.Println(*n)
    }
    return nil
}