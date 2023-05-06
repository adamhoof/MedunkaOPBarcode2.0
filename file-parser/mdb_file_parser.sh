#!/bin/bash

# $1 is path to .mdb file INCLUDING it's name
# $2 where to save this file
# $3 delimiter

# $25 is name
# $24 is barcode
# $54 is price
# $32 is unit_of_measure
# $34 is unit_of_measure_koef
# $36 is stock

if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <mdb_file_path> <output_csv_path>" >&2
  exit 1
fi

mdb-export --row-delimiter="\n" --delimiter=";" "$1" SKz | awk -F";" 'BEGIN { OFS=";" } { print $25, $24, $54, $32, $34, $36 }' > "$2"

requiredHeader="Nazev;EAN;ProdejDPH;MJ2;MJ2Koef;StavZ"

normalisedFileHeader=$(sed -n '1p' "$2" | tr '[:upper:]' '[:lower:]')
normalisedRequiredHeader=$(echo "$requiredHeader" | tr '[:upper:]' '[:lower:]')

if [ "$normalisedFileHeader" != "$normalisedRequiredHeader" ]; then
  echo "CSV file header and required header are not the same" >&2
fi