# MESSAGE-BROKER

## Описание

Данный проект представляет из себя брокер очередей, реализованный в виде веб-сервиса.

Сервис обрабатывает 2 метода:

1. ```PUT /queue?v=....``` - положить сообщение в очередь с именем queue (имя очереди может быть любое)

Примеры:

- ```curl -XPUT http://127.0.0.1/color?v=red```
- ```curl -XPUT http://127.0.0.1/name?v=ana```

Ответы:

Status             | Description
------------------ | -----------------------------
200 - OK           | Удалось положить сообщение
400 - Bad request  | Отсутствует параметр ```v```

2. ```GET /queue``` забрать (по принципу FIFO) из очереди с названием queue сообщение и вернуть в теле http запроса, пример (результат, который должен быть при выполненных put’ах выше):

- ```curl http://127.0.0.1/color => red```
- ```curl http://127.0.0.1/color => {пустое тело + статус 404 (not found)}```
- ```curl http://127.0.0.1/name => ana```
- ```curl http://127.0.0.1/name => {пустое тело + статус 404 (not found)}```

В запросе можно задать query-параметр timeout:

```curl http://127.0.0.1/color?timeout=N```

Если в очереди нет готового сообщения получатель ждет либо - до момента прихода сообщения, либо до истечения таймаута (N - кол-во секунд). 
В случае, если сообщение так и не появилось возвращается код 404.

Получатели получают сообщения в том же порядке как от них поступал запрос. Если 2 получателя ждут сообщения (используют таймаут), то первое сообщение получает  тот, кто первый запросил.

Порт, на котором запускается сервис задавается в аргументах командной строки:

```go run cmd/main.go -port=":8000"```

По дефолту сервис запускается на ```:8080```




*В проекте целенаправленно не использовались сторонние пакеты, кроме стандартных библиотек.*
