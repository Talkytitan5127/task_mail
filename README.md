Создать прототип системы, котороя позволит общаться нескольким людям через специальные CLI (command line interface) приложения. Задача требует реализации двух подсистем.

Подсистема #1 — сервер, который позволит создавать "комнаты" (room) для обмена текстовыми сообщениями, длина каждого сообщения не должна превышать 254 байта, кол-во сообщений в room не должно превышать 128 — история “Комнаты” (СУБД не использовать все in-memory).

Подсистема #2 — "клиент", который позволит "публиковать" (publish) сообщения в определенные "комнаты" и, который позволит подписываться (subscribe) на определенные комнаты. ”Комнаты” полностью изолированны друг от друга, т.е. клиенты могут взаимодействовать только с “комнатами”, на которые они были подписаны ранее — список таких комнат задается в файле конфигурации или через CLI.

Подсистемы имеют ряд ограничений и свойств:

- Клиент должен получать историю каждой комнаты при успешном подключение.
- Каждый клиент должен иметь имя, имена должны быть уникальными в рамках “комнаты”. Клиент сообщает свое имя клиент при “подписание” на определеную “комнату”.
- Сервер и клиент — это консольные приложения, которые могут иметь либо файл конфигурации либо все опции передаются через CLI.

Ожидаемый результат:
1) Исходные код сервера и клиента реализованные на Go и выложенный на github.com.
2) Минимальный набор тестов.


------------------------------------------------

Собрать проект:
```console
chmod 700 build.sh
./build.sh
```

Запустить сервер:
```console
cd Server/
$GOPATH/bin/Server
```
![Image alt](https://github.com/Talkytitan5127/task_mail/raw/picture/desc/json_runserver.png)

Запустить клиент:
```console
cd Client/
$GOPATH/bin/Client
```
![Image alt](https://github.com/Talkytitan5127/task_mail/raw/picture/desc/json_runclient.png)

Доступные команды пользователю:
```console
subscribe room nickname
publish room message
get_history room
```

Выполнение команды "publish":
![Image alt](https://github.com/Talkytitan5127/task_mail/raw/picture/desc/json_publish.png)

Выполнение команды "subscribe":
![Image alt](https://github.com/Talkytitan5127/task_mail/raw/picture/desc/json_subscribe.png)

Выполнение команды "get_history":
![Image alt](https://github.com/Talkytitan5127/task_mail/raw/picture/desc/json_history.png)

Запустить тесты (из корневой папки проекта):
```console
go test -v
```
![Image alt](https://github.com/Talkytitan5127/task_mail/raw/picture/desc/json_test.png)

## Note:
В файле конфигурации для сервера в поле "room_name" указываются названия комнат, которые будут созданы после запуска сервера.
В файле конфигурации для клиента в поле "Rooms" указываются комнаты в формате {"название комнаты":"имя клиента"}, на которые подписан Клиент.