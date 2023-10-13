package main

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExample(t *testing.T) {
	name, surname, patr := "Dmitriy", "Ushakov", "Vasilevich"
	testBody := BodyData{
		Name:       &name,
		Surname:    &surname,
		Patronymic: &patr,
	}
	testChIn, testChOut := make(chan BodyData), make(chan EnrichedData)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		enrichHandler(testChIn, testChOut)
	}()
	testChIn <- testBody
	res := <-testChOut
	assert.Equal(t, 42, res.Age)
	assert.Equal(t, "male", res.Gender)
	assert.Equal(t, "UA", res.Nationality)
}
