# Wahrwelt Patch

Исправленная reference-версия A/D-сервиса Wahrwelt. SQL queries являются
handwritten source-of-truth, поэтому V3 включает перегенерированный код `sqlc`.

Для RBAC сценариев подготовлены учетные записи `carrier` / `carrier` и
`inspector` / `inspector`.

## Сравнение с уязвимой версией

Patch V1-V10 сохранен в [`wahrwelt.diff`](wahrwelt.diff). В него входят только
production-файлы с исправлениями. README, тесты и generated-код исключены.

## V1 - IDOR редакции манифеста

Дефект позволял авторизованному пользователю прочитать `sealed_route` чужой
редакции по публичному ID. Исправление сверяет владельца манифеста до загрузки
закрытой редакции и возвращает `not found` постороннему.

- `backend/internal/usecase/manifest.go:84-103`

## V2 - SSRF проверки сертификата

Дефект позволял передать внутренний URL или redirect на внутреннюю сеть и
получить `origin_phrase`. Исправление разрешает только HTTP/HTTPS, отклоняет
loopback, private, link-local, unspecified и multicast адреса, проверяет каждый
redirect и подключается только к IP из уже проверенного DNS-ответа.

- `backend/internal/adapter/secondary/registry/client.go:17-200`

## V3 - SQL injection поиска счетов

Дефект вставлял поисковую строку прямо в SQL и позволял вывести `payment_ref`
в публичном поле результата. Исправление переносит запрос в source-of-truth
`invoice.sql`, генерирует типизированный метод `sqlc` и передает шаблон `ILIKE`
отдельным параметром.

- `backend/db/queries/invoice.sql:21-30`
- `backend/internal/adapter/secondary/postgres/invoice.go:36-63`

## V4 - Mass assignment акта досмотра

Дефект переносил legacy-поле `visibility` из комментария в акт и открывал
`inspector_note`. Исправление полностью удаляет изменение видимости из comment
flow. Поле остается в публичном contract для совместимости, но backend его
игнорирует.

- `backend/internal/adapter/primary/http/v1/inspection.go:63-77`
- `backend/internal/usecase/inspection.go:94-113`
- `backend/internal/port/outbound/inspection.go:10-15`
- `backend/internal/adapter/secondary/postgres/inspection.go:54-85`

## V5 - Непривязанный токен лицензии

Дефект подписывал только scope и срок, поэтому токен своей лицензии подходил к
чужому номеру и раскрывал `release_pin`. Исправление подписывает номер лицензии,
owner ID, scope, срок и nonce, а затем сверяет claims с загруженным объектом.

- `backend/internal/usecase/license.go:27-33`
- `backend/internal/usecase/license.go:86-191`

## V6 - Утечка через шаблон договора

Дефект выполнял пользовательский шаблон над полной моделью договора и позволял
вывести `crew_phrase`. Исправление передает шаблону отдельную публичную модель
только с `CrewName`, `Service`, `StartsAt` и `EndsAt`.

- `backend/internal/usecase/contract.go:87-116`

## V7 - Повторное согласование слота

Дефект использовал общий Redis `INCR`, поэтому две подписи одного перевозчика
открывали `slot_code`. Исправление атомарно хранит отдельные carrier и inspector
user ID, не продлевает TTL повторной подписью и сверяет обе записи с назначенными
участниками слота.

- `backend/internal/domain/route_slot.go:32-54`
- `backend/internal/port/outbound/route_slot.go:16-19`
- `backend/internal/usecase/route_slot.go:78-125`
- `backend/internal/adapter/secondary/redis/route_slot_approval.go:14-95`

## V8 - Обход allowlist полей квоты

Дефект проверял `quota_code` до приведения регистра, поэтому другой регистр
обходил запрет. Исправление сначала нормализует имя, затем применяет allowlist,
а публичная domain-модель физически не содержит `quota_code`.

- `backend/internal/domain/quota.go:19-25`
- `backend/internal/usecase/quota.go:68-120`
- `backend/internal/adapter/primary/http/v1/quota.go:35-52`

## V9 - Ghostscript CVE-2024-29510

Дефект использовал Ghostscript 10.03.0, уязвимый к обходу `-dSAFER`, и позволял
читать `archive_code` через выполнение EPS в изолированном renderer. Исправление
закрепляет Ghostscript 10.03.1 и проверяет официальный SHA-512 архива до
распаковки. EPS -> PNG остается рабочим.

- `renderer/Dockerfile:1-23`

## V10 - Доверие X-Forwarded-Role

Дефект считал клиентский `X-Forwarded-Role: gate_operator` доказательством роли
и раскрывал `gate_word`. Исправление сохраняет заголовок только ради contract,
игнорирует его значение и разрешает чтение по identity из сессии владельцу или
назначенному оператору конкретного пропуска.

- `backend/internal/adapter/primary/http/v1/gate_pass.go:47-58`
- `backend/internal/usecase/gate_pass.go:59-83`
