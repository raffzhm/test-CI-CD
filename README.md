# pomodoro

![image](https://github.com/bukped/pomodoro/assets/11188109/6488ed71-f5fb-459f-8d52-6e839c1dcf22)

Command line [pomodoro timer](https://en.wikipedia.org/wiki/Pomodoro_Technique), implemented in Go.

## Installation
Go to [Release Page](https://github.com/pomokit/pomodoro/releases)

## Usage
4 pomodoro 1 long break is 1 cycle

## Dev

Release dev
```sh
$env:GOOS = 'linux'
$env:GOOS = 'windows'
$env:GOOS = 'darwin'
$env:CGO_ENABLED = '1'

go mod tidy
git tag                                 #check current version
git tag v0.0.3                          #set tag version
git push origin --tags                  #push tag version to repo
```
