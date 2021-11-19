package resource

import (
	"fmt"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

type UserInfo struct {
	UserName string
	Password string
	Key      string
	Value    int
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
