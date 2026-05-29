# Нагрузочное тестирование Noterian

Основной сущеностью нашего приложения заметок являются, очевидно, заметки. Поэтому главный путь, который стоит тестировать - `/notes`.

## Метод

Мы проверяем следующие пути:

1. Создание новой заметки
2. Перечисление всех заметок пользователя
3. Получение содержимого какой-то одной заметки

Для всех из них требуется авторизация через JWT и для путя с записью CSRF-токен. Эту проблему, а также проблему получения списка ID заметок для теста чтения, решают вспомогательные скрипты.

Мы создаем 100 аккаунтов и затем в режиме round-robin создаем через них 100,000 заметок.
Потом мы читаем каждую заметки отдельно и наконец получаем список заметок для каждого пользователя.

### Требования

* Запущенный backend, работающий через `nginx.conf` (см. Frontend-репозиторий).
* Чистая база данных с проведенными на ней миграциями (т.е. соответсвующая [init.sql](init.sql))
* Установленная утилита [`vegeta`](https://github.com/tsenart/vegeta)
* Go >= 1.25 для вспомогательных тестов

### Тест на запись

Здесь и далее предполагается, что сервер запущен на том же компьютере, что и проводит тест.

```sh
# Для начала, создаем аккаунты
# Мы забираем их JWT и CSRF, и записываем это все
# в NDJSON файл для JSON с `POST /notes` запросами
go run ./prepare \
    -base=http://localhost:8000/api \
    -users=100 \
    -notes=100000 \
    -out=write_targets.json \
    -users-out=users.json

# Теперь проводим тесты. 1000RPS на 100 секунд.
vegeta attack \
    -targets=write_targets.json \
    -format=json \
    -rate=1000 \
    -duration=100s \
    > write_results.bin

vegeta report < write_results.bin
vegeta report -type=hist -buckets='[0,2ms,5ms,10ms,25ms,50ms,100ms,250ms,500ms,1s,5s]' < write_results.bin
vegeta plot < write_results.bin > write_latency.html
```

### Тест на чтение

```sh
# Для начала для каждого пользователя вызываем `/notes`
# и таким образом получаем ID всех наших заметок для
# теста чтения.
go run ./harvest \
    -base=http://localhost:8000 \
    -users=users.json \
    -reads=100000 \
    -list-out=read_list_targets.json \
    -get-out=read_get_targets.json

# Тест чтения заметки. 2000RPS на 50 секунд.
vegeta attack \
    -targets=read_get_targets.json \
    -format=json \
    -rate=2000 \
    -duration=50s \
    > read_get_results.bin

vegeta report < read_get_results.bin
vegeta report -type=hist -buckets='[0,1ms,5ms,10ms,25ms,50ms,100ms,250ms,500ms,1s]' < read_get_results.bin

# Тест чтения списка заметок
vegeta attack \
    -targets=read_list_targets.json \
    -format=json \
    -rate=500 \
    -duration=60s \
    > read_list_results.bin

vegeta report < read_list_results.bin
```

## Результаты

Результаты теста представлены в папке `results/first`

### Запись

Сразу виден плохой результат теста на запись:
```
Requests      [total, rate, throughput]         100000, 1000.01, 564.54
Duration      [total, attack, wait]             1m40s, 1m40s, 15.469ms
Latencies     [min, mean, 50, 90, 95, 99, max]  847µs, 3.133ms, 2.535ms, 3.684ms, 5.33ms, 18.321ms, 191.897ms
Bytes In      [total, mean]                     21672450, 216.72
Bytes Out     [total, mean]                     6800000, 68.00
Success       [ratio]                           56.46%
Status Codes  [code:count]                      201:56462  502:43538
Error Set:
502 Bad Gateway

Bucket           #      %       Histogram
[0s,     2ms]    28645  28.64%  #####################
[2ms,    5ms]    65758  65.76%  #################################################
[5ms,    10ms]   3566   3.57%   ##
[10ms,   25ms]   1407   1.41%   #
[25ms,   50ms]   251    0.25%
[50ms,   100ms]  218    0.22%
[100ms,  250ms]  155    0.15%
[250ms,  500ms]  0      0.00%
[500ms,  1s]     0      0.00%
[1s,     5s]     0      0.00%
[5s,     +Inf]   0      0.00%
```

Как можно заметить, главная проблема здесь не в сервере, а в nginx.
Перевод API сервера с HTTP/1.0 и добавление следующего текста в nginx.conf:
```
upstream main_backend {
    server main:8000;
    keepalive 64;
    keepalive_requests 10000;
    keepalive_timeout 60s;
}
```
сразу исправил данную проблему, дав нам следующие результаты (`results/second`):

```
Requests      [total, rate, throughput]         100000, 1000.01, 1000.00
Duration      [total, attack, wait]             1m40s, 1m40s, 1.208ms
Latencies     [min, mean, 50, 90, 95, 99, max]  766.375µs, 1.655ms, 1.046ms, 1.272ms, 1.517ms, 10.569ms, 248.062ms
Bytes In      [total, mean]                     26277256, 262.77
Bytes Out     [total, mean]                     6800000, 68.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      201:100000
Error Set:

Bucket           #      %       Histogram
[0s,     2ms]    96990  96.99%  ########################################################################
[2ms,    5ms]    1661   1.66%   #
[5ms,    10ms]   370    0.37%
[10ms,   25ms]   353    0.35%
[25ms,   50ms]   186    0.19%
[50ms,   100ms]  290    0.29%
[100ms,  250ms]  150    0.15%
[250ms,  500ms]  0      0.00%
[500ms,  1s]     0      0.00%
[1s,     5s]     0      0.00%
[5s,     +Inf]   0      0.00%
```

### Чтение

Основная проблема с чтением была в том же 502 и нехватке портов у ноутбука,
на это приходится половина, если не более, из бесчисленных ошибков в логе :-)

```
Requests      [total, rate, throughput]         86510, 1685.05, 399.44
Duration      [total, attack, wait]             1m13s, 51.34s, 22.146s
Latencies     [min, mean, 50, 90, 95, 99, max]  779.25µs, 7.285s, 966.578ms, 30.11s, 30.621s, 32.107s, 34.174s
Bytes In      [total, mean]                     9851953, 113.88
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           33.93%
Status Codes  [code:count]                      0:47462  200:29353  502:9695  
Error Set:
Get "http://localhost:8000/api/notes/9fd4849d-ae2b-49d5-bb05-2a5e304ee9f7": EOF
502 Bad Gateway
Get "http://localhost:8000/api/notes/36af2606-40c6-4c7b-a093-7b651793c4b1": EOF
...
Get "http://localhost:8000/api/notes/9e029eb1-d31c-4d96-a13e-b0dd888261f5": dial tcp 0.0.0.0:0->[::1]:8000: bind: resource temporarily unavailable
...
Get "http://localhost:8000/api/notes/81328713-4379-44aa-bbeb-e2e9054055c2": read tcp [::1]:60475->[::1]:8000: read: connection reset by peer
...
Get "http://localhost:8000/api/notes/8159a060-1357-4e7e-8632-cde2cdec5b2c": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
...

Bucket           #      %       Histogram
[0s,     1ms]    2      0.00%
[1ms,    5ms]    11957  13.82%  ##########
[5ms,    10ms]   2712   3.13%   ##
[10ms,   25ms]   13343  15.42%  ###########
[25ms,   50ms]   4656   5.38%   ####
[50ms,   100ms]  1551   1.79%   #
[100ms,  250ms]  1640   1.90%   #
[250ms,  500ms]  1997   2.31%   #
[500ms,  1s]     5892   6.81%   #####
[1s,     +Inf]   42760  49.43%  #####################################
```

Вместо 2000RPS мы в итоге получили 399.
Однако как можно заметить, даже те запросы, которые выполнились, были медленны.
Интереснее то, что запросы на, "более медленный" endpoint
получения всех заметок пользователя, который еще и возвращает гораздо больше
данных, оказались быстрее:

```
Requests      [total, rate, throughput]         30000, 500.02, 470.51
Duration      [total, attack, wait]             1m0s, 59.998s, 2.222ms
Latencies     [min, mean, 50, 90, 95, 99, max]  1.677ms, 5.626ms, 5.054ms, 6.856ms, 7.838ms, 17.417ms, 152.026ms
Bytes In      [total, mean]                     4205493182, 140183.11
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           94.10%
Status Codes  [code:count]                      200:28231  502:1769
Error Set:
502 Bad Gateway
```

Вероятной причиной этого является следующий кусок `GetNote` в `internal/notes/usecase/notes.go`:
```go
header, err := u.attachmentsClient.GetHeader(ctx, noteID, userID)
if err != nil {
	if status.Code(err) != codes.NotFound {
		return nil, nil, nil, err
	}
}
```

Для всех заметок -- даже тех, у которых нет шапки -- мы делаем запрос в другой
микросервис, что стоит нам целого gRPC запроса. Это можно обойти, если
мы добавим легкий способ узнать, есть ли у заметки шапка.
