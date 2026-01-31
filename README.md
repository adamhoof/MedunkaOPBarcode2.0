## What is it?

#### A system for shop owner and employees enabling faster goods unpacking and information checking. It contains hardware in the form of ["stations"](#) as well as main server to coordinate stations and update database of products.

---

## Demo video of a station

https://github.com/user-attachments/assets/53558eb1-a326-4118-b87e-26fe000d4632

### Technologies
#### Protocols: HTTP, MQTT
#### Devices: Raspberry Pi
#### Languages: Go, C, Bash
#### Database: PostgreSQL
#### MQTT broker: Mosquitto
#### Container engine and orchestrator: Docker with docker compose (used for PostgreSQL, Mosquitto, mqtt-database-api, http-database-update-server, cli-control-app)

---

### Main project parts

- [mqtt-database-api](#mqtt-database-api)
- [http-database-update-server](#http-database-update-server)
- [cli-control-app](#cli-control-app)

---

#### What is a "station"

- custom-made hardware containing network enabled device, barcode reader and a display, enclosed in a 3D printed box
- displays product information obtained via [mqtt-database-api](#mqtt-database-api) which allows light and efficient communication

---
  
#### mqtt-database-api

- database API layer using the MQTT protocol, [stations](#what-is-a-station) request data from it
- prevents incorrect queries
- removes the burden of data querying for small, not very powerful devices like stations
  
---

#### http-database-update-server

- HTTP server that performs updates of product database on request
- listens for .mdb /update endpoint, parses it
  
---

#### cli-control-app

- dockerised CLI app which allows user to update product database via [http-database-update-server](#http-database-update-server) or communicate with stations to send them commands
