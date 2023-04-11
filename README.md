# Promotions

## Overview

A Simple program that receives promotions.csv file and stores objects in mongoDb.
Program use `bufio.NewReader` for processing big files. As CSV file is immutable, the program erases mongoDb storage and stores new file(if exists) every 30 minutes.
Also program use mongoDb replication.
By given ID the endpoint returns the object, otherwise, return not found.

Ex.
curl http://localhost:8080/promotions/172FFC14-D229-4C93-B06B-F48B8C09551

{"id":"172FFC14-D229-4C93-B06B-F48B8C095512", "price":9.68,
"expiration_date": "2022-06-04 06:01:20"}

## Requirements

* Docker Engine  20.10.24
* Docker Compose v2.17.2

## Starting services

Use the command `compose up` to start all services in your local environment.

```bash
docker compose up --detach
```

## Sending CSV File

Use the command `docker cp` to send CSV file to container.

```bash
docker cp promotions.csv verve:/app/promotions/promotions.csv
```
