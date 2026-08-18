#!/bin/bash

swag i -g init_router.go -dir app/admin/router --instanceName admin --parseDependency -o docs/admin
swag i -g apis/swagger.go -dir app/report --parseInternal --instanceName report -o docs/report
