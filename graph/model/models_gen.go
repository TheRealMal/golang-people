// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type EnrichedData struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Surname     string  `json:"surname"`
	Patronymic  *string `json:"patronymic,omitempty"`
	Age         int     `json:"age"`
	Gender      string  `json:"gender"`
	Nationality string  `json:"nationality"`
}

type NewEnrichedData struct {
	Name        string  `json:"name"`
	Surname     string  `json:"surname"`
	Patronymic  *string `json:"patronymic,omitempty"`
	Age         int     `json:"age"`
	Gender      string  `json:"gender"`
	Nationality string  `json:"nationality"`
}

type UpdateEnrichedData struct {
	Name        *string `json:"name,omitempty"`
	Surname     *string `json:"surname,omitempty"`
	Patronymic  *string `json:"patronymic,omitempty"`
	Age         *int    `json:"age,omitempty"`
	Gender      *string `json:"gender,omitempty"`
	Nationality *string `json:"nationality,omitempty"`
}