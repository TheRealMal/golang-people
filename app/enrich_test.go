package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTotal(t *testing.T) {
	name, surname, patr := "Dmitriy", "Ushakov", "Vasilevich"
	testBody := BodyData{
		Name:       &name,
		Surname:    &surname,
		Patronymic: &patr,
	}
	testChIn, testChOut := make(chan BodyData), make(chan EnrichedData)
	go enrichHandler(testChIn, testChOut)
	testChIn <- testBody
	res := <-testChOut
	assert.Equal(t, 42, res.Age)
	assert.Equal(t, "male", res.Gender)
	assert.Equal(t, "UA", res.Nationality)
}
