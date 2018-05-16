package app

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega" // test package
)

func TestError(t *testing.T) {
	RegisterTestingT(t)

	var ok bool
	e := fmt.Errorf("basic error")

	err := NewError(ErrorBadRequest, "bad request 1", nil)
	_, ok = err.(error)
	Expect(ok).To(BeTrue())
	Expect(err.GetCause()).To(BeNil())
	Expect(err.GetCode()).To(Equal(ErrorBadRequest))

	err2 := NewError(ErrorBadRequest, "bad request 2", e)
	_, ok = err2.(error)
	Expect(ok).To(BeTrue())
	Expect(err2.GetCause()).To(BeIdenticalTo(e))
	Expect(err2.GetCode()).To(Equal(ErrorBadRequest))
	Expect(err2.Error()).To(Equal("bad request 2"))
	Expect(StringifyError(err2)).To(Equal("bad request 2; basic error"))

	err3 := NewError(ErrorBadRequest, "bad request 3", err2)
	_, ok = err3.(error)
	Expect(ok).To(BeTrue())
	Expect(err3.GetCause()).To(BeIdenticalTo(err2))
	Expect(StringifyError(err3)).To(Equal("bad request 3; bad request 2; basic error"))
	causes := ErrorCauses(err3)
	Expect(len(causes)).To(Equal(2))
	Expect(causes[0]).To(Equal(err2.Error()))
	Expect(causes[1]).To(Equal(e.Error()))
}

func TestErrof(t *testing.T) {
	RegisterTestingT(t)

	var ok bool
	e := fmt.Errorf("basic error")

	err := NewErrorf(ErrorBadRequest, nil, "sample %s", "test")
	_, ok = err.(error)
	Expect(ok).To(BeTrue())
	Expect(err.GetCause()).To(BeNil())
	Expect(err.GetCode()).To(Equal(ErrorBadRequest))
	Expect(err.Error()).To(Equal("sample test"))

	err2 := NewErrorf(ErrorBadRequest, e, "sample %s", "test")
	_, ok = err2.(error)
	Expect(ok).To(BeTrue())
	Expect(err2.GetCause()).To(BeIdenticalTo(e))
	Expect(err2.GetCode()).To(Equal(ErrorBadRequest))
	Expect(err2.Error()).To(Equal("sample test"))
}

func TestErrorData(t *testing.T) {
	RegisterTestingT(t)

	var ok bool

	err := NewErrorData(500, "test", nil, KV{"test": "stuff"})
	_, ok = err.(error)
	Expect(ok).To(BeTrue())
	Expect(err.GetCause()).To(BeNil())
	Expect(err.GetCode()).To(Equal(500))
	Expect(err.Error()).To(Equal("test"))
	Expect(err.GetData()).To(Equal(map[string]interface{}{"test": "stuff"}))
}

func TestErrorStack(t *testing.T) {
	RegisterTestingT(t)

	var stack []string

	e01 := fmt.Errorf("default error")
	stack = GetErrorStack(e01)
	Expect(len(stack)).To(Equal(1))
	Expect(stack[0]).To(Equal(`{"code":-1,"message":"default error"}`))

	e02 := NewError(ErrorUnexpected, "sample error", e01)
	stack = GetErrorStack(e02)
	Expect(len(stack)).To(Equal(2))
	Expect(stack[0]).To(MatchJSON(`{"code":5002,"message":"sample error"}`))
	Expect(stack[1]).To(Equal(`{"code":-1,"message":"default error"}`))

	e03 := NewErrorData(ErrorUnexpected, "sample error data", e02, KV{"id": 1, "name": "john smith"})
	stack = GetErrorStack(e03)
	Expect(len(stack)).To(Equal(3))
	Expect(stack[0]).To(MatchJSON(`{"code":5002,"message":"sample error data", "data":{"id":1, "name":"john smith"}}`))
	Expect(stack[1]).To(MatchJSON(`{"code":5002,"message":"sample error"}`))
	Expect(stack[2]).To(Equal(`{"code":-1,"message":"default error"}`))
}
