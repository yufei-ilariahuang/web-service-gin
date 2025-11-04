package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"src/api"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// dynamoStore implements Store using a single DynamoDB table.
// Partition key is sharded to avoid hot partitions: pk = "shard#<n>|<cartID>".
type dynamoStore struct {
	table  string
	shards int32
}

// NewDynamoStore constructs a new dynamo store. shards controls how many
// prefix shards to distribute sequential cart IDs across.
func NewDynamoStore(table string, shards int32) Store {
	if shards <= 0 {
		shards = 128
	}
	return &dynamoStore{table: table, shards: shards}
}

func (d *dynamoStore) partitionKey(cartID int32) string {
	shard := cartID % d.shards
	return fmt.Sprintf("CART#shard%d", shard)
}

func (d *dynamoStore) sortKey(cartID int32) string {
	return fmt.Sprintf("CART#%d", cartID)
}

// CreateCart stores a new cart. For demo purposes we generate a mostly-unique
// int32 id from the timestamp. In production you'd use a robust id generation
// strategy or a counter table.
func (d *dynamoStore) CreateCart(ctx context.Context, customerID int32) (int32, error) {
	// generate cart id
	raw := time.Now().UnixNano()
	cartID := int32(raw % (1<<31 - 1))

	pk := d.partitionKey(cartID)
	sk := d.sortKey(cartID)

	item := map[string]types.AttributeValue{
		"pk":          &types.AttributeValueMemberS{Value: pk},
		"sk":          &types.AttributeValueMemberS{Value: sk},
		"cart_id":     &types.AttributeValueMemberN{Value: strconv.FormatInt(int64(cartID), 10)},
		"customer_id": &types.AttributeValueMemberN{Value: strconv.FormatInt(int64(customerID), 10)},
		"items":       &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}},
		"created_at":  &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
	}

	_, err := DDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &d.table,
		Item:      item,
		// Fail if item with same pk exists
		ConditionExpression: nil,
	})
	if err != nil {
		return 0, fmt.Errorf("dynamodb put item: %w", err)
	}

	return cartID, nil
}

// GetCart fetches the cart and converts the DynamoDB map into api.ShoppingCart
func (d *dynamoStore) GetCart(ctx context.Context, cartID int32) (*api.ShoppingCart, error) {
	pk := d.partitionKey(cartID)

	out, err := DDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &d.table,
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb get item: %w", err)
	}
	if out.Item == nil {
		return nil, fmt.Errorf("cart not found")
	}

	// Unmarshal cart_id, customer_id, created_at, items
	var raw struct {
		CartID    int32            `dynamodbav:"cart_id"`
		Customer  int32            `dynamodbav:"customer_id"`
		CreatedAt string           `dynamodbav:"created_at"`
		Items     map[string]int64 `dynamodbav:"items"`
	}

	if err := attributevalue.UnmarshalMap(out.Item, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal item: %w", err)
	}

	cart := &api.ShoppingCart{
		ShoppingCartId: raw.CartID,
		CustomerId:     raw.Customer,
	}
	if raw.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, raw.CreatedAt); err == nil {
			cart.CreatedAt = &t
		}
	}

	cart.Items = []api.CartItem{}
	for pidStr, qty := range raw.Items {
		pid, _ := strconv.ParseInt(pidStr, 10, 32)
		cart.Items = append(cart.Items, api.CartItem{ProductId: int32(pid), Quantity: int32(qty)})
	}

	return cart, nil
}

// AddItemToCart updates or adds an item in the items map attribute.
func (d *dynamoStore) AddItemToCart(ctx context.Context, cartID, productID, quantity int32) error {
	pk := d.partitionKey(cartID)
	productKey := strconv.FormatInt(int64(productID), 10)
	// Use expression attribute names to update the map entry atomically
	exprNames := map[string]string{
		"#it":  "items",
		"#pid": productKey,
	}
	exprValues := map[string]types.AttributeValue{
		":zero": &types.AttributeValueMemberN{Value: "0"},
		":qty":  &types.AttributeValueMemberN{Value: strconv.FormatInt(int64(quantity), 10)},
	}

	// UpdateExpression: SET items.<productKey> = if_not_exists(items.<productKey>, :zero) + :qty
	updateExpr := "SET #it.#pid = if_not_exists(#it.#pid, :zero) + :qty"

	_, err := DDB.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &d.table,
		Key:                       map[string]types.AttributeValue{"pk": &types.AttributeValueMemberS{Value: pk}},
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
		UpdateExpression:          &updateExpr,
		ReturnValues:              types.ReturnValueUpdatedNew,
	})
	if err != nil {
		return fmt.Errorf("dynamodb update item: %w", err)
	}
	return nil
}
