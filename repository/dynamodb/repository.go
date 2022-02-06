package dynamodb

import (
    "context"
    "fmt"
    "log"

    _ "github.com/suared/core/infra"
    "github.com/suared/core/repository"
)

//<<Repository>> - Database interface to the X table
type Repository[dao DAO, usermodel ModelObject] struct {
    config  repository.Config
    session repository.Session
	baseDAO dao
	baseModel usermodel
}

//Config - Returns the current configuration
func (repo *Repository[dao, usermodel]) Config() repository.Config {
    return repo.config
}

//DAO - Returns a DAO associated with this repository from a model object
func (repo *Repository[dao, usermodel]) DAO(ctx context.Context, userModel usermodel) (dao, error) {
	localdao := repo.baseDAO.New()
	localdao.SetUser(ctx)
	localdao.Setup(userModel)

	/* TODO: Move to DAO.Setup(model)
    localdao.theModel = userModel

    if zipme == true {
        localdao.zipData = ziptools.GetGzipDataFromStruct(userModel)
    }
	*/
    return localdao.(dao), nil
}

//Insert - Sample of a basic insert method with validation
func (repo *Repository[dao, usermodel]) Insert(ctx context.Context, userModel usermodel) error {
    // Populate the Data object First //  active?, audit?
    localdao, err := repo.DAO(ctx, userModel)

    if err != nil {
        log.Printf("Unable to Insert DAO, error Getting DAO: %v", err)
        return err
    }

    // Repository layer is responsible for validating auth rules
    err = ValidAction(ctx, "insert", localdao)
    if err != nil {
        return err
    }

    return InsertOrUpdate(ctx, repo, localdao)
}

//Update - Sample of updating a DB entry
func (repo *Repository[dao, usermodel]) Update(ctx context.Context, userModel usermodel) error {
    localdao, err := repo.DAO(ctx, userModel)
    if err != nil {
        log.Printf("Unable to Update, error getting DAO, err: %v", err)
        return err
    }

    err = ValidAction(ctx, "update", localdao)
    if err != nil {
        return err
    }

    return InsertOrUpdate(ctx, repo, localdao)

}

//Delete - Sample of deleting a DB entry
func (repo *Repository[dao, usermodel]) Delete(ctx context.Context, template usermodel) error {
    localdao, err := repo.DAO(ctx, template)
    if err != nil {
        log.Printf("Unable getting Dao in Delete, err: %v", err)
        return err
    }

    return Delete(ctx, repo, localdao)
}

//Select - Sample of a get all by hashkey
func (repo *Repository[dao, usermodel]) Select(ctx context.Context, template usermodel, strFilterVals ...string) ([]usermodel, error) {
    localdao, err := repo.DAO(ctx, template)
    if err != nil {
        log.Printf("Unable to Select, error getting DAO, err: %v", err)
        return nil, err
    }

    result, err := Select(ctx, repo, localdao, strFilterVals...)

    var outputList []usermodel
    //since the search is for user, validation only needs to occur on one item..
    var validated bool
    for i := range result {
        resultDAO := result[i]
        //Check once only...
        if !validated {
            //ignore if no result/ empty
            if resultDAO.HashKey() != "" {
                err = ValidAction(ctx, "selectAll", resultDAO)

                if err != nil {
                    return []usermodel{}, err
                }
            }
        }
        resultItem := result[i].GetModel()
        outputList = append(outputList, resultItem.(usermodel))
        //log.Printf("output list: %v", outputList)
    }

    return outputList, err

}

//SelectOne - Returns one model object, can be empty if no results
func (repo *Repository[dao, usermodel]) SelectOne(ctx context.Context, template usermodel) (usermodel, error) {
    var localmodel usermodel
    localmodel = *new(usermodel)
    localdao, err := repo.DAO(ctx, template)
    if err != nil {
        log.Printf("Unable to SelectOne, error getting DAO, err: %v", err)
        return template, err
    }

	//change this to return the specific type - deleted the convesion with this assumption so will error at first
    result, err := SelectOne(ctx, repo, localdao)

    resultItem := result.GetModel().(usermodel)

    if resultItem.IsEmpty() {
        err = ValidAction(ctx, "selectOne", result)

        if err != nil {
            return localmodel, err
        }
    }

    return resultItem, err

}

//SetSession - enables the library to store/ reuse the session for efficiency vs. creating new on each call
func (repo *Repository[dao, usermodel]) SetSession(session repository.Session) {
    repo.session = session
}

//Session - Returns the session associated with this repository
func (repo *Repository[dao, usermodel]) Session() repository.Session {
    return repo.session
}

//NewCategoryRepository - Initializes a sample repository with config values set
func NewRepository[dao DAO, usermodel ModelObject](configMap repository.Config, localdao dao, localmodel usermodel) (*Repository[dao, usermodel], error) {
    repo := &Repository[dao, usermodel]{}
	repo.baseDAO = localdao
	repo.baseModel = localmodel
    repo.config = configMap

    //Convert the config into an initialized dynamoo table
    _, err := CreateTable(repo)
    if err != nil {
        return nil, fmt.Errorf("Unable to initialize category database session, received error: %v", err)
    }

    return repo, nil
}
