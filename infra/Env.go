package infra

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

//Go treats these as singleton so do not need to code singleton manually here
func init() {
	//fmt.Println("This will get called on main initialization only")
	setEnvironmentVariables()
	loadEnvironmentVariables()

	runtime.GOMAXPROCS(runtime.NumCPU())

}

func setEnvironmentVariables() {
	//Tells AWS SDK Client to use its default config chain - ~/.aws/config
	//Overrides can be done at various levels including the code accessors
	os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
}

//Assumes .env file will be in either working directory or infra/ sub-directory.  Will check up to 2 previous directories and then give up
//The environment variable PROCESS_ENV_CHECK is used to validate this is the right kind of env file if an invalid file is found

func loadEnvironmentVariables() {
	// Directly from the godotenv README here --> https://github.com/joho/godotenv
	// Existing envs take precendence of envs that are loaded later.
	// e.g. load prod.local followed by .env.local followed by .env.prod followed by .env
	// NOTE:  I am removing the interim one, seems like overkill for non/test, will bring back only if needed
	//fmt.Printf("Starting up, model dir set to: %v", os.Getenv("PROCESS_MODEL_TESTFILE_DIR"))
	//var PATHPREFIX string

	// Do a search for the env file so that we can keep the src structure clean and test at different sub-package levels without impacting the relative check here
	dir, _ := os.Getwd()

	//as soon as a .env file is found, the env base dir search will stop
	//Give up if have tried this for up to 4 previous directories, note that accidental .env will impact this so will both a) log and b) exit with a panic depending on the situation

	newdir, found := checkForEnvFile(dir)
	if !found {
		newdir, found = checkForEnvFile(newdir)
	}

	if !found {
		newdir, found = checkForEnvFile(newdir)
	}

	if !found {
		newdir, found = checkForEnvFile(newdir)
	}

	if !found {
		panic("Env Setup: Unable to find environment variable file from base: " + dir)
	}

	//Intentional startup print
	fmt.Println("environment directory is " + newdir)

	//Start processing environment files using the standard logic
	env := os.Getenv("PROCESS_ENV")
	if "" == env || env == "dev" {
		env = "development"
	}

	godotenv.Load(filepath.Join(newdir, ".env."+env+".local"))

	//fmt.Printf("After initial load, model dir set to: %v", os.Getenv("PROCESS_MODEL_TESTFILE_DIR"))
	/*
		if "test" != env {
			godotenv.Load(".env.local")
		}
	*/

	godotenv.Load(filepath.Join(newdir, "/.env."+env))

	//fmt.Printf("after .env.<env> point, model dir set to: %v", os.Getenv("PROCESS_MODEL_TESTFILE_DIR"))
	godotenv.Load(filepath.Join(newdir, ".env")) // The Original .env

	//fmt.Println("attempted to load: " + filepath.Join(newdir, ".env"))

	testVal := os.Getenv("PROCESS_ENV_CHECK")
	if testVal == "" {
		panic("Env Setup: Invalid environment variable file, missing: PROCESS_ENV_CHECK, received: " + testVal + ", otherval: " + os.Getenv("PROCESS_MODEL_TESTFILE_DIR"))
	}
}

func fileExists(filename string) bool {
	//log.Printf("checking file %v", filename)
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// with return the environment directory or the previous directory in the path along with a bool to indicate if found in which case return director is the directory where found
// or false if not found and the previous directory was returned
func checkForEnvFile(dir string) (string, bool) {
	//log.Printf("checking directory %v", dir)
	if fileExists(filepath.Join(dir, ".env")) {
		return dir, true
	} else if fileExists(filepath.Join(dir, "infra", ".env")) {
		//	log.Printf("checking directory %v/infra", dir)
		return filepath.Join(dir, "infra"), true
	} else {
		//	log.Println(".env not found")
		return filepath.Join(dir, ".."), false
	}
}
