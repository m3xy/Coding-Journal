# Build from latest mariadb image
FROM mariadb:latest

# Initialising container with databases by copying them to the /docker-entrypoint-initdb.d/ directory
# -> Each .sh/.sql/.sql.gz file within this directory is run in alphabetical order
COPY db.sql /docker-entrypoint-initdb.d/
COPY test.sql /docker-entrypoint-initdb.d/

#Environment variable - Root user's password
ENV MARIADB_ROOT_PASSWORD secret

#Environent variables for a user
ENV MARIADB_USER myuser
ENV MARIADB_PASSWORD mypass

#Exposes Port: TCP 3306
EXPOSE 3306