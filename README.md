
# Product Versioned Configuration System
[![Build Status](https://travis-ci.com/vecosy/vecosy.svg)](https://travis-ci.com/vecosy/vecosy)
[![codecov](https://codecov.io/gh/vecosy/vecosy/branch/develop/graph/badge.svg)](https://codecov.io/gh/vecosy/vecosy)
[![Build Status](https://img.shields.io/badge/docker-pull%20vecosy%2Fvecosy%3Adev-blue)](https://hub.docker.com/repository/docker/vecosy/vecosy)
![GitHub](https://img.shields.io/github/license/vecosy/vecosy)

**vecosy** is a centralized system based on the application version.


## Features
* GIT configuration repo
* GRPC
* Spring cloud configuration compatible
* REST



# Quick start
## create a configuration repo
The configuration repository is a GIT repository that follow the branch naming convention
``[appName]/[appVersion]``
example
``myApp/v1.0.0``

### Init the repo
```shell script
$> mkdir configData
$> cd configData
$> git init .
```
### create the first app [myApp]
```shell script
$> cd configData
$> git checkout --orphan myApp/v1.0.0
```
### add configuration file and commit
```shell script
$> vim myApp.yml
$> vim myApp-profile.yml # used on spring cloud
$> git commit -a -m "myApp configuration"
```

## Run the server
### Configure the server
Edit the `config/vconf.yml` and set `repo.url` to the configData folder created before

example:
```
repo:
  type: git
  url: ~/configData
  path: /tmp/vconfData
```
### Run
```shell script
./vecosy
```
## Call the endpoint
### REST API
```shell script
GET http://localhost:8080/v1/config/myApp/v1.0.0/myApp.yml
```
### Spring cloud config API
example: http://localhost:8080/v1/spring/v1.0.0/myApp/dev
