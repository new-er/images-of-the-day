# Images of the day 📸🌍

A small Go CLI that downloads **daily images** from various sources to your local machine.
This way you can enjoy stunning images of the day without visiting the websites every day.

## Supported sources

- **APOD** – Astronomy Picture of the Day (NASA)
- **Bing** – Bing Image of the Day
- **Earth Observatory** – NASA Earth Observatory Image of the Day
- **EPOD** – Earth Science Picture of the Day
- **NASA** – NASA Image of the Day

## Installation

```bash
go install github.com/new-er/images-of-the-day@latest
```

## Usage

Download all images from all sources to `~/Pictures`:
```
images-of-the-day download
```
You can specify a different directory with the `-d` flag:
```
images-of-the-day download -d /path/to/directory
```

## Options
View other available options:
```
images-of-the-day -h
```

## Built with

- [Go](https://golang.org/)
- [Cobra](https://github.com/spf13/cobra)
- [Colly](https://github.com/gocolly/colly)

## Contributing

Contributions are welcome! If you have any ideas for new features, improvements, or bug fixes, please feel free to submit an issue/pull request.
