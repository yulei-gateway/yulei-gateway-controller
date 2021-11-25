package resource

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"

	"google.golang.org/protobuf/types/known/structpb"
)

type UserInfo struct {
	UserName string
	Password string
	Key      string
	Value    int
}

type TestStruct struct {
	RouterFilters []HttpFilter `yaml:"routerFilters,omitempty"`
}

func TestYAMLTestStruct(t *testing.T) {
	var a1 = &WASMFilter{Name: "test"}
	var a2 = &LuaFilter{InlineCode: "test"}
	var filters []HttpFilter
	filters = append(filters, a1)
	filters = append(filters, a2)
	data, err := yaml.Marshal(filters)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))

}

func TestResourceMataDataBuild(t *testing.T) {
	var testMap = map[string]interface{}{}
	testMap["key1"] = 1
	testMap["key2"] = true
	//testMap["Key3"] = []string{"test1", "test2", "test3"}
	//testMap["key4"] = &UserInfo{UserName: "test1", Password: "TET2", Key: "TEST3", Value: 4}

	s, err := structpb.NewStruct(testMap)
	if err == nil {
		fmt.Println(fmt.Sprintf("test %v", s))
	}
}
