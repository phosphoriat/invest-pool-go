# NOTES

## Что решил сам

- Каркас приложения на Fiber v3
- Фронтенд на React, а не на HTML+Go templates
- Деньги в `int64` минорных единицах, произведение через `math/big`
- Фильтрация видимости по ролям на сервере, а не в UI

## Что сгенерировал Claude

- Кодовую базу сервисного слоя, тесты движка, Dockerfile, CI, черновики README, ARCHITECTURE, NOTES

## Где правил руками

- Все кодовую базу и README, ARCHITECTURE, NOTES

## Что отклонил

- Kafka вместо in-process шины, Fabric/Solidity, JWT