package main

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoodExample(t *testing.T) {
	name, surname, patr := "Dmitriy", "Ushakov", "Vasilevich"
	testBody := BodyData{
		Name:       &name,
		Surname:    &surname,
		Patronymic: &patr,
	}

	testChIn, testChOut, testChErr := make(chan BodyData), make(chan EnrichedData), make(chan []byte)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		enrichHandler(testChIn, testChOut, testChErr)
	}()
	testChIn <- testBody
	res := <-testChOut
	assert.Equal(t, 42, res.Age)
	assert.Equal(t, "male", res.Gender)
	assert.Equal(t, "UA", res.Nationality)
}

func TestBadExample(t *testing.T) {
	name, surname, patr := "0", "0", "0"
	testBody := BodyData{
		Name:       &name,
		Surname:    &surname,
		Patronymic: &patr,
	}

	testChIn, testChOut, testChErr := make(chan BodyData), make(chan EnrichedData), make(chan []byte)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		enrichHandler(testChIn, testChOut, testChErr)
	}()
	testChIn <- testBody
	res := <-testChErr
	var testError ErrorData
	if err := json.Unmarshal(res, &testError); err == nil {
		assert.Equal(t, "Provided FIO does not exist", *testError.ErrorMsg)
		return
	}
	t.Error("Wrong error message format")
}
