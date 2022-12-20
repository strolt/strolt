# Postgresql (psql, pg_dump, pg_restore)
apk --update add --no-cache postgresql15-client=15.1-r0

rm /usr/bin/clusterdb
rm /usr/bin/createdb
rm /usr/bin/createuser
rm /usr/bin/dropdb
rm /usr/bin/dropuser
rm /usr/bin/pg_basebackup
rm /usr/bin/pg_dumpall
rm /usr/bin/pg_isready
rm /usr/bin/pg_receivewal
rm /usr/bin/pg_recvlogical
rm /usr/bin/reindexdb
rm /usr/bin/vacuumdb
rm /usr/bin/pg_amcheck
rm /usr/bin/pgbench
rm /usr/bin/pg_verifybackup
# rm /usr/bin/pg_versions
rm /usr/bin/PG_VERSION
