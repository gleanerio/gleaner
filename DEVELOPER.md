
# Developer notes

# glcon
glcon integrates gleaner and nabu to make a single env

## integrating with local nabu
[see article](https://levelup.gitconnected.com/import-and-use-local-packages-in-your-go-application-885c35e5624)

`require (
....
github.com/gleanerio/nabu v0.0.0-20211214151422-eda9e525f196
...
}

replace (
github.com/gleanerio/nabu v0.0.0-20211214151422-eda9e525f196 => ../nabu
)`

REMOVE WHEN COMMITTING FOR A PULL REQUEST


## update nabu dependency
if nabu code has been updated, the you need to update the dependency
`go get -u github.com/gleanerio/nabu`

