#!/bin/bash
set -e

# Perform the the steps necessary before report creation.
# Call this script from within the dir with the results that should be reported.

# 1. Check whether the active site is available in current dir
#    and if it is not, try to copy it from the parent dir
if [ ! -f as.pdb ]; then
  if [ -f ../as.pdb ]; then
    echo "copying active site file from parent dir."
    cp ../as.pdb ./
  else
    echo "active site file 'as.pdb' was not found in current or parent dir!."
    exit 1
  fi
fi

# 2. Combine Ligands and AS and Create PNGs from Ligands
for i in lig-*.pdbqt; do
  bn=$(basename $i .pdbqt)
  obabel as.pdb -l 1 $i -O $bn.pdb -j
  obabel $i -f 1 -l 1 -O $bn.png --gen2d
done

echo ""
echo "All done."