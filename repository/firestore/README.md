# Firestore Repository Usage

This package provides a generic and type-safe repository for interacting with Google Firestore. It is designed to reduce boilerplate code by offering two distinct pathways for data persistence: a simple path for direct model storage and an advanced path for when a custom database structure is required.

---

## Simple Path: Storing a Model Directly

This is the most common use case. You only need to define your model struct. The repository will automatically handle the creation of a standard database object that includes your model as a sub-field, along with audit fields (`CreatedBy`, `UpdatedBy`, etc.).

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

**2. Initialize the repository with a `ModelMapping`:**

```go
import (
    "github.com/suared/core/v2/repository"
    "github.com/suared/core/v2/repository/firestore"
)

func main() {
    // Standard repository configuration
    config := repository.NewBasicConfig("UserProfileRepo")
    config.AddEntry("backend", "firestore")
    config.AddEntry("projectID", "your-gcp-project-id")
    // ... add other config entries (endpoint for emulator, etc.)

    // Create the repository
    repo, err := firestore.NewRepository(config, firestore.ModelMapping[UserProfile]{
        Collection: "user_profiles",
        DocIDFunc:  func(m UserProfile) string { return m.ID },
    })

    if err != nil {
        // handle error
    }

    // Now you can use the repo with your UserProfile model
    // repo.Insert(ctx, UserProfile{...})
}
```

---

## Advanced Path: Using a Custom DAO

Use this path when you need to store database-only fields that don't belong in your application model (e.g., search keywords, denormalized counters, legacy fields).

This involves defining your own DAO (Data Access Object) struct where you embed the provided `firestore.BaseDAO` and add your custom fields.

### Example

**1. Define your model (same as before):**

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

**2. Define your custom DAO:**

Embed `firestore.BaseDAO[YourModelType]` and add your custom fields.

```go
import (
    "strings"
    "github.com/suared/core/v2/repository/firestore"
)

// UserProfileDAO is our custom struct for the database.
type UserProfileDAO struct {
    firestore.BaseDAO[UserProfile] // Embed the generic BaseDAO
    SearchKeywords []string `firestore:"searchKeywords"`
    AccountStatus  string   `firestore:"accountStatus"`
}
```

**3. Initialize the repository with a `DaoMapping`:**

Provide functions that teach the repository how to map between your model and your DAO.

```go
import (
    "context"
    "github.com/suared/core/v2/repository"
    "github.com/suared/core/v2/repository/firestore"
)

func main() {
    config := repository.NewBasicConfig("UserProfileRepo")
    // ... setup config ...

    repo, err := firestore.NewRepositoryWithDAO(config,
        firestore.DaoMapping[UserProfile, UserProfileDAO]{
            Collection: "user_profiles_advanced",
            DocIDFunc: func(dao UserProfileDAO) string {
                return dao.Model.ID // The doc ID comes from the model inside the DAO
            },
            // Converts your model to your DAO
            FromModelFunc: func(ctx context.Context, model UserProfile) UserProfileDAO {
                return UserProfileDAO{
                    BaseDAO:        firestore.BaseDAO[UserProfile]{Model: model},
                    SearchKeywords: strings.Split(strings.ToLower(model.Name), " "),
                    AccountStatus:  "active",
                }
            },
            // Converts your DAO back to your model
            ToModelFunc: func(dao UserProfileDAO) UserProfile {
                return dao.Model
            },
        },
    )
    // ...
}
```
