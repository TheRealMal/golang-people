# Golang people
Service that listens Kafka topic for new people and adds them to PostgreSQL database. All failed requests being redirected to another Kafka topic. Database can also be accessed through REST/GraphQL API (Get, add, update or delete people). REST API has Redis cache.

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
> kafka producer + consumer w/ docker
> https://habr.com/ru/articles/738874/
### GraphQL Generate
```sh
go run github.com/99designs/gqlgen init # USE ONLY ONCE: Init files
go run github.com/99designs/gqlgen generate # Generate code if schema.graphqls updated
```
### REST + GraphQL API
```sh
go run app/app
```

### TODO
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
- [x] Предусмотреть кэширование данных в Redis (Не добавил для graphql)
- [x] Покрыть код логами
- [ ] Покрыть бизнес-логику unit-тестами
- [x] Вынести все конфигурационные данные в .env