package messager_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMessager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Messager Suite")
}
