# Images of the day

A go cli application downloading images of the day from following sources:
- apod
- bing
- earthobservatory
- epod
- nasa

# Usage

To download images simply call:
```
go install github.com/new-er/images-of-the-day@latest
images-of-the-day download
```
This will install the images-of-the-day tool and download all images from all sources to `~/Pictures`.
For more configuration options call:
```
images-of-the-day -h
```
