# smina

Please also see the excellent blog post [Ligand docking with Smina](https://www.wildcardconsulting.dk/useful-information/ligand-docking-with-smina/) by E.J. Bjerrum

## Prepare Ligand Files

From single mol files with the Compound_Id as name, distributed over 4 subdirs:

    mkdir 1 2 3 4; LSET=3d_vs; for SUBDIR in {1..4}; do CTR=0; cd ${LSET}_${SUBDIR}_mols; for f in *.mol; do bn=$(basename $f .mol); obabel $f -O ../$LSET/$SUBDIR/$bn.pdbqt -h > /dev/null 2>&1; let CTR++; echo -ne "molecules converted: $CTR\r"; done; echo ""; cd ..; done

## Prepare Receptor
### 1. In Pymol

```
remove resn HOH
h_add elem O or elem N
select lig, resn LIG and chain A  # create a selection for the ligand
select rec, prot_xyz and not lig  # select receptor w/o lig

save rec.pdb, rec
save lig.pdb, lig

```

### 2. With OBabel

    obabel rec.pdb -xr -O rec.pdbqt

## Docking on the Cluster

### Script

vs_screen.sh:

```Bash
#!/bin/bash -l
# usage: bsub <bsub options> ./vs_screen.sh <subdir>

# all ligand and receptor files are expected to be in .pdbqt format
REC=clpp
REC_DIR=/scratch/xxx/vs/rec
REF_LIG=$REC_DIR/av145   # reference ligand for autobox (full path without extension)
LIGANDS=3d_vs

NUM_JOBS=200

SUB_DIR=$1  # ligand set subdir
if [[ $SUB_DIR == "" ]]; then
  echo "missing argument: <subdir>"
  exit 1
fi

LIG_DIR=/scratch/xxx/vs/lig
OUT_DIR=/scratch/xxx/vs/out

LFILES=($LIG_DIR/$LIGANDS/$SUB_DIR/lig*.pdbqt)
NUM_LIGS=${#LFILES[@]}
SIZE=$((NUM_LIGS / NUM_JOBS + 1))

mkdir -p $OUT_DIR/$SUB_DIR

# sleep $((LSB_JOBINDEX * 5))
# zero-based indexing!
FIRST=$(((LSB_JOBINDEX - 1) * SIZE))
LAST=$((LSB_JOBINDEX * SIZE - 1))
if [ $LAST -ge $NUM_LIGS ]; then
  LAST=$((NUM_LIGS - 1))
fi

for ((i=FIRST;i<=LAST;i++)); do
  LF=${LFILES[$i]}
  BN=$(basename $LF .pdbqt)
  smina -r $REC_DIR/$REC.pdbqt -l $LF --autobox_add 6 \
        --autobox_ligand $REF_LIG.pdbqt --num_modes 3 \
        --scoring vinardo \
        -o $OUT_DIR/$SUB_DIR/$BN.pdbqt -q --cpu 1 \
        --log $OUT_DIR/$SUB_DIR/$BN.log
done
```

### Starting the Job

    bsub -J "vsscr[1-200]" -w 'done(xxx)' -W "30:00" -R scratch -o /home/users/xxx/vs/jobout/vsscr_%J-%I.txt -e /home/users/xxx/vs/jobout/vsscr_%J-%I.txt ./vs_screen.sh 1  # ligand set subdir

### Watching

    while true; do echo $(date) >> vs.log; echo $(my_jobs) >> vs.log; echo $(ls -1 /scratch/xxx/vs/out*.log | wc -l) >> vs.log; echo "---------------------------" >> vs.log; sleep 900; done

## Scanning Logs on the Cluster

vs_scan_logs.sh:

```Bash
#!/bin/bash -l
# usage: bsub <bsub options> ./vs_scan_logs.sh

RES_DIR="vs"  # result dir

mkdir -p $RES_DIR

for i in out/*; do
  echo $i...
  smina_scan_logs --topha 100 --maxha 35 --minha 22 out/$i $RES_DIR
done
```

    bsub -J "vsscan" -w 'done()' -W "24:00" -R scratch -o /home/users/xxx/vs/jobout/vsscan_%J.txt -e /home/users/xxx/vs/jobout/vsscan_%J.txt ./vs_scan_logs.sh

## Reports

## Example Uses of `smina_scan_logs`

    smina_scan_logs --highest 45 --maxle "-1.10" --sortby le vs vs_le

    smina_scan_logs --topha 15 --maxha 30 --minha 23 vs vs_ha

Get the top 100 per heavy atom from 35 HA down to 22 HAs for each of the three result dirs:

    for i in {1..3}; do echo $i...; smina_scan_logs --topha 100 --maxha 35 --minha 22 out/$i vs; done

## Misc Snippets

Distributing files over multiple folders:

    FILES=*.pdbqt; mkdir 1 2 3; i=0; for f in $FILES; do d=$(printf %d $((1+i/100000))); mv $f $d; let i++; done

Copy files from a text list of file names

    xargs -a files.txt mv -t ../vs/

