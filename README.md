# meetjescraper

[![Build Status](https://travis-ci.com/fiskeben/meetjescraper.svg?branch=master)](https://travis-ci.com/fiskeben/meetjescraper)

A HTTP proxy for the
[Meet je stad](https://meetjestad.net)
project.

## Purpose

This proxy will scrape the table of data for 
a given sensor 
([like this](https://meetjestad.net/data/sensors_recent.php?sensor=242&limit=10))
using the
[scrapejestad](https://github.com/fiskeben/scrapejestad)
scraper and return it as good old machine-friendly JSON.

## API

There is only one endpoint, `/`,
which takes two query parameters:

* `sensor` (number, required) the sensor ID to scrape.
* `limit` (number) maximum number of items to return.
  Defaults to 50, max 100.

Readings are returned as a list of objects, like this:

```json
[
  {
    "SensorID": "242",
    "timestamp": "2019-02-10T11:30:56Z",
    "temperature": 2.75,
    "humidity": 90.75,
    "light": 0,
    "pm25": 0,
    "pm10": 0,
    "Voltage": 3.37,
    "firmware_version": "v2",
    "coordinates": {
      "lat": 60.4308,
      "lng": 5.23242
    },
    "fcnt": 2973,
    "gateways": [
      {
        "name": "mjs-bergen-gateway-2",
        "coordinates": {
          "lat": 0,
          "lng": 0
        },
        "distance": 5.627,
        "rssi": -120,
        "lsnr": 0.25,
        "radio_settings": {
          "frequency": 867.7,
          "sf": "SF9BW125",
          "cr": "4/5CR"
        }
      },
      {
        "name": "mjs-bergen-gateway-3",
        "coordinates": {
          "lat": 0,
          "lng": 0
        },
        "distance": 7.604,
        "rssi": -117,
        "lsnr": -8.75,
        "radio_settings": {
          "frequency": 867.7,
          "sf": "SF9BW125",
          "cr": "4/5CR"
        }
      },
      {
        "name": "mjs-bergen-gateway-5",
        "coordinates": {
          "lat": 0,
          "lng": 0
        },
        "distance": 5.459,
        "rssi": -114,
        "lsnr": -9.75,
        "radio_settings": {
          "frequency": 867.7,
          "sf": "SF9BW125",
          "cr": "4/5CR"
        }
      }
    ]
  }
]
```
