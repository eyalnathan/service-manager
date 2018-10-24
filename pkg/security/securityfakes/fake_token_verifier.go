// Code generated by counterfeiter. DO NOT EDIT.
package securityfakes

import (
	"context"
	"sync"

	"github.com/Peripli/service-manager/pkg/security"
	"github.com/Peripli/service-manager/pkg/web"
)

type FakeTokenVerifier struct {
	VerifyStub        func(ctx context.Context, token string) (web.TokenData, error)
	verifyMutex       sync.RWMutex
	verifyArgsForCall []struct {
		ctx   context.Context
		token string
	}
	verifyReturns struct {
		result1 web.TokenData
		result2 error
	}
	verifyReturnsOnCall map[int]struct {
		result1 web.TokenData
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeTokenVerifier) Verify(ctx context.Context, token string) (web.TokenData, error) {
	fake.verifyMutex.Lock()
	ret, specificReturn := fake.verifyReturnsOnCall[len(fake.verifyArgsForCall)]
	fake.verifyArgsForCall = append(fake.verifyArgsForCall, struct {
		ctx   context.Context
		token string
	}{ctx, token})
	fake.recordInvocation("Verify", []interface{}{ctx, token})
	fake.verifyMutex.Unlock()
	if fake.VerifyStub != nil {
		return fake.VerifyStub(ctx, token)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.verifyReturns.result1, fake.verifyReturns.result2
}

func (fake *FakeTokenVerifier) VerifyCallCount() int {
	fake.verifyMutex.RLock()
	defer fake.verifyMutex.RUnlock()
	return len(fake.verifyArgsForCall)
}

func (fake *FakeTokenVerifier) VerifyArgsForCall(i int) (context.Context, string) {
	fake.verifyMutex.RLock()
	defer fake.verifyMutex.RUnlock()
	return fake.verifyArgsForCall[i].ctx, fake.verifyArgsForCall[i].token
}

func (fake *FakeTokenVerifier) VerifyReturns(result1 web.TokenData, result2 error) {
	fake.VerifyStub = nil
	fake.verifyReturns = struct {
		result1 web.TokenData
		result2 error
	}{result1, result2}
}

func (fake *FakeTokenVerifier) VerifyReturnsOnCall(i int, result1 web.TokenData, result2 error) {
	fake.VerifyStub = nil
	if fake.verifyReturnsOnCall == nil {
		fake.verifyReturnsOnCall = make(map[int]struct {
			result1 web.TokenData
			result2 error
		})
	}
	fake.verifyReturnsOnCall[i] = struct {
		result1 web.TokenData
		result2 error
	}{result1, result2}
}

func (fake *FakeTokenVerifier) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.verifyMutex.RLock()
	defer fake.verifyMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeTokenVerifier) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ security.TokenVerifier = new(FakeTokenVerifier)