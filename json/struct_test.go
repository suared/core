package json

import (
	"testing"

	"github.com/suared/core/file"
	_ "github.com/suared/core/infra"
)

/*

Tests;
	- save 2 types of options
	- read 2 types without differentiating
	- do the same but add references // future

*/

//Structs to use for testing
type AnOption interface {
	GetID() string
	GetText() string
}
type BasicOption struct {
	ID   string `json:"ID"`
	Text string `json:"Text"`
}

func (opt *BasicOption) GetID() string {
	return opt.ID
}

func (opt *BasicOption) GetText() string {
	return opt.Text
}

type RandomOption struct {
	Random string `json:"Random"`
	BasicOption
}

func (opt *RandomOption) GetID() string {
	return "Random: " + opt.ID
}
func (opt *RandomOption) GetRandom() string {
	return opt.Random
}

//Tests...
func TestReadWriteGenerics(t *testing.T) {
	opt := BasicOption{ID: "1", Text: "basic"}
	opt2 := RandomOption{Random: "Whatever"}
	opt2.ID = "2"
	opt2.Text = "Hmmm...."

	//Store in files - TODO: add convenience to file or json to do both these steps together?  which?
	data, err := StructAsBytes(opt)
	if err != nil {
		t.Errorf("Could not conert opt to bytes: %v", err)
	}
	err = file.WriteAndClose("temp_optFile.json", data)
	if err != nil {
		t.Errorf("Could not save opt to file: %v", err)
	}

	data, err = StructAsBytes(opt2)
	if err != nil {
		t.Errorf("Could not conert opt to bytes: %v", err)
	}

	err = file.WriteAndClose("temp_optFile2.json", data)
	if err != nil {
		t.Errorf("Could not save opt2 to file: %v", err)
	}

	data, err = file.ReadAndCLose("temp_optFile.json")
	if err != nil {
		t.Errorf("Could not conert file to bytes: %v", err)
	}

	compareopt, err := BytesAsStruct(data, BasicOption{})
	if err != nil {
		t.Errorf("Could not conert bytes to struct: %v", err)
	}

	data, err = file.ReadAndCLose("temp_optFile2.json")
	if err != nil {
		t.Errorf("Could not conert file to bytes: %v", err)
	}
	compareopt2, err := BytesAsStruct(data, RandomOption{})
	if err != nil {
		t.Errorf("Could not conert bytes to struct: %v", err)
	}

	if compareopt != opt {
		t.Errorf("struct does not match %v, %v", compareopt, opt)
	}

	if compareopt2 != opt2 {
		t.Errorf("struct2 does not match %v, %v", compareopt2, opt2)
	}
}
