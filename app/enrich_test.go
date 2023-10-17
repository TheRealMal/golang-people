package main

import (
	"context"
	"encoding/json"
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
	ctx, cancelCtx := context.WithCancel(context.Background())
	go enrichListener(ctx, testChIn, testChOut, testChErr)
	testChIn <- testBody
	res := <-testChOut
	cancelCtx()
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
	ctx, cancelCtx := context.WithCancel(context.Background())
	go enrichListener(ctx, testChIn, testChOut, testChErr)
	testChIn <- testBody
	res := <-testChErr
	cancelCtx()
	var testError ErrorData
	if err := json.Unmarshal(res, &testError); err == nil {
		assert.Equal(t, "provided FIO does not exist", *testError.ErrorMsg)
		return
	}
	t.Error("wrong error message format")
}
