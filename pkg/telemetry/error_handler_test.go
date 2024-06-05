package telemetry

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
)

func TestErrorHandler(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	testErr1 := errors.New("test error 1")
	testErr2 := errors.New("test error 2")

	handler := NewErrorHandler()
	g.Expect(handler.Error()).ToNot(HaveOccurred())

	handler.Clear()
	g.Expect(handler.Error()).ToNot(HaveOccurred())

	handler.Handle(testErr1)
	g.Expect(handler.Error()).To(Equal(testErr1))

	handler.Handle(testErr2)
	g.Expect(handler.Error()).To(Equal(testErr2))

	handler.Clear()
	g.Expect(handler.Error()).ToNot(HaveOccurred())
}
