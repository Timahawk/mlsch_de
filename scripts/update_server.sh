#!/bin/bash

cd mlsch_de
git pull
go build .
nohup sudo ./mlsch_de -dev=false >>log.txt &