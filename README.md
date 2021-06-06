# Scrape functions fom Database Docs

Quick and dirty script to build a JSON dump of all functions from Posgres, MySQL and SQLite.

## Usage 

```
go run main.go > dump.json

# Get a csv with of all Postgres functions
cat dump.json | jq -r '.Postgres | (.[0] | keys_unsorted) as $keys | $keys, map([.[ $keys[] ]])[] | @csv'
```

## Example output

```json
{
  "Postgres": [
    {
      "Name": "abs(x)",
      "Description": "absolute value",
      "Theme": "Mathematical Functions"
    },
    {
      "Name": "cbrt(dp)",
      "Description": "cube root",
      "Theme": "Mathematical Functions"
    },
    {
      "Name": "ceil(dp or numeric)",
      "Description": "nearest integer greater than or equal to argument",
      "Theme": "Mathematical Functions"
    },
    {
      "Name": "ceiling(dp or numeric)",
      "Description": "nearest integer greater than or equal to argument (same as ceil)",
      "Theme": "Mathematical Functions"
    },
```
