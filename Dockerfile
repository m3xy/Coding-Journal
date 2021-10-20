FROM mariadb:latest

COPY db.sql /docker-entrypoint-initdb.d/

ENV MYSQL_ROOT_PASSWORD secret
ENV MYSQL_DATABASE mydb

EXPOSE 3306