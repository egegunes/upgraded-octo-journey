# Percona

## Build

```
$ go build
```

## Adding tasks

```
$ curl \
    -d '{"task1": 135, "task1": 314, "task1": 5431, "task1": 4141, "task5": 47841, "task6": 74641, "task7": 4841, "task8": 4999}' \
    http://localhost:8080
```

Alternatively, you can use `test.sh` to add tasks (it requires `jq`):

```
$ ./test.sh <task count>
```

## Listing tasks

```
$ curl http://localhost:8080/tasks
```
