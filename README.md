# Centralized Configuration System
[![Build Status](https://travis-ci.com/vecosy/vecosy.svg)](https://travis-ci.com/vecosy/vecosy)
[![codecov](https://codecov.io/gh/vecosy/vecosy/branch/develop/graph/badge.svg)](https://codecov.io/gh/vecosy/vecosy)
[![Build Status](https://img.shields.io/badge/docker-pull%20vecosy%2Fvecosy-blue)](https://hub.docker.com/repository/docker/vecosy/vecosy)
[![Go Report Card](https://goreportcard.com/badge/github.com/vecosy/vecosy)](https://goreportcard.com/report/github.com/vecosy/vecosy)
[![Gitter chat](https://badges.gitter.im/gitterHQ/gitter.png)](https://gitter.im/vecosy/community)

![GitHub](https://img.shields.io/github/license/vecosy/vecosy)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fvecosy%2Fvecosy.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fvecosy%2Fvecosy?ref=badge_shield)

![](./docs/schema.png)

## Features
* configurable by a GIT repo
* GRPC
* Spring cloud configuration compatible
* REST
* Auto update (currently only with golang client)

# QuickStart (demo)
The demo uses the [config-sample](https://github.com/vecosy/config-sample) repository
## Run the server
```shell script
$ docker pull vecosy/vecosy:demo
$ docker run --rm  -p 8080:8080 -p 8081:8081 vecosy/vecosy:demo
```

### Generate the JWS token
Use the [app1/1.0.0 branch](https://github.com/vecosy/config-sample/tree/app1/1.0.0) to run
```shell script
$ echo "app1" | jose-util sign --key priv.key --alg RS256
```
*the code below has already a valid token for the vecosy:demo*

## Golang Client 
[Example repo](https://github.com/vecosy/golang-client-example)
```go
package main

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/vecosy/vecosy/v2/pkg/vecosy"
)

func main() {
	jwsToken := "eyJhbGciOiJSUzI1NiJ9.YXBwMQo.A98GFL-P3vtehn0r5GCO_a0OYb5h6trxg3a8WE9hOPDzJ40yOEGtZxyUM6_3Exk65c52-nzWEEc5P-QtgGrgJFOOZlKneKoa1bYBlWRONoysuq95UtSY0doEOMWGvI9AqB685OzmVPuW2UlHg_HlQuuTO6Re1uKc5gr1qZPlyyWEsfoVYTFbfidLoBKWPOuZTxpd8uRx0Rv3LrrmFEcGPHaMNQ2WiXAEJG6OaMTBtwKiynEFH3DU5Rx2WP9M98bH-emC_w7Zq1xKaCOsj2t09F00KohcGC49zSPgPVpp_TwF1qt6_0d0Mnh_Eqi_NHpobVvO85ZOLS05AyW9LQyA5A"
	vecosyCl, err := vecosy.NewClientBuilder("localhost:8081", "app1", "1.0.0", "dev").WithJWSToken(jwsToken).Build(nil)
	panicOnError(err)
	err = vecosyCl.WatchChanges()
	panicOnError(err)
	fmt.Printf("db.user:%s\n", viper.GetString("db.user"))
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
```

## Spring Client
Take a look to [spring-boot-example](https://github.com/vecosy/spring-boot-example)

## Endpoints
Remember to add the `Authorization` header with the token generated before `Bearer [token]` 
(except if the server has been started with `--insecure` option)

### SmartConfig Strategies
from [app1/1.0.0](https://github.com/vecosy/config-sample/tree/app1/1.0.0)
* http://localhost:8080/v1/config/app1/1.0.0/dev
* http://localhost:8080/v1/config/app1/1.0.0/int

### Spring-could Strategies
for [spring-app1/v1.0.0](https://github.com/vecosy/config-sample/tree/spring-app1/1.0.0) 
* http://localhost:8080/v1/spring/v1.0.0/spring-app1/dev
* http://localhost:8080/v1/spring/v1.0.0/spring-app1/int

### Raw file
for [app1/1.0.0](https://github.com/vecosy/config-sample/tree/app1/1.0.0)
* http://localhost:8080/v1/raw/app1/1.0.0/config.yml
* http://localhost:8080/v1/raw/app1/1.0.0/dev/config.yml

for [spring-app1/1.0.0](https://github.com/vecosy/config-sample/tree/spring-app1/1.0.0) 
* http://localhost:8080/v1/raw/spring-app1/1.0.0/application.yml
* http://localhost:8080/v1/raw/spring-app1/1.0.0/spring-app1-dev.yml


# Docker run
## Prepare the configuration
Create a folder for the server configuration `$HOME/myVecosyConf`.

Create a `$HOME/myVecosyConf/vecosy.yml`with your configuration (see Server Configuration chapter)

## Run
```shell script
$> docker -d --name myVecosyInstance -v $HOME/myVecosyConf:/config -p 8080:8080 -p 8081:8081 vecosy/vecosy:latest
```

# Server Configuration
*some configuration options can be passed via command line run `vecosy-server --help` for the options*

## TLS
```yaml
server:
  tls:
    enabled: true
    certificateFile: ./myCert.crt
    keyFile: ./myCert.key
  rest:
    address: ":8443"
  grpc:
    address: ":8081"
...
```
## GIT authentication
### No authentication
```yaml
...
repo:
  remote:
    url: https://github.com/vecosy/config-sample.git
    pullEvery: 30s
  local:
    path: /tmp/vecosyData
```

### plain authentication
```yaml
...
repo:
  remote:
    url: https://github.com/vecosy/config-sample.git
    pullEvery: 30s
    auth:
      type: plain
      username: gitRepoUsername
      password: gitRepoPassword
  local:
    path: /tmp/vecosyData
```

### http (basic) authentication
```yaml
...
repo:
  remote:
    url: https://github.com/vecosy/config-sample.git
    pullEvery: 30s
    auth:
      type: http
      username: gitRepoUsername
      password: gitRepoPassword
  local:
    path: /tmp/vecosyData
```

### public key authentication
```yaml
...
repo:
  remote:
    url: github.com:vecosy/config-sample.git
    pullEvery: 30s
    auth:
      type: pubKey
      username: git
      keyFile: ./myPubKeyFile
      keyFilePassword: myPubKeyPassword
  local:
    path: /tmp/vecosyData
```

## Full Example
```yaml
server:
  tls:
    enabled: true
    certificateFile: ./myCert.crt
    keyFile: ./myCert.key
  rest:
    address: ":8443"
  grpc:
    address: ":8081"
repo:
  remote:
    url: github.com:vecosy/config-sample.git
    pullEvery: 30s
    auth:
      type: pubKey
      username: git
      keyFile: ./myPubKeyFile
      keyFilePassword: myPubKeyPassword
  local:
    path: /tmp/vecosyData
```


# Configuration Repo

## Branching convention
The app configuration is stored in a git repository, vecosy use a branch name convention to manage different configuration on the same repository  `appName/version`
(i.e [app1/1.0.0](https://github.com/vecosy/config-sample/tree/app1/1.0.0)).

## Versions
When a configuration request is processed, the system will find the related branch on the git repo `appname/appVersion` if the specific version is not present, the nearest (`<=`) version will be used. 

## Merging strategies
Vecosy supports two different merging systems, each one use a different naming convention to merge configuration files.

### SmartConfig
![smart config](./docs/smart_config.png)

The `config.yml` in the root folder is the common configuration that will be merged by the specific environment (for dev env: `dev/config.yml`).


#### Example
https://github.com/vecosy/config-sample/tree/app1/1.0.0

### Spring style
![spring config](./docs/spring_config.png)

It uses the [spring-cloud](https://cloud.spring.io/spring-cloud-config/reference/html) naming convention.

The `application.yml` will be overriden by `[appname].yml` file that will be overriden by `[appname]-[profile].yml`

#### Example
https://github.com/vecosy/config-sample/tree/spring-app1/1.0.0

# Security
The security is based on a JWS token.

Every application branch has to contains a `pub.key` file with the public key of the specific application.

## Example 
### 1. Generate the application keys
```shell script
$ openssl genrsa -out priv.key 2048
$ openssl rsa -in priv.key -outform PEM -pubout -out pub.key
```

The `pub.key` has to be added on the application branch on the git repo at the root level.

The `priv.key` will be necessary generating the JWS token for each application that will need the configuration
and should be saved in an external safe place like [vault](https://www.vaultproject.io/)

### 2. generate a jws token
#### install jose-util
```shell script
$ go get -u github.com/square/go-jose/jose-util
$ go install github.com/square/go-jose/jose-util
```
#### generate jws token
```shell script
# the jws payload is not important
$ echo "myAppName" | jose-util sign --key priv.key --alg RS256
```
the generated token can be used as *Bearer* Authorization header, in the `token` variable in the GRPC metadata header or as spring cloud configuration [token](https://github.com/vecosy/spring-boot-example/blob/master/src/main/resources/bootstrap.yml)  

### 3. Configure your application to use the JWS token

#### vecosy-client (golang)
passing on the `vecosy.NewClientBuilder(...).WithJWSToken(jwsToken)` parameter 

#### Spring-cloud application (java)
by Spring cloud configuration [token](https://github.com/vecosy/spring-boot-example/blob/master/src/main/resources/bootstrap.yml)

## Disable the security
the `--insecure` command line option will disable the security system.

# Client (Golang)
Vecosy client use [viper](https://github.com/spf13/viper) as configuration system. 

## Specific viper configuration
```go
    cfg := viper.New()
    vecosyCl,err := vecosy.NewClientBuilder("my-vecosy-server:8080","myApp", "myAppVersion", "integration").
        WithJWStoken(jwsToken).
        Build(cfg)
    // now you can use cfg to get the your app configuration
    cfg.getString("my.app.config")
```

## Default viper configuration 
```go
    vecosyCl,err := vecosy.NewClientBuilder("my-vecosy-server:8080","myApp", "myAppVersion", "integration").
        WithJWStoken(jwsToken).
        Build(nil)
    viper.getString("my.app.config")
```

## Insecure connection 
The server has to be started with `--insecure` option
```go
    vecosyCl,err := vecosy.NewClientBuilder("my-vecosy-server:8080","myApp", "myAppVersion", "integration").
        Insecure().
        Build(nil)
    viper.getString("my.app.config")
```

## TLS connection
```go
    vecosyCl,err:= vecosy.NewClientBuilder("my-vecosy-server:8080","myApp", "myAppVersion", "integration").
        WithTLS("./myTrust.crt").
        WithJWSToken(jwsToken).
        Build(nil)
    viper.getString("my.app.config")
```

## Watch changes
```go
    vecosyCl,err:= vecosy.NewClientBuilder("my-vecosy-server:8080","myApp", "myAppVersion", "integration").
        WithJWStoken(jwsToken).
        Build(nil)
    vecosyCl.WatchChanges()
```
This will maintain a GRPC connection with the server that will inform the client on every configuration changes on the git repo.

It's also possible to add handlers to react to the changes
```go
    vecosyCl.AddOnChangeHandler(func() {
        fmt.Println("something has changed")
    })
```

## More info
have a look to the [integration test](https://github.com/vecosy/vecosy/blob/develop/pkg/vecosy/client_integration_test.go) for more details

# Future features/improvements
* web interface
* metrics
* different config repo type (etcd, redis,...)
* improving spring compatibility (watch changes doesn't work right now)
## Work in progress
Kubernetes helm chart https://github.com/vecosy/helm


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fvecosy%2Fvecosy.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fvecosy%2Fvecosy?ref=badge_large)
