# Use the official MySQL image as the base image
FROM mysql:8.0.35

# Copy the script to the container
COPY configs/init-db.sh /docker-entrypoint-initdb.d/

## Copy the SQL file to the container
#COPY configs/export.sql /docker-entrypoint-initdb.d/

# Make sure the script is executable
RUN chmod +x /docker-entrypoint-initdb.d/init-db.sh