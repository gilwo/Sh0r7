# Environment variable brief description

## Storage provider
### *SH0R7_STORE_LOCAL*
start with local storage (memory map) - non empty to enable
### *SH0R7_STORE_REDIS*
start with redis storage - redis DSN to enable
### *SH0R7_STORE_FALLBACK*
use local storage provider as a fallback to redis

---
## Metric
### *SH0R7_OTEL_UPTRACE*
use uptrace.dev services for metrics and more - uptrace DSN to enable

---
## Web support
### *SH0R7_WEBAPP*
enable webapp -  non empty to enable
### *SH0R7_ADMIN_KEY*
use specific admin key password - empty will generate random key
### *SH0R7_WEBAPP_TOKEN_EXPIRATION_SHORT_LIVE*
amount of time for a legitimate session of the webapp without refreshing - any valid duration (parsed by go ParseDuration) to override default

---
## Operational 
### *SH0R7_PRODUCTION*
run service in production level use `true` to enable
### *SH0R7_ADDR*
valid `ip:addr` to override default
### *SH0R7_EXPIRATION*
amount of time for default expiration of shorts - any valid duration (parsed by go ParseDuration) to override default

--
## Deployment
### *SH0R7_DEPLOY*
indicate the deployment type:
- `production` : production service
- `developmenty` : development service
- `staging` : staging service
- `testing` : test service
- `localdev` : run as local development service

## Dev 
---
### *SH0R7__DEV_ENV*
enable some dev helper stuff
### *SH0R7_DEV_HOST*
host name of the dev machine
### *SH0R7_METRIC_DB_TABLE_DEV_PREFIX*
suffix to table name in metric db
### *SH0R7_METRIC_DB_GROUP_TYPE_DEV_PREFIX*
suffix to group type name in metric db