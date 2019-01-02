# core
Go Lang reusable code bits for private projects.  The intent here is to be a library for my projects or others' if they find the patterns helpful.

#Usage examples:

###Setup environment variables
`import	_ "github.com/suared/core/infra"`

The side-effects here are:
1. load all the environments variables so they are available from os.Getenv standard system call in the users of the library.
2. setup shared project defaults

Performs .env file search in current and previous sub-directories going back up to two sub-directories.  Does expect all .env files to be in one directory when found.  As it searches for the environment directory it also looks for an infra/ sub-directory which I am using in my projects to organize all environment info.  Prioritizes environment files in this order:
1. .env.<environmentname>.local
2. .env.<environmentname>
3. .env
See: github.com/joho/godotenv

Shared project defaults are also here:
1. Set GOMAXPROCS
2. Set Default AWS security properties 

All defaults here will be low weight, mostly environment variables so that it is generailly applicable for easy setup.  Can be referenced from all classes and will only run once


### Setup Http Listener for APIs
`
//import "github.com/suared/core/api"
api.StartHTTPListener(apiRoutes{})`

In the example above, apiRoutes is a struct that implements the Config interface so that the API "mains()" just have to call the above to get started.  All shared middleware and healthchecks will be setup here.

### Infrastructure shared scripts
This is organized by environment and only a dynamodb start is there for now and will expand as needed if not specific to one api

### Starter Templates
Any common copy/ pastes to start a new API will be here and can eventually include generators.  For now this has:
1. .gitignore - added to the base directory
2. Makefile - added to the base directory
3. .env - added to the infra directory

### Common tool needs
Over time as shared tools are identified they will be moved here from sub-projects. For now, this has:
1. ziptools
`// import "github.com/suared/core/ziptools"
err := ziptools.GetGzipData(&buf, processdata)
// -- and/ or --
err := ziptools.GetGunzipData(&buf, StructToZip)`
These zippers implement a shared pool for reuse across the runtime


