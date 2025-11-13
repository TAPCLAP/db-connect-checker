# MySQL Metrics Exporter

## Описание

Экспортер метрик для мониторинга подключений к MySQL базам данных. Предоставляет метрики для Prometheus.

Экспортер выполняет проверки подключений **периодически в фоновом режиме** (по умолчанию каждые 30 секунд).

## Метрики

Экспортер предоставляет следующие метрики:

### 1. `mysql_connection_available`
- **Тип**: Gauge
- **Описание**: Доступность подключения к MySQL (1 = доступно, 0 = недоступно)
- **Labels**:
  - `host` - хост базы данных
  - `port` - порт базы данных
  - `database` - имя базы данных

### 2. `mysql_connection_duration_seconds`
- **Тип**: Gauge
- **Описание**: Время выполнения проверки подключения в секундах
- **Labels**:
  - `host` - хост базы данных
  - `port` - порт базы данных
  - `database` - имя базы данных

## Использование

### Режим экспортера

Для запуска приложения в режиме экспортера используйте переменную окружения `EXPORTER=true`:

```bash
export EXPORTER=true
export DB_TYPE=mysql
export EXPORTER_PORT=8080  # Опционально, по умолчанию 38080
export CHECK_INTERVAL=30   # Интервал проверки в секундах, по умолчанию 30

# Конфигурация MySQL баз данных
export MYSQL_NAME_0=mydb
export MYSQL_USER_0=root
export MYSQL_PASS_0=password
export MYSQL_HOST_0=localhost
export MYSQL_PORT_0=3306

# Можно добавить несколько баз данных
export MYSQL_NAME_1=anotherdb
export MYSQL_USER_1=admin
export MYSQL_PASS_1=secret
export MYSQL_HOST_1=db.example.com
export MYSQL_PORT_1=3306

./db-connect-checker
```

### Переменные окружения

- `EXPORTER` - включить режим экспортера (true/false)
- `EXPORTER_PORT` - порт для HTTP сервера (по умолчанию 38080)
- `CHECK_INTERVAL` - интервал проверки подключений в секундах (по умолчанию 30)
- `DB_TYPE` - тип базы данных (mysql/mongodb)

### Просмотр метрик

После запуска метрики будут доступны по адресу:
```
http://localhost:8080/metrics
```

### Пример метрик

```
# HELP mysql_connection_available MySQL connection availability (1 = available, 0 = unavailable)
# TYPE mysql_connection_available gauge
mysql_connection_available{database="mydb",host="localhost",port="3306"} 1
mysql_connection_available{database="anotherdb",host="db.example.com",port="3306"} 1

# HELP mysql_connection_duration_seconds MySQL connection check duration in seconds
# TYPE mysql_connection_duration_seconds gauge
mysql_connection_duration_seconds{database="mydb",host="localhost",port="3306"} 0.045
mysql_connection_duration_seconds{database="anotherdb",host="db.example.com",port="3306"} 0.123
```

## Интеграция с Prometheus

Добавьте следующую конфигурацию в `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'mysql-connection-checker'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 30s
```

## Docker

Пример запуска в Docker:

```bash
docker run -d \
  -e EXPORTER=true \
  -e DB_TYPE=mysql \
  -e EXPORTER_PORT=8080 \
  -e CHECK_INTERVAL=30 \
  -e MYSQL_NAME=mydb \
  -e MYSQL_USER=root \
  -e MYSQL_PASS=password \
  -e MYSQL_HOST=mysql-server \
  -e MYSQL_PORT=3306 \
  -p 8080:8080 \
  db-connect-checker
```

