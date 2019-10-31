# Product Versioned Configuration System
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
./vconf
```
## Call the endpoint
### REST API
```shell script
GET http://localhost:8080/v1/config/myApp/v1.0.0/myApp.yml
```
### Spring cloud config API
example: http://localhost:8080/v1/spring/v1.0.0/myApp/dev
