package generator

import "testing"

func TestGenStruct(t *testing.T) {
	GenStruct("project", "Project")
}

func TestGenMessage(t *testing.T) {
	GenProtoMessage("project", "ProjectMessage")
}
