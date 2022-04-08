#!/bin/bash

# This is the Bashscript used to update the Server.
# It is a copy from the one in the Server.
# There it is located directly in ~
# before it can be run make executable with 
# chmod +x ./update_server.sh

echo "LISTEN Ports when Starting"
sudo lsof -i -P -n | grep LISTEN
sudo pkill mlsch_de
echo "LISTEN Ports after pkill"
sudo lsof -i -P -n | grep LISTEN

cd mlsch_de
git pull
go build .
echo "Starting detached."
nohup sudo ./mlsch_de -dev=false >>log.txt &

echo "Sleep 2 Seconds"
sleep 2

echo "LISTEN Ports after  Restart"
sudo lsof -i -P -n | grep LISTEN
echo "Tail of Logfile"
tail ./log.txt
echo "-> Finished the Script."

exit 0