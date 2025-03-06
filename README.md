## What is it?

#### A system for shop owner and employees enabling faster goods unpacking and information checking. It contains hardware in the form of ["stations"](#) as well as main server to coordinate stations and update database of products.

---

### Technologies
#### Protocols: HTTP, MQTT
#### Devices: ESP32, Raspberry Pi
#### Languages: Go, C, C++, Bash
#### Database: PostgreSQL
#### MQTT broker: Mosquitto
#### Container engine and orchestrator: Docker with docker compose (used for PostgreSQL, Mosquitto, mqtt-database-api, http-database-update-server)

---

### Main project parts

- [mqtt-database-api](#mqtt-database-api)
- [http-database-update-server](#http-database-update-server)
- [cli-control-app](#cli-control-app)
- [product-data-response-displayer-rpi](#rpi-station)
- [product-data-response-displayer-esp32](#esp32-station)

---

#### What is a "station"

- custom-made hardware containing network enabled device, barcode reader and a display, enclosed in a 3D printed box
- displays product information obtained from main server via [mqtt-database-api](#mqtt-database-api) which allows light and efficient communication

---

#### rpi-station

- type of [station](#what-is-a-station) powered by Raspberry Pi

---
  
#### esp32-station

- type of [station](#what-is-a-station) powered by ESP32

---
  
#### mqtt-database-api

- database API layer using the MQTT protocol, so that [stations](#what-is-a-station) can request data
- prevents incorrect queries
- removes the burden of data querying for small, not very powerful devices like [ESP32 station](#product-data-response-displayer-esp32)
  
---

#### http-database-update-server

- HTTP server that performs updates of product database on request
- listens for .mdb or .csv files from [/update](#cli-control-app) and then creates database from those files
  
---

#### cli-control-app

- dockerised CLI app which allows user to update product database via [http-database-update-server](#http-database-update-server) or communicate with stations to send them commands
