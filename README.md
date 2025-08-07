# Compose Updater

A solution for watching your Docker® containers running via Docker Compose for image updates and automatically restarting the compositions whenever an image is refreshed.

## Overview
Compose Updater is an application which continuously monitors your running docker containers. When an image is updated, the updated version gets pulled (or built via --pull) from the registry and the docker compose composition gets restarted (via down and up -d).

Compose Updater is useful for your when you're using image tags which are updated regularly (such as ```image:latest``` or a specific major version like ```image:v3```).

Currently, Compose Updater doesn't help you when your're using image tags that won't change (such as an unchangable SemVer, i.e. ```image:1.2.3```). It won't update your Docker Compose files to use newer image tags.

## Usage
### 1. Prepare your services
You'll need to add two labels to the services you want to watch:

```yaml
services:
  web:
    image: nginx:alpine
    labels:
      - "docker-compose-watcher.watch=1"
      - "docker-compose-watcher.dir=/home/docker/dir"
```

```docker-compose-watcher.watch=1``` exposes the service to Compose Updater.

```docker-compose-watcher.dir``` specifies the path to the directory where this docker-compose.yml lives. If the file is not named docker-compose.yml, you can instead use the label ```docker-compose-watcher.file``` to specify the correct path and file name. This is necessary because it's not possible to find the docker-compose.yml from a running container.

### 2. Run Compose Updater
Copy `docker-compose.override.example.yml` to `docker-compose.override.yml`, and
uncomment the volumes you require inside, depending if you are running Mac,
Linux, or Windows.

Then, run Compose Updater:

```shell
docker compose up -d --build
```

If the registry you're pulling from require authentification, you could mount `~/.docker/config.json` from the host inside the `watcher` service.
Assuming your host user is called `ubuntu`, adding this line to the `volumes` declaration of the `watcher` service should work :
```yaml
volumes:
  # Mount repository configuration (including http(s) settings and credentials) from the host to the container (assuming the host user is called ubuntu)
  - "/home/ubuntu/.docker/config.json:/root/.docker/config.json:ro"
```

**Note:** You'll only need one Compose Updater instance for all your compose services (not one per docker-compose.yml).

## Settings
Configure Compose Updater via environment variables (recommended) or command line arguments:

Env | Param | Default | Meaning
--- | --- | --- | ---
INTERVAL | -interval | 60 | Minutes between checks
CLEANUP | -cleanup | 0 | Run docker system prune -a -f after each run
ONCE | -once | 0 | Run once and exit
PRINT_SETTINGS | -printSettings | 1 | Print settings on start
UPDATE_LOG | -updateLog | '' | Log file for updates and restarts
BUILD | -build | 0 | Build the image of a service with "build:" section in YAML file every run
MQTT_BROKER | -mqttBroker | '' | MQTT Broker address (i.e. tcp://127.0.0.1:1883)
MQTT_CLIENT_ID | -mqttClientId | composeupdater | MQTT Client ID
MQTT_TOPIC_PREFIX | -mqttTopicPrefix | composeupdater | MQTT Topic Prefix
MQTT_USERNAME | -mqttUsername | '' | MQTT Username
MQTT_PASSWORD | -mqttPassword | '' | MQTT Password

## Connecting an MQTT Broker
You can connect Compose Updater to an MQTT Broker (such as Eclispe Mosquitto or HiveMQ). This way, the actions of each run (i.e. image pulls, composition restarts) are published to an MQTT topic. You can use these informations to send push notifications using a solution like [mqttwarn](https://github.com/jpmens/mqttwarn) or [Home Assistant](https://www.home-assistant.io).

To connect to an MQTT broker, specify the required connection parameters in the settings (see above).

Compose Updater published the following topics:

Topic | Corresponding event | Example content
--- | --- | ---
update | On update run start and done | 'start' or 'done'
update/composition/start | On start checking for updates for a specific Docker Composition | YAML File Path
update/composition/restart/dry | On skipping composition restart due to dry-run | YAML File Path
update/composition/restart/skip | On skipping composition restart due to no updated found | YAML File Path
update/composition/restart/start | On restarting a composition | ```{"composeFile": "/path/to/docker-compose.yml", "services":[{"name": "service1", "image": "image:tag"}]}```
update/composition/restart/done | On finished restarting a composition | ```{"composeFile": "/path/to/docker-compose.yml", "services":[{"name": "service1", "image": "image:tag"}]}```
update/composition/service/built | On service's image built | ```{"composeFile": "/path/to/docker-compose.yml", "services":[{"name": "service1", "image": "image:tag"}]}```
update/composition/service/pulled | On service's image pulled | ```{"composeFile": "/path/to/docker-compose.yml", "services":[{"name": "service1", "image": "image:tag"}]}```

### Push notification example
The following [Home Assistant](https://www.home-assistant.io) configuration sends a message via Telegram whenever a Docker composition has been restarted after updating at least one image:

```yaml
automation:
  - alias: "Docker Compose Update"
    trigger:
      - platform: mqtt
        topic: "composeupdater/update/composition/restart/done"
        value_template: "{{ value_json.composeFile }}"
    action:
      - service: notify.telegram_bot
        data:
          message: "Docker images updated: {{ trigger.payload_json.composeFile }}"
```

Read more about how to set up the [MQTT](https://www.home-assistant.io/integrations/mqtt/) and [Telegram](https://www.home-assistant.io/integrations/telegram/) integrations in Home Assistant.

## Troubleshooting

### Use Different Directories

The Compose Updater service should not be placed in the same directory with a target service. Otherwise, both will shut down on an update with 0 (zero) code and no additional logs. Ensure they are in separate directories to avoid unintended shutdowns and outages.

### Relative Paths

Docker Compose resolves relative paths within a Compose file to be absolute paths from the perspective of the Compose CLI ([source](https://github.com/docker/compose/blob/main/pkg/compose/create.go#L1143-L1173)).

Since Compose Updater works by calling the Compose CLI from within the container, it's important that the filepaths match up with the host filesystem, at least for the directories that contain the Compose files you wish to update.

This is why it's important that both sides (e.g. ```/home/docker:/home/docker:ro```) must match, otherwise you may get bind mount errors when an auto update happens.

### Windows Users

If you have a Windows host, your Docker actually runs inside a WSL2 distro,
which accesses your Windows filesystem via the `/mnt/<disk>` format. Therefore,
your target service compose file's volumes will look something like this:

```yaml
volumes:
  - "./data:/data"
  - "./config:/config:ro"
labels:
  - "docker-compose-watcher.watch=1"
  - "docker-compose-watcher.dir=/mnt/c/Users/Foobar/TestServer"
```

While your watcher compose file's volumes will look something like this:

```yaml
volumes:
  - "/var/run/docker.sock:/var/run/docker.sock:ro"
  - "/mnt/c/Users/Foobar:/mnt/c/Users/Foobar:ro"
```

## Governance and Contribution

This repository loosely follows the [Collective Code Construction Contract](./C4.txt), which is somewhat like the [Wikipedia BOLD](https://en.wikipedia.org/wiki/Wikipedia:Be_bold) model.

We aim to quickly merge pull requests that don't break the codebase, even if there are some initial disagreements. A pull request doesn't have to be perfect; future contributions can always make it better.

Frequent contributors who show dedication and expertise will be promoted to maintainers or admins, giving them the power to approve more pull requests. This is because we are just volunteers and might leave at any time, so promoting contributors ensures that governance remains robust and sustainable.

## Contributors

- Heiner
- Mathéo Cimbaro
- wtfprogrammer 
- Lim Ding Wen

## License
GNU General Public License v3.0

Docker® is a trademark of Docker, Inc.

This project is not affiliated with Docker, Inc.
