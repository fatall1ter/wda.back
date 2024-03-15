# wda.back - web сервер для wda.front + проверка сессий

[![pipeline status](https://git.countmax.ru/countmax/wda.back/badges/master/pipeline.svg)](https://git.countmax.ru/countmax/wda.back/-/commits/master) [![coverage report](https://git.countmax.ru/countmax/wda.back/badges/master/coverage.svg)](https://git.countmax.ru/countmax/wda.back/-/commits/master)

## Назначение компонента

- Создавался для раздачи статики из [wda.front](https://git.countmax.ru/countmax/wda.front)
- Проксирует запросы к layoutconfig.api по роуту `/v2`
- Проверяет сессию в kratos-e
- Обогощает запросы к layoutconfig.api заголовками
  - X-User-ID (uuid из kratos-a)
  - X-User-EMAIL (login из kratos-a)
  - X-User-Permissions (base64 permissions из keto)

## Техническое решение

Взаимодействует с ory/kratos & ory/keto через соответствующие REST API  
Так же в коде еще остаются куски взаимодействияс бд countmax523 // TODO: выпилить их, почистить код

## Перед первым началом работы с исходным кодом компонента

* Установить go >=1.15
* Установить git
* Клонировать репозиторий `git clone git@git.countmax.ru:countmax/wda.back.git`
* Перейти в папку проекта и установить go зависимости `cd wda.back && go mod download`

## Для построения компонента

### makefile

```bash
cd path/to/wda.back
make build
```

### В ручном режиме

```bash
cd path/to/wda.back
go build
./wda.back -p=8000 # start wda.back bind to 8000 tcp port
```

### docker

```bash
cd path/to/wda.back
make docker
```

## CI

В качестве CI используется [gitlab-ci](https://docs.gitlab.com/ee/ci/)  
Детали отображены в файле .gitlab-ci.yml в корне проекта  
При чери-пике в ветку `release` создается docker образ и загружается на приватный [docker-hub](https://hub.watcom.ru)  
Статика сайта помещается внурь docker образа
### Основной способ сборки для обновления статики сайта

Руками запускается pipeline в gitlab-e и указывается переменная окружения A_SERVER со значением url к внешнему API kratos-a.  
По умолчанию A_SERVER=https://devauth.watcom.ru используется для dev версии приложения, если необходима сборка для другого стека, то неуобходимо указать соответсвующий сервер аутентификации

## Шаги необходимые выполнить для получения результатов построения компонента

* Выполнить билд (см пред пункт)
* Сконфигурировать приложение
* Запустить его

### запуск в docker-е

```bash
$ docker run --name wda-back -p 8080:8000 \
  -d hub.watcom.ru/wda.back
```

Или настроить параметры в файле `docker-compose.yml`

```bash
$ docker-compose up -d
```

Если необходимо запускать вне docker-a, то необходимо собрать соответсвующее приложение и запускать его из директории в октрой будет ледать папка web со статикой сайта.

## Требования к окружению для работы компонента

- Требуется наличие конфигурационного файла `config.yaml` в той же папке, где и сам исполняемый файл или запуск с флагом `-c=/path/to/config.yaml`
- Сетевой доступ к API kratos-a
- Сетевой доступ к API keto
- Сетевой доступ к CONSUL серверу, сервис пытается зарегистрироваться при запуске

## Описание параметров конфигурации компонента

> конфигурация стандартно файл > переменные окружения > флаги

Префикс для переменных окружения `WDA`, тогда если задана переменная окружения `WDA_PROXY_UPSTREAM=http://localhost:8080`, она будет замещать значение из файла конфигурации

```yaml
proxy:
  upstream: http://localhost:9001 # layoutconfig.api url
  health: http://localhost:9002 # см доку к layoutconfig.api, порты апи и healthcheck-ов разнесены
```

Флаги, значения которых используются:

* level - уровень логирования
* logfile - путь в лог файлу
* c - путь к файлу конфигурации
* consul - адрес:TCPport CONSUL сервера
* p - port на котором будет отвечать основное API и статика сайта

Конфигурационный файл содержит комментарии, объясняющие смысл каждого поля

```yaml
app: # метаданные приложения
  name: "wda.back" # наименование приложения
httpd:
  port: "8000" # http порт, который будет пытаться открыть приложения и принимать на него http запросы
  host: "" # ip адрес хоста который будет занимать приложение, можно отсавить пустым
  service:
    port: "8001" # http port для pprof и metrics. Этот порт необходимо указывать для health check-a consul-a
  allow_origins:
    - "*"
proxy:
  upstream: http://localhost:9001 # layoutconfig.api url
  timeout_sec: 30 # timeout запросов к сервису upstream
  health: http://localhost:9002 # см доку к layoutconfig.api, порты апи и healthcheck-ов разнесены
countmax: # неиспользуемый раздел, удалить в будущем
  url: "sqlserver://root:master@study-app.watcom.local:1433?database=CM_Karpov523&connection_timeout=0&encrypt=disable" 
  timeout_sec: 30 # timeout с которым будут работать запросы к БД
session:
  source: kratos # memory | kratos - каким образом инициировать менеджер сессий, в памяти или внешний сервис аутентификации
  url: https://devauth.watcom.ru # url внешнего сервиса аутентификации
  timeout: 10s
permissions:
  source: keto # memory | keto - каким образом инициировать менеджер прав, в памяти или внешний сервис хранения прав
  url: http://elk-02:4466 # url внешнего сервиса хранения прав
  timeout: 10s
env: production # тип окружения в котором запускается сервис, production - логи в json формате, все отсальное обычный logrus формат, котрый лучше выводить в текстовый файл и смотреть VSCode-ом
log:
  level: warn # уровень логирования сервиса: debug, info, warn, error
  file: "" # имя файла лога, если пусто или stdout - будет выводить в stdout, если указано имя фацйла, будет писать в него
layout: # неиспользуемый раздел - удалить в будущем
  proxy: "off" # Параметр, отвечающий за включение проксирования. on - включить, off - выключить
  visible:
    online: "off" # Параметр, отвечающий за видимость в меню вкладки "Онлайн". on - отображается, off - не отображается
    queue: "on" # Параметр, отвечающий за видимость в меню вкладки "Очередь". on - отображается, off - не отображается
    monitoring: "on" # Параметр, отвечающий за видимость в меню вкладки "Мониторинг". on - отображается, off - не отображается
    report: "on" # Параметр, отвечающий за видимость в меню вкладки "Отчет". on - отображается, off - не отображается
consul:
  url: "elk-01:8500" # адрес consul сервера
  serviceid: "wda.back-dev" # уникальный идентификатор сервиса, соответсвует имени контейнера (имена контейнеров во всей системе не должны совпадать)!
  address: "elk-01.watcom.local" # адрес/fqnd имя сервера по которому будет видент данный сервис, host docker машины
  port: 8001 # порт по которому доступны метрики и проверка здоровья сервиса снаружи
tags: "develop,countmax,wda.back,office" # теги сервиса по которым будет осущестялться поиск и разметка в мониторинге, количетсво и порядок строго определенные
  #- develop # №1 окружение: develop, stage, production
  #- countmax # №2 проект откуда сервис: countmax, grib, focus etc...
  #- layoutconfig.api # №3 семейство сервисов: commonapi, dbscanner, incidentmaker, transport.webui etc...
  #- office # №4 локация/датацентр где работает сервис
```

## Особенности публикации и эксплуатации компонента

имеет стандартный набор метрик для Prometheus-a `/metrics`  
при запуске регистрируется в consul-e для service discovering-a  
