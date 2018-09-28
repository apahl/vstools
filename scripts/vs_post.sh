#!/bin/bash -l
# This job writes all docking scores into one file `vs_scores.txt`
# and removes all result files of ligands with scores > -8.0
# In addition, all remaining result files are put into one folder
# and the sub-folders are removed.
# usage: bsub <bsub options> ./vs_post.sh

OUT_DIR=/scratch/apahl/vs/out

cd $OUT_DIR
echo -e "Ligand\tScore" > vs_scores.txt

for d in *; do
  smina_post $OUT_DIR/$d >> vs_scores.txt
  # rm -rf $OUT_DIR/$d  # removal of sub-dirs disabled while testing
done
