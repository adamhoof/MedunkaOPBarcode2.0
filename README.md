# MedunkaOPBarcode2.0

Upgrade and migration of previous repo that is now private.

---
### What is it for?

#### 1. Speed up goods unpacking for shop owner or employees by providing ["stations"](), where they can quickly check for product information by scanning it's code.

#### 2. Stations also serve as a current product information displayers for customers (when the price changes literary 1x/week, it might come in handy, same with weight changes etc. ...).

---

## What is a "Station"

- station is custom-made hardware, that is connected to MQTT broker
- there can be many (each could be different) stations in one store
- subscribe to topics that allow to control them ([/sleep](#cli-control-app), [/info](#cli-control-app) MQTT endpoints)


- it is able to create request in a specified format and send it to the [mqtt-database-api](#mqtt-database-api) (expecting response containing product
  data)
- print the response on a display
---

## Individual project parts

- cli-control-app
- config
- database
- http-database-update-server
- mqtt-database-api
- product-data-response-displayer-rpi
- product-data-response-displayer-esp32
- product-data

---

### config

*What does it do?*

Config options provider, that is able to parse specific config options from .json file.

###### Usage

- see config/parser.go to see example of the following steps and format:
- create .json file and put in your required configuration options
- create structs inside parser.go as in the example
- load where needed <br>

---

### database

*What does it do?*

Extendable database package, that includes interface of database with required methods and postgresql sample
implementation that I use.

###### Functionality

- standard database operations like connecting, disconnecting, creating and deleting table.
- includes app specific methods to query product data, mentioned later on
- import database from csv file

---

### mqtt-database-api

*What does it do?*

Database API using the MQTT protocol, so that [stations](#what-is-a-station) can request data from database without knowing how to query a
database.

Purpose of this API is to remove the burden of data querying for small, not very powerful devices like [ESP32 station](#product-data-response-displayer-esp32).

It can be used with any device that has all of the [station](#what-is-a-station) requirements, like [Raspberry PI station](#product-data-response-displayer-rpi).

Therefore, increasing the speed of station operation and unifying access to data.

###### Functionality

- listen to specific MQTT topic, so that stations can request product data
- query database based on barcode that is included in the request
- return operation result to the requesting station

---

### http-database-update-server

*What does it do?*

HTTP server that performs updates of product database on request.

Listens for .mdb or .csv files and then creates database from those files.

###### Functionality

- listen to [/update](#cli-control-app) HTTP endpoint, expecting to receive mentioned file
- parse it, create database out of it

---

### cli-control-app

*What does it do?*

CLI app runnable on Windows, Linux and macOS, that allows user to send requests to HTTP server,
control devices connected via MQTT broker.

###### Functionality

- /update - send .mdb or .csv file to http server to update database with current stock information
- /sleep - tell one of the devices to go sleep
- /info - display info about all the devices connected to the mqtt broker

---

### product-data-response-displayer-rpi

*What does it do?*

This is a type of [station](#what-is-a-station).

###### Functionality

- running on raspberry pi (or any other device that is not a microcontroller)

---

### product-data-response-displayer-esp32

*What does it do?*

This is a type of [station](#what-is-a-station).

###### Functionality

- running on esp32 device

---