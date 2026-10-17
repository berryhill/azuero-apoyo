#!/bin/sh
# Sube la versión de la app (VERSION y helm/Chart.yaml): scripts/bump-version.sh [major|minor|patch] (patch si no se dice).
# Se corre antes de cada commit que se publica; el menú de la app muestra la versión y avisa cuando hay una nueva.
set -eu
cd "$(dirname "$0")/.."
part=${1:-patch}
old=$(tr -d ' \n' < VERSION)
IFS=. read -r major minor patch <<EOV
$old
EOV
case $part in
  major) major=$((major + 1)); minor=0; patch=0 ;;
  minor) minor=$((minor + 1)); patch=0 ;;
  patch) patch=$((patch + 1)) ;;
  *) echo "uso: $0 [major|minor|patch]" >&2; exit 2 ;;
esac
new="$major.$minor.$patch"
printf '%s\n' "$new" > VERSION
sed -i.bak -e "s/^version: .*/version: $new/" -e "s/^appVersion: .*/appVersion: \"$new\"/" helm/Chart.yaml && rm -f helm/Chart.yaml.bak
echo "$old → $new"
