package file

import (
	"io/ioutil"
	"log"
	"os"
)

func WriteAndClose(filename string, data []byte) error {
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Received error on create file e: %v", err)
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		log.Printf("Partial write error on file e: %v", err)
		return err
	}

	err = f.Sync()
	if err != nil {
		log.Printf("Error flusing data to file e: %v", err)
		return err
	}

	return err
}

func ReadAndCLose(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		log.Printf("Received error on open file e: %v", err)
		return nil, err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("Received error reading from file e: %v", err)
		return nil, err
	}
	defer f.Close()

	return data, nil

}
