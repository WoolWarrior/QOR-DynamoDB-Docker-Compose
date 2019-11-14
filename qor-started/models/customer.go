package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"

	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/roles"

	"github.com/jinzhu/gorm"
)

// Customer data structure
type Customer struct {
	// ID          uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	ID            string
	Email         string
	CreatedAtTime time.Time
	UpdatedAtTime time.Time
	DeletedAt     *time.Time `sql:"index"`
	Surname       string
	FirstName     string
	PhoneNumber   string
	Description   string
}

// DeepCopy method is to copy interface object
func DeepCopy(source interface{}, destination interface{}) {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(source)
	json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&destination)
}

// a variable to copy interface object
var filterCustomers []Customer

// ConfigureQorResourceDynamoDB is to configure the resource to DynamoDB CRUD
func ConfigureQorResourceDynamoDB(r resource.Resourcer) {

	// Configure resource with DynamoDB
	tableName := "Customers"
	config := &aws.Config{
		Endpoint: aws.String("http://dynamodb:8000"), 
		// Endpoint: aws.String("http://localhost:8000"),
	}

	// Create DynamoDB client
	svc := dynamodb.New(session.New(), config)

	customer, ok := r.(*admin.Resource)
	if !ok {
		panic(fmt.Sprintf("Unexpected resource! T: %T", r))
	}

	// Open one item from the database
	customer.FindOneHandler = func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
		fmt.Println("FindOneHandler")

		if customer.HasPermission(roles.Read, context) {

			customerIDString := context.ResourceID

			// input to define the data to
			input := &dynamodb.GetItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"ID": {
						S: aws.String(customerIDString),
					},
				},
				TableName: aws.String(tableName),
			}

			resultFromDB, err := svc.GetItem(input)
			fmt.Println("Found item resultFromDB: ", resultFromDB)

			dbCustomer := Customer{}
			fmt.Println("Found item: ", dbCustomer)
			err = dynamodbattribute.UnmarshalMap(resultFromDB.Item, &dbCustomer)

			if err != nil {
				panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
			}

			DeepCopy(dbCustomer, &result)

			fmt.Println("Found item: ", dbCustomer)

			return err
		}

		return roles.ErrPermissionDenied
	}

	// Show found items
	customer.FindManyHandler = func(result interface{}, context *qor.Context) error {
		fmt.Println("FindManyHandler")

		if customer.HasPermission(roles.Read, context) {

			if len(filterCustomers) == 0 {
				input := &dynamodb.ScanInput{
					TableName: aws.String(tableName),
				}

				resultFromDB, err := svc.Scan(input)

				if err != nil {
					fmt.Println("Query API call failed:")
					fmt.Println((err.Error()))
					os.Exit(1)
				}

				// create a slice to store result
				dbCustomers := make([]Customer, 0)
				numResult := 0

				for _, i := range resultFromDB.Items {
					dbcustomersTMP := Customer{}
					err = dynamodbattribute.UnmarshalMap(i, &dbcustomersTMP)
					if err != nil {
						fmt.Println("Got error unmarshalling:")
						fmt.Println(err.Error())
						os.Exit(1)
					}
					dbCustomers = append(dbCustomers, dbcustomersTMP)
					numResult++

				}

				DeepCopy(dbCustomers, &result)

				fmt.Println("Found", numResult, "result(s) as below: ", dbCustomers)

				return err
				// return nil
			} else {
				dbCustomers := filterCustomers

				numResult := len(dbCustomers)

				// dbCustomers = context.DB.Value

				DeepCopy(dbCustomers, &result)

				fmt.Println("Found", numResult, "result(s) as below: ", dbCustomers)

				return nil
			}

		}

		return roles.ErrPermissionDenied
	}

	customer.SaveHandler = func(result interface{}, context *qor.Context) error {
		fmt.Println("SaveHandler")
		if customer.HasPermission(roles.Create, context) || customer.HasPermission(roles.Update, context) {

			var customerTMP Customer

			DeepCopy(result, &customerTMP)

			newUUID, _ := uuid.NewRandom()

			if customerTMP.ID == "" {
				customerTMP.ID = newUUID.String()
				customerTMP.CreatedAtTime = time.Now()
			}
			customerTMP.UpdatedAtTime = time.Now()

			input := &dynamodb.UpdateItemInput{
				ExpressionAttributeNames: map[string]*string{
					"#E": aws.String("Email"),
					"#S": aws.String("Surname"),
					"#F": aws.String("FirstName"),
					"#P": aws.String("PhoneNumber"),
					"#D": aws.String("Description"),
					"#C": aws.String("CreatedAtTime"),
					"#U": aws.String("UpdatedAtTime"),
				},
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":email": {
						S: aws.String(customerTMP.Email),
					},
					":surname": {
						S: aws.String(customerTMP.Surname),
					},
					":firstname": {
						S: aws.String(customerTMP.FirstName),
					},
					":phonenumber": {
						S: aws.String(customerTMP.PhoneNumber),
					},
					":description": {
						S: aws.String(customerTMP.Description),
					},
					":createdattime": {
						S: aws.String(customerTMP.CreatedAtTime.Format(time.RFC3339)),
					},
					":updatedattime": {
						S: aws.String(customerTMP.UpdatedAtTime.Format(time.RFC3339)),
					},
				},

				Key: map[string]*dynamodb.AttributeValue{
					"ID": {
						S: aws.String(customerTMP.ID),
					},
				},
				ReturnValues:     aws.String("UPDATED_NEW"),
				TableName:        aws.String(tableName),
				UpdateExpression: aws.String("SET #E =:email, #S =:surname, #F =:firstname, #P =:phonenumber, #D =:description, #C =:createdattime, #U =:updatedattime "),
			}

			_, err := svc.UpdateItem(input)

			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Successfully updated ", customerTMP)
			}

			return err

		}
		return roles.ErrPermissionDenied
	}

	customer.DeleteHandler = func(result interface{}, context *qor.Context) error {
		fmt.Println("DeleteHandler")
		if customer.HasPermission(roles.Delete, context) {
			// var dbCustomerTMP Customer
			// dbCustomerTMP.ID, _ = uuid.Parse(context.ResourceID)

			customerIDString := context.ResourceID

			input := &dynamodb.DeleteItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"ID": {
						S: aws.String(customerIDString),
					},
				},
				TableName: aws.String(tableName),
			}

			_, err := svc.DeleteItem(input)
			if err != nil {
				fmt.Println("Got error calling DeleteItem")
				fmt.Println(err.Error())
				return nil
			}

			fmt.Println("Deleted ", customerIDString)

			return err
		}
		return roles.ErrPermissionDenied
	}

	customer.SearchAttrs("Email")

	oldSearchHandler := customer.SearchHandler

	// Keyword is the text field to be searched
	customer.SearchHandler = func(keyword string, context *qor.Context) *gorm.DB {
		fmt.Println("SearchHandler")

		filt := expression.Name("Email").Equal(expression.Value(keyword))
		proj := expression.NamesList(expression.Name("ID"), expression.Name("CreatedAtTime"), expression.Name("UpdatedAtTime"), expression.Name("Email"), expression.Name("PhoneNumber"), expression.Name("Surname"), expression.Name("FirstName"), expression.Name("Description"))
		expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()

		if err != nil {
			fmt.Println("Got error building expression:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		params := &dynamodb.ScanInput{
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
			ProjectionExpression:      expr.Projection(),
			TableName:                 aws.String(tableName),
		}

		resultFromDB, err := svc.Scan(params)
		if err != nil {
			fmt.Println("Query API call failed:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		dbCustomers := make([]Customer, 0)
		numResult := 0

		for _, i := range resultFromDB.Items {
			dbcustomersTMP := Customer{}
			err = dynamodbattribute.UnmarshalMap(i, &dbcustomersTMP)
			if err != nil {
				fmt.Println("Got error unmarshalling:")
				fmt.Println(err.Error())
				os.Exit(1)
			}
			dbCustomers = append(dbCustomers, dbcustomersTMP)
			numResult++

		}

		filterCustomers = dbCustomers

		fmt.Println("Found ", numResult, "result(s) as below: ", filterCustomers)

		return oldSearchHandler(keyword, context)
	}
}
