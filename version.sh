#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 1 ]; then
	echo "⚠️  Usage: $0 [major|minor|patch]"
	exit 1
fi
TYPE="$1"

echo "🔍 Checking current branch..."
BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [ "$BRANCH" != "main" ]; then
	echo "❌ You must be on the main branch"
	exit 1
fi
echo "✅ On main branch"

echo "📥 Fetching latest tags and code..."
git fetch --tags
git pull

echo "🏷️  Detecting latest tag..."
LAST_TAG="$(git tag -l 'v*' --sort=-v:refname | head -n1)"
if [ -z "$LAST_TAG" ]; then
	LAST_TAG="v0.0.0"
fi
echo "➡️  Latest tag: $LAST_TAG"

VERSION="${LAST_TAG#v}"
MAJOR="$(echo "$VERSION" | cut -d. -f1)"
MINOR="$(echo "$VERSION" | cut -d. -f2)"
PATCH="$(echo "$VERSION" | cut -d. -f3)"

echo "🧮 Bumping version ($TYPE)..."
case "$TYPE" in
major)
	MAJOR=$((MAJOR + 1))
	MINOR=0
	PATCH=0
	;;
minor)
	MINOR=$((MINOR + 1))
	PATCH=0
	;;
patch)
	PATCH=$((PATCH + 1))
	;;
*)
	echo "❌ Invalid type: $TYPE"
	exit 1
	;;
esac

NEW_TAG="v$MAJOR.$MINOR.$PATCH"
echo "✨ New tag: $NEW_TAG"

echo "🏷️  Creating git tag..."
git tag "$NEW_TAG"

echo "🚀 Pushing new tag..."
git push origin "$NEW_TAG"

echo "🎉 Done: $NEW_TAG"
