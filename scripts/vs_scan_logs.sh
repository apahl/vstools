#!/bin/bash -l
# usage: bsub <bsub options> ./vs_scan_logs.sh

RES_DIR="vs_clpp_comas"  # result dir
VS_DIR=/scratch/apahl/vs

cd $VS_DIR
mkdir -p $RES_DIR

smina_scan_logs --topha 100 --maxha 35 --minha 22 out $RES_DIR