# csgo-demo-stats

> v0.2.0

API is not complete, see roadmap & known issues below.

## Description

`csgo-demo-stats` is a microservice for parsing CSGO demo files and returning relevant statistical data as JSON. The
intended use case is for developers that want to quickly and easily implement CSGO statistics in their applications.

## How to Use

### Set Environment Variables

- `DEMO_STATS_USER` - username for basic auth
- `DEMO_STATS_PASSWORD` - password for basic auth

### Endpoints

|Path|Method|Body|Parameters|
|---|---|---|---|
|`api/parse`|POST|Binary `.dem` file| n/a|
|`api/parse-remote`|GET| n/a|`url` - remote url, `auth` - Full Authorization header (optional)|

### Docker
```bash
sudo docker run \
-p 8080 \ 
-e DEMO_STATS_USER=user \ 
-e DEMO_STATS_PASSWORD=password \
-e PORT=8080 \
martig3/csgo-demo-stats
```

## Example JSON response

```json
{
  "players": [
    {
      "isbot": false,
      "isamember": false,
      "team": "B",
      "steamid": "STEAM_1:1:15055357",
      "steamid64": 76561197990376443,
      "name": "mart1g3",
      "atag": "",
      "rank": 0,
      "kills": 28,
      "mvps": 0,
      "deaths": 14,
      "assists": 2,
      "kd": 2,
      "adr": 103.04,
      "kast": 0,
      "kastRounds": 0,
      "rws": 10.590703528271126,
      "rating": 1.531787871295369,
      "headshots": 15,
      "hsprecent": 53.57142857142857,
      "firstkills": 5,
      "firstdeaths": 4,
      "tradekills": 0,
      "tradedeaths": 0,
      "tradefirstkills": 0,
      "tradefirstdeaths": 0,
      "roundswonv5": 0,
      "roundswonv4": 0,
      "roundswonv3": 1,
      "rounds5k": 0,
      "rounds4k": 0,
      "rounds3k": 2,
      "rounds2k": 7,
      "rounds1k": 8,
      "effFlashes": 17,
      "efpr": 1.0333333333333334,
      "flashDuration": 81894,
      "weapon_stats": {
        "kills": {
          "2": 1,
          "303": 19,
          "304": 6,
          "9": 2
        },
        "headshots": {
          "2": 1,
          "303": 11,
          "304": 1,
          "9": 2
        },
        "accuracy": {
          "2": 14,
          "3": 16,
          "303": 16,
          "304": 17,
          "309": 0,
          "4": 0,
          "405": 0,
          "502": 27,
          "503": 0,
          "504": 0,
          "505": 0,
          "506": 0,
          "9": 18
        },
        "damage": {
          "2": 180,
          "3": 25,
          "303": 2486,
          "304": 654,
          "502": 6,
          "9": 308
        },
        "shots": {
          "2": 27,
          "3": 6,
          "303": 289,
          "304": 102,
          "309": 1,
          "4": 2,
          "405": 36,
          "502": 11,
          "503": 6,
          "504": 27,
          "505": 14,
          "506": 1,
          "9": 27
        },
        "hits": {
          "2": 4,
          "3": 1,
          "303": 50,
          "304": 18,
          "502": 3,
          "9": 5
        }
      },
      "player_damages": {
        "damages": {
          "76561197971293742": 173,
          "76561198012040228": 504,
          "76561198019648240": 720,
          "76561198028320203": 479,
          "76561199020244132": 700
        }
      }
    }
  ]
}
```

## Roadmap

### Fields to Implement

- [ ] KAST aka "kill, assist, survived, traded"
- [ ] HLTV 2 Rating

## Known Issues

### Fields that are not reliably accurate yet

- RWS: `rws`
- HLTV 1.0 rating: `rating`
- Trade Kills: `tradefirstkills`
- Trade First Kills: `tradekills`
- Trade Deaths: `tradefirstdeaths`
- Trade First Deaths: `tradefirstdeaths`

## Libraries Used

- [gin-gonic](https://github.com/gin-gonic/) - web server
- [demoinfocs-golang](https://github.com/markus-wa/demoinfocs-golang) - base library for demo parsing
- [Lots of code taken from this repo](https://github.com/megaclan3000/megaclan3000)
