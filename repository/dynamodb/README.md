# DynamoDB Repository Usage

This package provides a generic and type-safe repository for interacting with AWS DynamoDB. It is designed to reduce boilerplate code by simplifying the mapping between your application's models and a database-specific DAO (Data Access Object).

**Note on DynamoDB vs. Firestore:** Unlike the Firestore repository, the DynamoDB repository does not have a separate "Simple Path" (`NewRepository`) and "Advanced Path" (`NewRepositoryWithDAO`). Due to the nature of the AWS SDK and DynamoDB's requirement for top-level key attributes, you **must always define a custom DAO struct**. The repository provides a `BaseDAO` to embed, simplifying this process significantly.

---

## Simple DAO Path

This is the most common and straightforward approach for DynamoDB. You define a minimal DAO struct that contains your table's key fields and embeds `dynamodb.BaseDAO` to hold your model and audit fields.

### Example

**1. Define your model:**

```go
package main

type UserProfile struct {
    ID    string
    Name  string
    Email string
}

func (u UserProfile) IsEmpty() bool {
    return u.ID == ""
}
```

**2. Define your minimal DAO:**

Define your hash and sort key fields. Use `dynamodbav` tags to match the field names to your table's schema. Embed `dynamodb.BaseDAO[YourModelType]`.

```go
import "github.com/suared/core/v2/repository/dynamodb"

// UserProfileDAO defines the database structure.
type UserProfileDAO struct {
    dynamodb.BaseDAO[UserProfile]
    PartitionKey string `dynamodbav:"PK"` // Maps to the table's hash key
    SortKey      string `dynamodbav:"SK"` // Maps to the table's sort key
}
```

**3. Initialize the repository with a `DaoMapping`:**

Provide functions that teach the repository how to map between your model and your DAO.

```go
import (
    "context"
    "github.com/suared/core/v2/repository"
    "github.com/suared/core/v2/repository/dynamodb"
    "github.com/suared/core/v2/security"
)

func main() {
    // Standard repository configuration for DynamoDB
    config := repository.NewBasicConfig("UserProfileRepo")
    config.AddEntry("backend", "dynamoDB")
    config.AddEntry("table", "my-app-table")
    config.AddEntry("hashKeyName", "PK")
    config.AddEntry("sortKeyName", "SK")
    // ... add other config entries (region, endpoint, etc.)

    // Define the mapping
    mapping := dynamodb.DaoMapping[UserProfile, UserProfileDAO]{
        // Converts your model to your DAO
        FromModelFunc: func(ctx context.Context, model UserProfile) UserProfileDAO {
            // Example: Use the user's ID for the hash key and a static prefix for the sort key
            auth := security.GetAuth(ctx)
            return UserProfileDAO{
                BaseDAO:      dynamodb.BaseDAO[UserProfile]{Model: model},
                PartitionKey: "USER#" + auth.GetUser(),
                SortKey:      "PROFILE#" + model.ID,
            }
        },
        // Converts your DAO back to your model
        ToModelFunc: func(dao UserProfileDAO) UserProfile {
            return dao.Model
        },
        // Extracts the hash key from the DAO
        HashKeyFunc: func(dao UserProfileDAO) string { return dao.PartitionKey },
        // Extracts the sort key from the DAO
        SortKeyFunc: func(dao UserProfileDAO) string { return dao.SortKey },
    }

    // Create the repository
    repo, err := dynamodb.NewRepository(config, mapping)
    if err != nil {
        // handle error
    }

    // Now you can use the repo with your UserProfile model
    // repo.Insert(ctx, UserProfile{...})
}
```

---

## Advanced DAO Path

This follows the exact same pattern as the Simple DAO Path, but you add your own custom, database-only fields to your DAO struct.

### Example

Simply add more fields to your DAO definition. The repository mapping and initialization logic remains the same.

```go
import "github.com/suared/core/v2/repository/dynamodb"

// UserProfileAdvancedDAO adds custom fields for server-side logic.
type UserProfileAdvancedDAO struct {
    dynamodb.BaseDAO[UserProfile]
    PartitionKey   string   `dynamodbav:"PK"`
    SortKey        string   `dynamodbav:"SK"`
    SearchKeywords []string `dynamodbav:"searchKeywords"` // Custom field
    IsArchived     bool     `dynamodbav:"isArchived"`     // Custom field
}

// The DaoMapping and NewRepository calls would be updated to use
// UserProfileAdvancedDAO instead of UserProfileDAO.
```
