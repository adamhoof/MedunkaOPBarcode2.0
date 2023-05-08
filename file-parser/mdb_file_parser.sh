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

mdb-export --delimiter=";" --quote='"' --escape-invisible "$1" SKz | awk -F";" 'BEGIN { OFS=";" } { print $25, $24, "\"" $54 "\"", $32, "\"" $34 "\"", "\"" $36 "\"" }' > "$2"

#detect header changes
requiredHeader="Nazev;EAN;ProdejDPH;MJ2;MJ2Koef;StavZ"
normalisedFileHeader=$(head -n 1 "$2" | tr '[:upper:]' '[:lower:]' | tr -d '"' | tr -d " ")
normalisedRequiredHeader=$(echo "$requiredHeader" | tr '[:upper:]' '[:lower:]' | tr -d " ")

if [ "$normalisedFileHeader" != "$normalisedRequiredHeader" ]; then
  echo "CSV file header and required header are not the same" >&2
  exit 1
fi