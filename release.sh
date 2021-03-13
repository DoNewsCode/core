git-chglog --next-tag $1 --output CHANGELOG.md v0.2.0..
git commit CHANGELOG.md -m "chore: generate changelog"
git tag $1
git push && git push --tags
