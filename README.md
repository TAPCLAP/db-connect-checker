# DB Connect Checker

Утилита для проверки подключений к базам данных MySQL и MongoDB с поддержкой экспорта метрик для Prometheus.

## Возможности

- ✅ Проверка подключений к MySQL и MongoDB базам данных
- ✅ Поддержка множественных MySQL подключений одновременно
- ✅ Режим экспортера метрик для Prometheus
- ✅ Поддержка TLS/SSL подключений для MySQL
- ✅ Автоматические повторные попытки при неудачном подключении
- ✅ Периодические проверки в фоновом режиме
- ✅ Docker образ на базе scratch для минимального размера

## Режимы работы

### 1. Режим проверки подключения (по умолчанию)

В этом режиме приложение пытается подключиться к указанным базам данных и завершает работу с соответствующим кодом выхода.

**Коды выхода:**
- `0` - все подключения успешны
- `1` - ошибка конфигурации или подключения
- `2` - все попытки подключения исчерпаны

### 2. Режим экспортера метрик

В этом режиме приложение запускает HTTP-сервер с эндпоинтом `/metrics` для Prometheus. Проверки выполняются периодически в фоновом режиме.

## Установка

### Сборка из исходников

```bash
git clone https://github.com/OrangeAppsRu/db-connect-checker.git
cd db-connect-checker
go build -o db-connect-checker .
```

### Docker

```bash
docker build -f docker/checker/Dockerfile -t db-connect-checker .
```

## Использование

### Режим проверки подключения

#### MySQL

```bash
export DB_TYPE=mysql
export TRIES=10

# Конфигурация первой БД
export MYSQL_NAME_0=mydb
export MYSQL_USER_0=root
export MYSQL_PASS_0=password
export MYSQL_HOST_0=localhost
export MYSQL_PORT_0=3306

# Конфигурация второй БД (опционально)
export MYSQL_NAME_1=anotherdb
export MYSQL_USER_1=admin
export MYSQL_PASS_1=secret
export MYSQL_HOST_1=db.example.com
export MYSQL_PORT_1=3306
export MYSQL_TLS_1=true  # Включить TLS

./db-connect-checker
```

#### MongoDB

```bash
export DB_TYPE=mongodb
export MONGODB_URI="mongodb://user:password@localhost:27017/mydb"
export TRIES=10

./db-connect-checker
```

### Режим экспортера метрик

```bash
export EXPORTER=true
export DB_TYPE=mysql
export EXPORTER_PORT=38080
export CHECK_INTERVAL=30

# MySQL конфигурация
export MYSQL_NAME_0=mydb
export MYSQL_USER_0=root
export MYSQL_PASS_0=password
export MYSQL_HOST_0=localhost
export MYSQL_PORT_0=3306

./db-connect-checker
```

Метрики будут доступны по адресу: `http://localhost:38080/metrics`

## Переменные окружения

### Общие

| Переменная | Описание | Значение по умолчанию |
|-----------|----------|----------------------|
| `DB_TYPE` | Тип базы данных (`mysql` или `mongodb`) | `mysql` |
| `EXPORTER` | Включить режим экспортера (`true`/`false`) | `false` |
| `TRIES` | Количество попыток подключения | `10` |

### Режим экспортера

| Переменная | Описание | Значение по умолчанию |
|-----------|----------|----------------------|
| `EXPORTER_PORT` | Порт для HTTP сервера | `38080` |
| `CHECK_INTERVAL` | Интервал проверки в секундах | `30` |

### MySQL конфигурация

Для каждой базы данных используйте индекс `N` (начиная с 0):

| Переменная | Описание | Обязательная |
|-----------|----------|-------------|
| `MYSQL_NAME_N` | Имя базы данных | Да |
| `MYSQL_USER_N` | Имя пользователя | Да |
| `MYSQL_PASS_N` | Пароль | Да |
| `MYSQL_HOST_N` | Хост | Да |
| `MYSQL_PORT_N` | Порт | Нет (по умолчанию `3306`) |
| `MYSQL_TLS_N` | Использовать TLS (`true`/`false`) | Нет |

**Пример для нескольких баз:**
```bash
export MYSQL_NAME_0=db1
export MYSQL_USER_0=user1
export MYSQL_PASS_0=pass1
export MYSQL_HOST_0=host1

export MYSQL_NAME_1=db2
export MYSQL_USER_1=user2
export MYSQL_PASS_1=pass2
export MYSQL_HOST_1=host2
export MYSQL_TLS_1=true
```

### MongoDB конфигурация

| Переменная | Описание | Обязательная |
|-----------|----------|-------------|
| `MONGODB_URI` | URI подключения к MongoDB | Да (если `DB_TYPE=mongodb`) |

## Метрики Prometheus

Подробное описание метрик доступно в [METRICS_USAGE.md](METRICS_USAGE.md).

### Доступные метрики

**`mysql_connection_available`** (Gauge)
- Доступность подключения (1 = доступно, 0 = недоступно)
- Labels: `host`, `port`, `database`

**`mysql_connection_duration_seconds`** (Gauge)
- Время выполнения проверки в секундах
- Labels: `host`, `port`, `database`

### Пример вывода метрик

```prometheus
# HELP mysql_connection_available MySQL connection availability (1 = available, 0 = unavailable)
# TYPE mysql_connection_available gauge
mysql_connection_available{database="mydb",host="localhost",port="3306"} 1
mysql_connection_available{database="anotherdb",host="db.example.com",port="3306"} 1

# HELP mysql_connection_duration_seconds MySQL connection check duration in seconds
# TYPE mysql_connection_duration_seconds gauge
mysql_connection_duration_seconds{database="mydb",host="localhost",port="3306"} 0.045
mysql_connection_duration_seconds{database="anotherdb",host="db.example.com",port="3306"} 0.123
```

### Интеграция с Prometheus

Добавьте в `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'mysql-connection-checker'
    static_configs:
      - targets: ['localhost:38080']
    scrape_interval: 30s
```

## Docker

### Сборка

```bash
docker build -f docker/checker/Dockerfile -t db-connect-checker .
```

### Запуск в режиме проверки

```bash
docker run --rm \
  -e DB_TYPE=mysql \
  -e MYSQL_NAME_0=mydb \
  -e MYSQL_USER_0=root \
  -e MYSQL_PASS_0=password \
  -e MYSQL_HOST_0=mysql-server \
  -e MYSQL_PORT_0=3306 \
  db-connect-checker
```

### Запуск в режиме экспортера

```bash
docker run -d \
  --name db-checker \
  -e EXPORTER=true \
  -e DB_TYPE=mysql \
  -e EXPORTER_PORT=38080 \
  -e CHECK_INTERVAL=30 \
  -e MYSQL_NAME_0=mydb \
  -e MYSQL_USER_0=root \
  -e MYSQL_PASS_0=password \
  -e MYSQL_HOST_0=mysql-server \
  -e MYSQL_PORT_0=3306 \
  -p 38080:38080 \
  db-connect-checker
```

Проверка метрик:
```bash
curl http://localhost:38080/metrics
```

## Примеры использования

### Kubernetes InitContainer

Используйте для ожидания готовности базы данных перед запуском основного контейнера:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  initContainers:
  - name: wait-for-db
    image: db-connect-checker:latest
    env:
    - name: DB_TYPE
      value: "mysql"
    - name: TRIES
      value: "30"
    - name: MYSQL_NAME_0
      value: "mydb"
    - name: MYSQL_USER_0
      valueFrom:
        secretKeyRef:
          name: db-credentials
          key: username
    - name: MYSQL_PASS_0
      valueFrom:
        secretKeyRef:
          name: db-credentials
          key: password
    - name: MYSQL_HOST_0
      value: "mysql-service"
    - name: MYSQL_PORT_0
      value: "3306"
  containers:
  - name: myapp
    image: myapp:latest
```

### Kubernetes Deployment с экспортером

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: db-connection-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: db-connection-exporter
  template:
    metadata:
      labels:
        app: db-connection-exporter
    spec:
      containers:
      - name: exporter
        image: db-connect-checker:latest
        env:
        - name: EXPORTER
          value: "true"
        - name: DB_TYPE
          value: "mysql"
        - name: EXPORTER_PORT
          value: "38080"
        - name: CHECK_INTERVAL
          value: "30"
        - name: MYSQL_NAME_0
          value: "mydb"
        - name: MYSQL_USER_0
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: username
        - name: MYSQL_PASS_0
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: password
        - name: MYSQL_HOST_0
          value: "mysql-service"
        ports:
        - containerPort: 38080
          name: metrics
---
apiVersion: v1
kind: Service
metadata:
  name: db-connection-exporter
  labels:
    app: db-connection-exporter
spec:
  ports:
  - port: 38080
    targetPort: metrics
    name: metrics
  selector:
    app: db-connection-exporter
```

### CI/CD Pipeline

Проверка доступности БД перед деплоем:

```bash
#!/bin/bash
# check-db-before-deploy.sh

docker run --rm \
  -e DB_TYPE=mysql \
  -e TRIES=5 \
  -e MYSQL_NAME_0=production_db \
  -e MYSQL_USER_0="${DB_USER}" \
  -e MYSQL_PASS_0="${DB_PASS}" \
  -e MYSQL_HOST_0="${DB_HOST}" \
  -e MYSQL_PORT_0=3306 \
  db-connect-checker

if [ $? -eq 0 ]; then
  echo "Database is ready, proceeding with deployment"
  # deploy commands here
else
  echo "Database is not ready, aborting deployment"
  exit 1
fi
```

## Разработка

### Структура проекта

```
.
├── main.go                 # Точка входа приложения
├── go.mod                  # Go модуль
├── go.sum                  # Go зависимости
├── envs.sh                 # Пример переменных окружения
├── METRICS_USAGE.md        # Документация по метрикам
├── README.md               # Этот файл
├── docker/
│   └── checker/
│       └── Dockerfile      # Dockerfile для сборки
└── pkg/
    ├── metrics/
    │   ├── mysql.go        # Экспортер метрик
    │   └── example_test.go # Тесты
    ├── mysqlcheck/
    │   ├── main.go         # Логика проверки MySQL
    │   └── main_test.go    # Тесты
    ├── types/
    │   └── main.go         # Типы данных
    └── util/
        ├── main.go         # Утилиты
        └── main_test.go    # Тесты
```

### Запуск тестов

```bash
go test ./...
```

### Сборка

```bash
go build -o db-connect-checker .
```

### Линтинг

```bash
go vet ./...
golangci-lint run
```

## Лицензия

[MIT License](LICENSE)

## Автор

OrangeApps

## Поддержка

Если у вас возникли проблемы или есть предложения, создайте [Issue](https://github.com/OrangeAppsRu/db-connect-checker/issues).
