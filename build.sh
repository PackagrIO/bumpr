echo 'Cleaning workspace'
rm -f ./packagr-bumper-linux-amd64
echo 'Setting up prereqs'
go mod vendor
mkdir -p vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build/
cp /usr/local/linux/lib/pkgconfig/libgit2.pc vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build/libgit2.pc
echo 'Testing'
go test -mod vendor -v -tags "static" $(go list ./... | grep -v /vendor/)
# build linux binary
echo 'Build binary'
. /scripts/toolchains/linux/linux-build-env.sh && go build -mod vendor -ldflags "-X main.goos=linux -X main.goarch=amd64" -o packagr-bumpr-linux-amd64 -tags "static" cmd/bumpr/bumpr.go