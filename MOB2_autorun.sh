#!/bin/bash

echo -e "\033[0;35mWaiting for internet connection...\033[0m"
until ping -c1 google.com &>/dev/null; do sleep 1; done

echo -e "\033[0;32mWaiting for Docker service to be active...\033[0m"
until sudo systemctl is-active docker.service &>/dev/null; do sleep 3; done

echo -e "\033[0;35mStarting Docker compose services...\033[0m"
# Navigate to the project directory
cd /home/rpi4/MOB2

# Docker compose commands
sudo docker compose up --no-start --build &&
sudo docker compose start db mosquitto &&
for service in "db" "mosquitto"; do
    while ! sudo docker compose ps $service | grep "$service" | grep "Up" >/dev/null; do
        sleep 1
    done
done

# Start services
sleep 20
sudo docker compose start http_database_update_server mqtt_database_api

echo -e "\033[0;32mEntering loop to monitor and restart services as necessary...\033[0m"
# Loop to check and restart mqtt_database_api if it stops
while true; do
    if ! sudo docker compose ps mqtt_database_api | grep "mqtt_database_api" | grep "Up" >/dev/null; then
        echo -e "\033[0;35mmqtt_database_api is not running. Restarting...\033[0m"
        sudo docker compose restart mqtt_database_api
    fi
    sleep 30 # Check every 30 seconds
done
