package graph

import (
	"app/graph/model"
	"reflect"
	"strconv"
	"strings"
)

func sqlCheckAndWriteArg(arg string, argName string, args *[]interface{}, queryBuilder *strings.Builder) {
	if arg == "" {
		return
	}
	*args = append(*args, arg)
	queryBuilder.WriteString(" AND ")
	queryBuilder.WriteString(argName)
	queryBuilder.WriteString(" = $")
	queryBuilder.WriteString(strconv.Itoa(len(*args)))
}
func sqlWriteArg(arg int, argName string, args *[]interface{}, queryBuilder *strings.Builder) {
	*args = append(*args, arg)
	queryBuilder.WriteString(" ")
	queryBuilder.WriteString(argName)
	queryBuilder.WriteString(" $")
	queryBuilder.WriteString(strconv.Itoa(len(*args)))
}
func generateGetQuery(age string, gender string, nationality string, pageInt int) (string, []interface{}) {
	limit := 10
	offset := (pageInt - 1) * limit

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("SELECT * FROM enriched_data WHERE 1=1")
	args := []interface{}{}

	sqlCheckAndWriteArg(age, "age", &args, &queryBuilder)
	sqlCheckAndWriteArg(gender, "gender", &args, &queryBuilder)
	sqlCheckAndWriteArg(nationality, "nationality", &args, &queryBuilder)
	sqlWriteArg(limit, "LIMIT", &args, &queryBuilder)
	sqlWriteArg(offset, "OFFSET", &args, &queryBuilder)

	return queryBuilder.String(), args
}

func parseUpdateRequestParams(input *model.UpdateEnrichedData) ([]string, []interface{}) {
	argsToCheck := []string{"Name", "Surname", "Patronymic", "Age", "Gender", "Nationality"}
	argsString := []string{}
	args := []interface{}{}
	r := reflect.ValueOf(input)
	for _, key := range argsToCheck {
		value := reflect.Indirect(r).FieldByName(key)
		if value != (reflect.Value{}) {
			queryBuilder := strings.Builder{}
			queryBuilder.WriteString(key)
			queryBuilder.WriteString(" = $")
			if key == "Age" {
				args = append(args, value.Elem().Int())
			} else {
				args = append(args, value.Elem())
			}
			queryBuilder.WriteString(strconv.Itoa(len(args)))
			argsString = append(argsString, queryBuilder.String())
		}
	}
	return argsString, args
}
