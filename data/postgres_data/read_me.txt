Data downloaded can be downlaoded here:
https://gadm.org/download_world.html 

sudo docker run --name mlsch-postgis -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=mlsch_data -d -v "/home/ubuntu/mlsch_de:/home/" -p 5555:5432 postgis/postgis


# 1) Data transfer via
https://www.station307.com/#/

# 2) wget
wget link

# 3) go into docker container.
sudo docker exec -it mlsch-postgis bash

# 4) pg_restore 
psql -U postgres -h localhost -p 5555 -d mlsch_data -f backup.sql