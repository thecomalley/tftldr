# Go CLI Release Procedure

## 1. Tag the Release
```sh
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

## 2. Build the Binary
```sh
VERSION=$(git describe --tags --abbrev=0)
go build -ldflags "-X main.version=$VERSION" -o tftldr
``` 

## 3. Test the Binary
```sh
./tftldr --version
```

## 4. Package the Binary (Optional)
```
tar -czvf tftldr-v$VERSION.tar.gz tftldr
```

## 5. Create GitHub Release (Optional)
```
gh release create "$VERSION" tftldr-v$VERSION.tar.gz --notes "Changelog for $VERSION"
```

## Development 🧪

To test locally while developing:

```bash
go run .
```