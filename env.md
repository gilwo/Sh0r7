# Environment variable brief description

## Storage provider
### *SH0R7_STORE_LOCAL*
start with local storage (memory map) - non empty to enable
### *SH0R7_STORE_REDIS*
start with redis storage - redis url to enable
### *SH0R7_STORE_FALLBACK*
use local storage provider as a fallback to redis

---
## Metric
### *SH0R7_METRIC_DB_PATH*
db path (postgresql only ATM) url - non empty postgresql db path to enable

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
use `true` to enable
### *SH0R7_ADDR*
valid `ip:addr` to override default
### *SH0R7_EXPIRATION*
amount of time for default expiration of shorts - any valid duration (parsed by go ParseDuration) to override default

## Dev 
---
### *__DEV_ENV*
enable some dev helper stuff
### *SH0R7_METRIC_DB_TABLE_DEV_PREFIX*
suffix to table name in metric db
### *SH0R7_METRIC_DB_GROUP_TYPE_DEV_PREFIX*
suffix to group type name in metric db