# Microservice homework

-----

В этом задании необходимо построить микросервис на базе фреймворка grpc

Требуется реализовать:

* Сгенерировать необходимый код из proto-файла
* Базу микросервиса в возможностью остановки сервера
* ACL - контроль доступа от разных клиентов
* Систему логирования вызываемых методов
* Систему сбора сборки статистики ( просто счетчики ) по вызываемым методам

Микросервис будет состоять из 2-х частей:
* Какая-то бизнес-логика. В нашем примере она ничего не делает, её достаточно просто вызывать
* Модуль администрирования, где находится логирование и статистика


Особенности задания:

* В этом задании нельзя использовать глобальные переменные. Всё что необходимо - храните в полях структуры.
* Запускать тесты с go test -v -race

Как сгенерить:
- ставим тулзу для генерации - protoc (это можно нагуглить)
- ставим модуль go для protoc `go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26`
- ставим модуль grpc для protoc `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1`
- генерим (на выходе получаем 2 файла. это ок): `protoc --go_out=. --go-grpc_out=. --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative *.proto`