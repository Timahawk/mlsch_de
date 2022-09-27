# mlsch.de 
This is the repository for my personal website.
Also the main repository for the game locator.io. 
This can be found under pkg/locator_v2. 
Naming structure is a mess. 

## Locator.io 
Locator.io is a websocket based multiplayer game.
Task is to locate cities, in different regions using only satellite imagery as basemap.
Points are awarded for the distance to the actual location. Highest points wins after 10 Rounds.

### Pros

- typical game flow generally works.
- actually fun (for some people)
- Somewhat configurable

### Cons
- Visuals look bad
- Very unintuitive design
- Not optimized for phones (should be main target)
- Performance 

## Watcher 
Monitors system Memory, CPU load and number of Goroutines. Platform independent. 
Can be found under [go-watcher](https://github.com/Timahawk/go_watcher)

## Chat
First test project for websockets, little chat app with Rooms. Can be found under pgk/chat

