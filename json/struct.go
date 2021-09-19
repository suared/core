package json

import (
	"encoding/json"
	"log"
)

func StructAsBytes[T any](theStruct T) ([]byte, error) {
	data, err := json.Marshal(theStruct)
	return data, err
}

func BytesAsStruct[T any](data []byte, theStruct T) (T, error) {
	err := json.Unmarshal(data, &theStruct)
	if err != nil {
		log.Printf("Error unmarshaling file: %v", err)
		return theStruct, err
	}
	return theStruct, nil
}

/*
func writeAsJson[T any](filename string, thetype T) {
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Received error on create file e: %v", err)
		panic(err)
	}
	defer f.Close()

	data, err := json.Marshal(thetype)
	f.Write(data)
	f.Sync()
}

func readAsJson[T any](filename string, thestruct T) *T {
		f, err := os.Open(filename)
	
		data, err := ioutil.ReadAll(f)
		if err != nil {
			log.Printf("Received error on open file e: %v", err)
			panic(err)
		}
		defer f.Close()
		
		err = json.Unmarshal(data, &thestruct)
		if err != nil {
			log.Printf("Error unmarshaling file: %v", err)
		}
		return &thestruct

}
*/