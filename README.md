# Golang people
## Development section
### SQL Migration
```sh
bash scripts/migration/go-get-psql.sh # Install go migrate for postgres
bash scripts/migration/migrate.sh # Migrate tables from scripts/migration/migrations
```
### Kafka
```sh
bash scripts/kafka/zookeeper.sh # Run zookeeper to run Kafka
bash scripts/kafka/kafka.sh # Run Kafka
bash scripts/kafka/k-producer.sh # Run Kafka producer for topic from .env (FIO)
bash scripts/kafka/k-consumer.sh # Run Kafka consumer for topic from .env (FIO_FAILED)
```
### REST API
```sh
go run app/...
```
### GrapgQL API
```sh
go run github.com/99designs/gqlgen init # Init files
go run github.com/99designs/gqlgen generate # Generate when schema written
go run ./gserver.go # Run graphql server
```

## TODO
Реализовать сервис, который будет получать поток ФИО, из открытых api обогащать ответ наиболее вероятными возрастом, полом и национальностью, и сохранять данные в БД. По запросу выдавать инфу о найденных людях. Необходимо реализовать следующее
- [x] Сервис слушает очередь Kafka FIO, в котором приходит информация с ФИО в формате
```go
{
    "name": "Dmitriy",
    "surname": "Ushakov",
    "patronymic": "Vasilevich" // необязательно
}
```
- [x] В случае некорректного сообщения, обогатить его причиной ошибки (нет обязательного поля, некорректный формат...) и отправить в очередь Kafka `FIO_FAILED`
- [x] Корректное сообщение обогатить
    - Возрастом - https://api.agify.io/?name=Dmitriy
    - Полом - https://api.genderize.io/?name=Dmitriy
    - Национальностью - https://api.nationalize.io/?name=Dmitriy
- [x] Обогащенное сообщение положить в БД postgres (структура БД должна быть создана путем миграций)
- [x] Выставить rest методы
    - [x] Для получения данных с различными фильтрами и пагинацией
    - [x] Для добавления новых людей
    - [x] Для удаления по идентификатору
    - [x] Для изменения сущности
- [x] Выставить GraphQL методы аналогичные п. 5
    - [x] Для получения данных с различными фильтрами и пагинацией
    - [x] Для добавления новых людей
    - [x] Для удаления по идентификатору
    - [x] Для изменения сущности
- [ ] Предусмотреть кэширование данных в Redis
- [ ] Покрыть код логами
- [ ] Покрыть бизнес-логику unit-тестами
- [x] Вынести все конфигурационные данные в .env