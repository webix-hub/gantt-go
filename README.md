Backend for Webix Gantt
===========================

### How to start

- create db
- create config.yml with DB access config

```yaml
db:
  host: localhost
  port: 3306
  user: root
  password: 1
  database: projects
```

- start the backend

```shell script
go build
./wg
```
