
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

# Demo
The demo uses the [config-sample](https://github.com/vecosy/config-sample) repository
## Run the server
```shell script
$> docker run --rm  -p 8080:8080 -p 8081:8081 vecosy/vecosy:demo
```

## Call the endpoints
### SmartConfig Strategies
from the [app1/1.0.0](https://github.com/vecosy/config-sample/tree/app1/1.0.0)
* http://localhost:8080/v1/config/app1/1.0.0/dev
* http://localhost:8080/v1/config/app1/1.0.0/int

### Spring-could Strategies
from the [spring-app1/v1.0.0](https://github.com/vecosy/config-sample/tree/spring-app1/1.0.0) 
* http://localhost:8080/v1/spring/v1.0.0/spring-app1/dev
* http://localhost:8080/v1/spring/v1.0.0/spring-app1/int

###Raw file
from the [app1/1.0.0](https://github.com/vecosy/config-sample/tree/app1/1.0.0)
* http://localhost:8080/v1/raw/app1/1.0.0/config.yml
* http://localhost:8080/v1/raw/app1/1.0.0/dev/config.yml

from the [spring-app1/1.0.0](https://github.com/vecosy/config-sample/tree/spring-app1/1.0.0) 
* http://localhost:8080/v1/raw/spring-app1/1.0.0/application.yml
* http://localhost:8080/v1/raw/spring-app1/1.0.0/spring-app1-dev.yml

