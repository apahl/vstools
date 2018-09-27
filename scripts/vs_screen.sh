#!/bin/bash -l
# The script for screening on a cluster (LSF platform)
# distributed over 200 jobs, the screning of 100k compounds takes ~24h
# usage: bsub <bsub options> ./vs_screen.sh <subdir>

# all ligand and receptor files are expected to be in .pdbqt format
REC="rec-x"                    # receptor name
REC_DIR=/scratch/user/vs/rec   # rec. dir on the cluster data partition, contains also the ref. ligand
REF_LIG="$REC_DIR/rec-x-lig"   # reference ligand for autobox (full path without extension)
LIGANDS=3d_vs

NUM_JOBS=200

SUB_DIR=$1  # ligand set subdir
if [[ $SUB_DIR == "" ]]; then
  echo "missing argument: <subdir>"
  exit 1
fi

LIG_DIR=/scratch/user/vs/lig
OUT_DIR=/scratch/user/vs/out

LFILES=($LIG_DIR/$LIGANDS/$SUB_DIR/lig*.pdbqt)
NUM_LIGS=${#LFILES[@]}
SIZE=$((NUM_LIGS / NUM_JOBS + 1))

mkdir -p $OUT_DIR/$SUB_DIR
FIRST=$(((LSB_JOBINDEX - 1) * SIZE))
LAST=$((LSB_JOBINDEX * SIZE - 1))
if [ $LAST -ge $NUM_LIGS ]; then
  LAST=$((NUM_LIGS - 1))
fi

for ((i=FIRST;i<=LAST;i++)); do
  LF=${LFILES[$i]}
  BN=$(basename $LF .pdbqt)
  smina -r $REC_DIR/$REC.pdbqt -l $LF --autobox_add 8 \
        --autobox_ligand $REF_LIG.pdbqt --num_modes 3 \
        --scoring vinardo --atom_terms $BN.terms \
        -o $OUT_DIR/$SUB_DIR/$BN.pdbqt -q --cpu 1 \
        --log $OUT_DIR/$SUB_DIR/$BN.log
done
