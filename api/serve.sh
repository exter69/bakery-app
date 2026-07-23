#!/bin/bash
# Serve OpenAPI docs with Swagger UI (requires Docker)
docker run -p 8081:8080 -e SWAGGER_JSON=/api/openapi.yaml -v $(pwd):/api swaggerapi/swagger-ui
