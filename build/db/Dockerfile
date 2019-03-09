ARG TAG

FROM postgres:${TAG}

COPY ./build/db/init.sql /docker-entrypoint-initdb.d/
